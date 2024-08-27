package qianfan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/apieat/aigw/platform"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Qianfan struct {
	// client *openai.Client
	urls  map[string]*url.URL
	token string
}

// GetModel implements platform.Platform.
func (*Qianfan) GetModel(typ string) string {
	return ""
}

func (q *Qianfan) Init(config *platform.AIConfig) (err error) {
	// q.client = config.GetClient()
	q.token = config.GetToken()
	q.urls = make(map[string]*url.URL)
	for k, u := range config.Url {
		q.urls[k], err = url.Parse("https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop" + u + "?access_token=" + q.token)
		if err != nil {
			return err
		}
	}
	if q.urls["embeddings"] == nil {
		q.urls["embeddings"], err = url.Parse("https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1?access_token=" + q.token)
		return err
	}
	return err
}

func (q *Qianfan) ToMessages(c platform.CompletionRequest, instructions, templates map[string]string) []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage
	// var content string
	var instruction = c.GetInstruction()
	// if instruction != "" {
	content := c.ToPrompt(instruction, instructions)
	// }
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content + ", " + c.ToPrompt(c.GetPrompt(), templates),
	})
	return messages
}

type ParameterDescriptions map[string]interface{}

type ParameterDescription struct {
	Description interface{}   `json:"description"`
	Enums       []interface{} `json:"enum"`
	Type        string        `json:"type"`
}

func (p *ParameterDescription) MarshalJSON() ([]byte, error) {

	if p.Type == "array" {
		switch td := p.Description.(type) {
		case ParameterDescriptions:
			bts, err := json.Marshal(td)
			if err != nil {
				return nil, err
			}
			return []byte(fmt.Sprintf(`[%s]`, string(bts))), nil
		default:
			return []byte(fmt.Sprintf(`["%v"]`, p.Description)), nil
		}
	} else {
		var enumDescriptionPart string
		if len(p.Enums) > 0 {
			var enumDescriptions []string
			for _, enum := range p.Enums {
				enumDescriptions = append(enumDescriptions, fmt.Sprintf("%v", enum))
			}
			enumDescriptionPart = "(" + strings.Join(enumDescriptions, "|") + ")"
		}
		return []byte(fmt.Sprintf(`"[%s%s]"`, p.Description, enumDescriptionPart)), nil
	}

}

func (q *Qianfan) CreateChatCompletion(req *openai.ChatCompletionRequest, typ string) (platform.ChatCompletionResponse, error) {
	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)
	encoder.SetIndent("", "")
	encoder.Encode(req)

	qUrl, ok := q.urls[typ]

	if !ok {
		if qUrl, ok = q.urls["default"]; !ok {
			return nil, fmt.Errorf("url not found for type %s", typ)
		} else {
			logrus.WithField("type", typ).WithField("url", qUrl.String()).Warnln("url not found for type, use default instead")
		}
	}

	resp, err := http.Post(qUrl.String(), "application/json", &buf)
	if err == nil {
		var bts []byte
		bts, err = io.ReadAll(resp.Body)
		if err == nil {
			logrus.Debug("create chat completion resposne", string(bts))
			var res ChatCompletionResponse
			err = json.Unmarshal(bts, &res)
			if err == nil {
				return &res, err
			}
		}
	}
	return nil, err
}

func (q *Qianfan) AddFunctionsToMessage(functions []openai.FunctionDefinition, fc *openai.FunctionCall, req *openai.ChatCompletionRequest) *openai.ChatCompletionRequest {
	var selectedFunction *openai.FunctionDefinition
	if fc != nil {
		for _, function := range functions {
			if function.Name == fc.Name {
				selectedFunction = &function
				break
			}
		}
	} else if len(functions) > 0 {
		selectedFunction = &functions[0]
	}
	var parameters ParameterDescriptions
	if schema, ok := selectedFunction.Parameters.(*openapi3.Schema); ok {
		parameters = schemaToParameterDescriptions(schema).(ParameterDescriptions)
	}

	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)
	encoder.SetIndent("", "")
	encoder.Encode(parameters)
	var lastMessage = req.Messages[len(req.Messages)-1]
	lastMessage.Content += "\n输出只能以以下JSON格式输出，格式如下：\"\"\"" + buf.String() + "\"\"\""
	req.Messages[len(req.Messages)-1] = lastMessage
	logrus.WithField("parameters", parameters).WithField("message", lastMessage).Debug("add function parameters")
	return req
}

func (q *Qianfan) AddResponseToMessage(req []openai.ChatCompletionMessage, resp platform.ChatCompletionResponse) []openai.ChatCompletionMessage {
	if tr, ok := resp.(*ChatCompletionResponse); ok {
		req = append(req, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: tr.Result,
		})
	} else {
		fmt.Printf("resp is not ChatCompletionResponse,resp is %T", resp)
	}
	return req
}

func schemaToParameterDescriptions(schema *openapi3.Schema) interface{} {

	if schema.Type == "object" {
		var parameters = make(ParameterDescriptions)

		for name, property := range schema.Properties {
			switch property.Value.Type {
			case "array":
				parameters[name] = &ParameterDescription{
					Description: schemaToParameterDescriptions(property.Value.Items.Value),
					Type:        property.Value.Type,
				}
			case "object":
				parameters[name] = schemaToParameterDescriptions(property.Value)
			default:
				var pDes = &ParameterDescription{
					Description: property.Value.Description,
					Type:        property.Value.Type,
				}
				if property.Value.Enum != nil {
					pDes.Enums = property.Value.Enum
				}
				parameters[name] = pDes
			}
		}
		return parameters
	} else {
		return &ParameterDescription{
			Description: schema.Description,
			Type:        schema.Type,
		}
	}
}

type EmbeddingRequest struct {
	Input []string `json:"input"`
}

func (b *Qianfan) Embed(input ...string) ([][]float32, error) {

	var reqData EmbeddingRequest
	reqData.Input = input
	payloadBytes, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}
	payload := bytes.NewReader(payloadBytes)
	client := &http.Client{}

	url := b.urls["embeddings"].String()

	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response ChatCompletionResponse
	json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	if response.ErrorCode != 0 {
		return nil, fmt.Errorf("error code: %d, error msg: %s", response.ErrorCode, response.ErrorMsg)
	}
	var embeddings [][]float32
	for _, data := range response.Data {
		embeddings = append(embeddings, data.Embedding)
	}
	return embeddings, nil
}

func (q *Qianfan) CreateChatStream(req *openai.ChatCompletionRequest, typ string, fn func(string)) error {
	return errors.New("not implemented")
}

func init() {
	platform.RegisterPlatform("qianfan", &Qianfan{})
}
