package deepseek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/apieat/aigw/platform"
	"github.com/extrame/jose/jws"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type DeepSeek struct {
	// client *openai.Client
	url   *url.URL
	token string
}

// GetModel implements platform.Platform.
func (*DeepSeek) GetModel(typ string) string {
	return "deepseek-reasoner"
}

func (q *DeepSeek) Init(config *platform.AIConfig) (err error) {
	// q.client = config.GetClient()
	q.token = config.GetToken()
	q.url, _ = url.Parse("https://api.deepseek.com/chat/completions")
	return nil
}

func (q *DeepSeek) ToMessages(c platform.CompletionRequest, instructions, templates map[string]string) []openai.ChatCompletionMessage {
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
	Description interface{}     `json:"description"`
	Enums       []interface{}   `json:"enum"`
	Type        *openapi3.Types `json:"type"`
}

func (p *ParameterDescription) MarshalJSON() ([]byte, error) {

	if p.Type.Is("array") {
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

func (q *DeepSeek) CreateChatCompletion(req *openai.ChatCompletionRequest, typ string) (platform.ChatCompletionResponse, error) {
	req.Stream = false
	if req.Temperature <= 0 || req.Temperature >= 1 {
		req.Temperature = 0.95
	}
	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)
	encoder.SetIndent("", "")
	encoder.Encode(req)

	// qUrl, ok := q.urls[typ]

	// if !ok {
	// if qUrl, ok = q.urls["default"]; !ok {
	// 	return nil, fmt.Errorf("url not found for type %s", typ)
	// } else {
	// 	logrus.WithField("type", typ).WithField("url", qUrl.String()).Warnln("url not found for type, use default instead")
	// }
	// }

	request, err := http.NewRequest(http.MethodPost, q.url.String(), &buf)

	if err == nil {
		request.Header.Set("Authorization", "Bearer "+q.token)
		request.Header.Set("Content-Type", "application/json")

		logrus.WithField("url", q.url.String()).WithField("token", q.token).Debug("create chat completion request")
	}

	resp, err := http.DefaultClient.Do(request)
	if err == nil {
		var bts []byte
		bts, err = io.ReadAll(resp.Body)
		if err == nil {
			logrus.Debug("create chat completion resposne", string(bts))
			var res ChatCompletionResponse
			err = json.Unmarshal(bts, &res)
			logrus.Debug("parsed resposne", res, err)
			if err == nil {
				return &res, err
			}
		}
	}
	return nil, err
}

func (q *DeepSeek) AddFunctionsToMessage(functions []openai.FunctionDefinition, fc *openai.FunctionCall, req *openai.ChatCompletionRequest) *openai.ChatCompletionRequest {
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
	lastMessage.Content += "\n你的回答只能以以下JSON格式输出，JSON里面不能有注释，格式如下：\"\"\"" + buf.String() + "\"\"\""
	req.Messages[len(req.Messages)-1] = lastMessage
	logrus.WithField("parameters", parameters).WithField("message", lastMessage).Debug("add function parameters")
	return req
}

func (q *DeepSeek) AddResponseToMessage(req []openai.ChatCompletionMessage, resp platform.ChatCompletionResponse) []openai.ChatCompletionMessage {
	if tr, ok := resp.(*ChatCompletionResponse); ok {
		req = append(req, tr.Choices[0].Message)
	} else {
		fmt.Printf("resp is not ChatCompletionResponse,resp is %T", resp)
	}
	return req
}

func (q *DeepSeek) CreateChatStream(req *openai.ChatCompletionRequest, typ string, fn func(string, string)) error {
	req.Stream = true
	var buf bytes.Buffer
	var encoder = json.NewEncoder(&buf)
	encoder.SetIndent("", "")
	encoder.Encode(req)

	request, err := http.NewRequest(http.MethodPost, q.url.String(), &buf)

	if err == nil {
		request.Header.Set("Authorization", "Bearer "+q.token)
		request.Header.Set("Content-Type", "application/json")
		//set allow stream in response
		request.Header.Set("Accept", "text/event-stream")

		logrus.WithField("url", q.url.String()).WithField("token", q.token).Debug("create chat completion request")
	} else {
		logrus.WithError(err).Error("create chat completion request error")
		return err
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
				// fmt.Print(string(bts))
				if bts[0] == '\n' || err == io.EOF {
					var wrapper platform.ChatCompletionStreamResponse
					var singleLine = line.String()
					logrus.WithField("response", singleLine).Info("read stream")
					if strings.HasPrefix(singleLine, "data:") {
						singleLine = strings.TrimPrefix(singleLine, "data:")
						logrus.Info("read stream data", singleLine)
						err = json.Unmarshal([]byte(singleLine), &wrapper)
						if err == nil {
							for _, c := range wrapper.Choices {
								if c.Delta != nil {
									fn(c.Delta.Content, c.Delta.ReasonContent)
								}
								if c.FinishReason != nil {
									break
								}
							}

						} else {
							logrus.WithField("response", singleLine).WithError(err).Error("read stream error")
						}
						line.Reset()
					} else {
						logrus.WithField("response", singleLine).Error("read stream format error")
						line.Reset()
					}
					if err == io.EOF {
						break
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
	} else {
		logrus.WithError(err).Debug("create chat completion request error")
	}
	return err
}

func schemaToParameterDescriptions(schema *openapi3.Schema) interface{} {

	if schema.Type.Is("object") {
		var parameters = make(ParameterDescriptions)

		for name, property := range schema.Properties {
			if property.Value.Type.Is("array") {
				parameters[name] = &ParameterDescription{
					Description: schemaToParameterDescriptions(property.Value.Items.Value),
					Type:        property.Value.Type,
				}
			} else if property.Value.Type.Is("object") {
				parameters[name] = schemaToParameterDescriptions(property.Value)
			} else {
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

func jwtEncode(token string) string {
	id, secret, right := strings.Cut(token, ".")
	if !right {
		logrus.WithField("token", token).Errorln("token is invalid")
		return ""
	}
	var claims = make(jws.Claims)
	var now = time.Now()
	claims.Set("api_key", id)
	claims.Set("exp", now.Add(1*time.Hour).Unix()*1000)
	claims.Set("timestamp", now.Unix()*1000)
	logrus.Info("claims", claims)
	j := jws.NewJWT(claims, jws.GetSigningMethod("HS256"))

	var secretBts = []byte(secret)

	b, err := j.Serialize(secretBts)
	if err == nil {
		return string(b)
	}
	logrus.WithError(err).WithField("token", token).Errorln("jwt encode failed")
	return ""
}

func init() {
	platform.RegisterPlatform("deepseek", &DeepSeek{})
}
