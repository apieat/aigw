package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"gitee.com/bjf-fhe/apinx/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ApiConfig struct {
	Def     string `yaml:"def"`
	BaseUrl string `yaml:"base_url"`
	def     *openapi3.T
	server  *openapi3.Server
	//you should keep the security inputs' safety by yourself
	securityArgs map[string]interface{} `yaml:"security_args"`
}

func (a *ApiConfig) Init() (err error) {
	a.def, err = openapi3.NewLoader().LoadFromFile(a.Def)
	if err == nil {
		if no, err := strconv.Atoi(a.BaseUrl); err == nil {
			a.server = a.def.Servers[no]
		} else {
			a.server = &openapi3.Server{
				URL: a.BaseUrl,
			}
		}
	}
	return err
}

func (a *ApiConfig) GetFunctions() []openai.FunctionDefinition {
	var ret []openai.FunctionDefinition
	for pName, v := range a.def.Paths {
		for mName, v2 := range v.Operations() {
			ret = append(ret, openai.FunctionDefinition{
				Name:        getFunctionName(pName, mName),
				Description: v2.Summary + "\n" + v2.Description,
				Parameters:  getParameterJson(v2),
			})
		}
	}
	return ret
}

func (a *ApiConfig) GetFunctionByName(pName, mName string) []openai.FunctionDefinition {
	var ret []openai.FunctionDefinition
	_, op, err := openapi.GetPathAndMethod(a.def, pName, mName)
	if err == nil {
		ret = append(ret, openai.FunctionDefinition{
			Name:        getFunctionName(pName, mName),
			Description: op.Summary + "\n" + op.Description,
			Parameters:  getParameterJson(op),
		})
	}
	return ret
}

func (a *ApiConfig) Call(id string, f *openai.FunctionCall) (string, string, json.RawMessage, error) {
	pName, mName := getPathFromFunctionName(f.Name)
	_, op, err := openapi.GetPathAndMethod(a.def, pName, mName)
	var callingArg openapi.CallConfig
	if err == nil {
		err = json.Unmarshal([]byte(f.Arguments), &callingArg.Input)
		if err == nil {
			var sess *openapi.CallingSession
			sess, err = openapi.NewCalling().
				SetPathAndMethod(pName, mName, op).
				SetServer(a.server).
				SetSecurityByDef(a.securityArgs, a.def).
				Do(&openapi.CallConfig{})
			if err == nil {
				return pName, mName, sess.GetResult().Response, nil
			}
		}
	}
	return pName, mName, nil, err
}

func getFunctionName(pName, mName string) string {
	pName = strings.ReplaceAll(pName, "/", "_")
	return mName + "_" + pName
}

func getPathFromFunctionName(fName string) (pName, mName string) {
	spl := strings.Split(fName, "_")
	pName = strings.ReplaceAll(spl[1], "_", "/")
	mName = spl[0]
	return
}

func getParameterJson(p *openapi3.Operation) *openapi3.Schema {
	//change body to parameters
	var content = p.RequestBody.Value.Content
	var typ string
	var parameters openapi3.Schema
	parameters.Type = "object"
	parameters.Properties = make(map[string]*openapi3.SchemaRef)
	if p.Parameters != nil {
		for _, v := range p.Parameters {
			parameters.Properties[v.Value.Name] = v.Value.Schema
		}
	}
	if len(content) > 1 {
		if _, ok := content["application/json"]; ok {
			typ = "application/json"
		} else {
			for k := range content {
				typ = k
				break
			}
		}
		logrus.WithField("selected", typ).Warn("there are more than one content type in request body, only one will be used")
	} else if len(content) == 1 {
		for k := range content {
			typ = k
			break
		}
	}
	if typ != "" {
		parameters.Properties["body"] = content[typ].Schema
	}

	return &parameters
}
