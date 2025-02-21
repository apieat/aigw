package config

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"gitee.com/bjf-fhe/apinx/openapi"
	"github.com/apieat/aigw/model"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ApiConfig struct {
	Def     string `goblet:"def"`
	BaseUrl string `goblet:"base_url"`
	def     *openapi3.T
	server  *openapi3.Server
	//you should keep the security inputs' safety by yourself
	SecurityArgs map[string]interface{} `goblet:"security_args"`
}

func (a *ApiConfig) Init() (err error) {

	if a.Def == "" {
		return fmt.Errorf("def file is not given, please check your config")
	}

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
	logrus.WithField("server", a.server).WithField("base_url", a.BaseUrl).Info("api config loaded")
	return err
}

func (a *ApiConfig) GetFunctions(allowed []model.AllowedFunction) []openai.FunctionDefinition {
	var allowdMap map[string]map[string]bool
	if (len(allowed)) > 0 {
		allowdMap = make(map[string]map[string]bool)
		for _, v := range allowed {
			allowdMap[v.Path] = map[string]bool{v.Method: true}
		}
	}

	var ret []openai.FunctionDefinition
	for pName, v := range a.def.Paths.Map() {
		for mName, v2 := range v.Operations() {
			if allowdMap != nil {
				if !allowdMap[pName][mName] {
					continue
				}
			}
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

func (a *ApiConfig) Call(id string, functions []openai.FunctionDefinition, f *openai.FunctionCall) (string, string, json.RawMessage, error) {
	pName, mName := getPathFromFunctionName(f.Name)
	logrus.WithField("url", a.server.URL).WithField("path", pName).
		WithField("name", f.Name).WithField("args", f.Arguments).WithField("id", id).
		WithField("method", mName).Info("calling back")
	_, op, err := openapi.GetPathAndMethod(a.def, pName, mName)
	var callingArg openapi.CallConfig
	callingArg.ExtraAuthInfos = []*openapi.ParameterValue{
		{
			In:    "header",
			Name:  "X-Request-Id",
			Value: id,
		},
	}
	if err == nil {
		var inputs map[string]interface{}
		err = json.Unmarshal([]byte(f.Arguments), &inputs)
		if err == nil {
			var sess *openapi.CallingSession
			if body, ok := inputs["body"]; ok {
				fmt.Printf("%T", body)
				if bodyObj, ok := body.(map[string]interface{}); ok {
					var selectedFunction *openai.FunctionDefinition
					if functions != nil {
						for _, function := range functions {
							if function.Name == f.Name {
								selectedFunction = &function
								break
							}
						}
					} else if len(functions) > 0 {
						selectedFunction = &functions[0]
					}
					if selectedFunction != nil {
						if schema, ok := selectedFunction.Parameters.(*openapi3.Schema); ok {
							bodyObj = fixFormate(bodyObj, schema)
						}
					}

					callingArg.Input = bodyObj
				} else {
					callingArg.RequestBody = body
					callingArg.InputType = openapi.InputType_AsWholeRaw
				}
			}
			logrus.
				WithField("input", callingArg.Input).
				WithField("body", callingArg.RequestBody).Info("calling back with input")
			sess, err = openapi.NewCalling().
				SetPathAndMethod(pName, mName, op).
				SetServer(a.server).
				SetSecurityByDef(a.SecurityArgs, a.def).
				Do(&callingArg)
			if err == nil {
				var bts []byte
				var resp = sess.GetResponse()
				if resp != nil {
					bts, err = io.ReadAll(resp.Body)
					if err == nil {
						return pName, mName, json.RawMessage(bts), nil
					}
				} else {
					err = fmt.Errorf("no response")
				}
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
	spl := strings.SplitN(fName, "_", 2)
	pName = strings.ReplaceAll(spl[1], "_", "/")
	mName = spl[0]
	return
}

func getParameterJson(p *openapi3.Operation) *openapi3.Schema {
	//change body to parameters
	var content = p.RequestBody.Value.Content
	var typ string
	var parameters openapi3.Schema
	parameters.Type = &openapi3.Types{"object"}
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

func fixFormate(bodyObj map[string]interface{}, schema *openapi3.Schema) map[string]interface{} {
	bodySchema, ok := schema.Properties["body"]
	if ok {
		for name, property := range bodySchema.Value.Properties {
			logrus.WithField("name", name).WithField("property", property).WithField("value", bodyObj[name]).Debug("fixing formate")
			if v, ok := bodyObj[name]; ok {
				newV, ok := fixSingle(property, name, v)
				if ok {
					bodyObj[name] = newV
				}
			}
		}
	}
	return bodyObj
}

func fixSingle(property *openapi3.SchemaRef, name string, v interface{}) (interface{}, bool) {
	if property == nil || property.Value == nil {
		return nil, false
	}
	if property.Value.Type.Is("boolean") {
		if v == "true" || v == "是" {
			return true, true
		} else if v == "false" || v == "否" {
			return false, true
		} else {
			logrus.WithField("property", property).WithField("wanted type", property.Value.Type).WithField("value", v).Warn("value is not a boolean")
			return nil, false
		}
	} else if property.Value.Type.Is("number") {
		if vStr, ok := v.(string); ok {
			if vFloat, err := strconv.ParseFloat(vStr, 64); err == nil {
				return vFloat, true
			} else {
				logrus.WithField("name", name).WithField("property", property).
					WithField("wanted type", property.Value.Type).WithField("value", v).Warn("value is not a number")
			}
		}
	} else if property.Value.Type.Is("array") {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			var arr []interface{}
			for i := 0; i < rv.Len(); i++ {
				var item = rv.Index(i).Interface()
				if fixed, ok := fixSingle(property.Value.Items, name, item); ok {
					arr = append(arr, fixed)
				} else {
					arr = append(arr, item)
				}
			}
			return arr, true
		} else {
			logrus.WithField("name", name).WithField("property", property).
				WithField("wanted type", property.Value.Type).WithField("value", v).Warn("value is not a array")
		}
	} else if property.Value.Type.Is("object") {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Map {
			var obj = make(map[string]interface{})
			for _, key := range rv.MapKeys() {
				var item = rv.MapIndex(key).Interface()
				if fixed, ok := fixSingle(property.Value.Properties[key.String()], name, item); ok {
					obj[key.String()] = fixed
				} else {
					obj[key.String()] = item
				}
			}
			return obj, true
		} else {
			logrus.WithField("name", name).WithField("property", property).
				WithField("wanted type", property.Value.Type).WithField("value", v).Warn("value is not a object")
		}
	} else {
		logrus.WithField("name", name).WithField("property", property).WithField("wanted type", property.Value.Type).WithField("value", v).Warn("unknown type")
	}
	return nil, false
}
