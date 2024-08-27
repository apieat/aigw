package aistudio

import (
	"bytes"
	"encoding/json"
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

type AiStudio struct {
	// client *openai.Client
	urls  map[string]*url.URL
	token string
}

func (q *AiStudio) Init(config *platform.AIConfig) (err error) {
	// q.client = config.GetClient()
	q.token = config.GetToken()
	q.urls = make(map[string]*url.URL)
	for k, u := range config.Url {
		q.urls[k], err = url.Parse("https://aistudio.baidu.com/llm/lmapi/v1" + u)
		if err != nil {
			return err
		}
	}
	return err
}

func (q *AiStudio) GetModel(typ string) string {
	return ""
}

func (q *AiStudio) ToMessages(c platform.CompletionRequest, instructions, templates map[string]string) []openai.ChatCompletionMessage {
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

func (q *AiStudio) CreateChatCompletion(req *openai.ChatCompletionRequest, typ string) (platform.ChatCompletionResponse, error) {
	req.Model = ""
	req.Stream = false
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

	request, err := http.NewRequest(http.MethodPost, qUrl.String(), &buf)

	if err == nil {
		request.Header.Set("Authorization", "token "+q.token)
		request.Header.Set("Content-Type", "application/json")

		logrus.WithField("url", qUrl.String()).WithField("token", q.token).Debug("create chat completion request")
	}

	resp, err := http.DefaultClient.Do(request)
	if err == nil {
		var bts []byte
		bts, err = io.ReadAll(resp.Body)
		if err == nil {
			logrus.Debug("create chat completion resposne", string(bts))
			var res ChatCompletionResponseWrapper
			err = json.Unmarshal(bts, &res)
			if err == nil {
				return &res.Result, err
			}
		}
	}
	return nil, err
}

func (q *AiStudio) CreateChatStream(req *openai.ChatCompletionRequest, typ string, fn func(string)) error {
	req.Model = ""
	req.Stream = true
	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)
	encoder.SetIndent("", "")
	encoder.Encode(req)

	qUrl, ok := q.urls[typ]

	if !ok {
		if qUrl, ok = q.urls["default"]; !ok {
			var urlKeys []string
			for k := range q.urls {
				urlKeys = append(urlKeys, k)
			}
			return fmt.Errorf("url not found for type %s in %v", typ, urlKeys)
		} else {
			logrus.WithField("type", typ).WithField("url", qUrl.String()).Warnln("url not found for type, use default instead")
		}
	}

	request, err := http.NewRequest(http.MethodPost, qUrl.String(), &buf)

	if err == nil {
		request.Header.Set("Authorization", "token "+q.token)
		request.Header.Set("Content-Type", "application/json")
		//set allow stream in response
		request.Header.Set("Accept", "text/event-stream")

		logrus.WithField("url", qUrl.String()).WithField("token", q.token).Debug("create chat completion request")
	}

	resp, err := http.DefaultClient.Do(request)
	if err == nil {
		// treat response as stream
		var finished = make(chan struct{})
		go func() {
			var bts = make([]byte, 1)
			var line strings.Builder
			for {
				_, err := resp.Body.Read(bts)
				if err == io.EOF {
					break
				}
				if bts[0] == '\n' {
					var wrapper ChatCompletionResponseWrapper
					var singleLine = line.String()
					logrus.WithField("response", singleLine).Debug("read stream")
					if strings.HasPrefix(singleLine, "data:") {
						singleLine = strings.TrimPrefix(singleLine, "data:")
						err = json.Unmarshal([]byte(singleLine), &wrapper)
						if err == nil {
							fn(wrapper.Result.Result)
							if wrapper.Result.IsEnd {
								break
							}
						} else {
							logrus.WithField("response", singleLine).WithError(err).Error("read stream error")
						}
						line.Reset()
					} else {
						logrus.WithField("response", singleLine).Error("read stream format error")
					}
				} else {
					line.Write(bts)
				}
				if err != nil {
					logrus.WithError(err).Error("read stream error")
					break
				}
			}

			close(finished)
		}()
		<-finished
	}
	return err
}

func (q *AiStudio) AddFunctionsToMessage(functions []openai.FunctionDefinition, fc *openai.FunctionCall, req *openai.ChatCompletionRequest) *openai.ChatCompletionRequest {
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
	if selectedFunction != nil {
		if schema, ok := selectedFunction.Parameters.(*openapi3.Schema); ok {
			parameters = schemaToParameterDescriptions(schema).(ParameterDescriptions)
		}
	}

	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)
	encoder.SetIndent("", "")
	encoder.Encode(parameters)
	var lastMessage = req.Messages[len(req.Messages)-1]
	lastMessage.Content += "\n你的回答只能以以下JSON格式输出，格式如下：\"\"\"" + buf.String() + "\"\"\""
	req.Messages[len(req.Messages)-1] = lastMessage
	logrus.WithField("parameters", parameters).WithField("message", lastMessage).Debug("add function parameters")
	return req
}

func (q *AiStudio) AddResponseToMessage(req []openai.ChatCompletionMessage, resp platform.ChatCompletionResponse) []openai.ChatCompletionMessage {
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

func init() {
	platform.RegisterPlatform("aistudio", &AiStudio{})
}
