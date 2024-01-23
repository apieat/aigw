package chatgpt

import (
	"context"

	"github.com/apieat/aigw/platform"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Chatgpt struct {
	client *openai.Client
}

// GetModel implements platform.Platform.
func (*Chatgpt) GetModel(typ string) string {
	return openai.GPT3Dot5Turbo0613
}

func (q *Chatgpt) Init(config *platform.AIConfig) error {
	q.client = q.GetClient(config)
	return nil
}

func (q *Chatgpt) GetClient(o *platform.AIConfig) *openai.Client {

	if q.client == nil {
		var config openai.ClientConfig
		if o.Url != nil {
			u, ok := o.Url["default"]
			if !ok {
				logrus.Fatal("url not found, if you define url, please config default item 0 in url")
				return nil
			}
			config = openai.DefaultAzureConfig(o.GetToken(), u)
		} else {
			config = openai.DefaultConfig(o.GetToken())
		}
		q.client = openai.NewClientWithConfig(config)

		openai.NewClient(o.GetToken())
	}
	return q.client
}

func (q *Chatgpt) ToMessages(c platform.CompletionRequest, instructions, templates map[string]string) []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage
	var instruction = c.ToPrompt(c.GetInstruction(), instructions)
	if instruction != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: instruction,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: c.ToPrompt(c.GetPrompt(), templates),
	})
	return messages
}

func (q *Chatgpt) AddFunctionsToMessage(functions []openai.FunctionDefinition, fc *openai.FunctionCall, req *openai.ChatCompletionRequest) *openai.ChatCompletionRequest {
	if fc != nil {
		req.FunctionCall = fc
	}
	req.Functions = functions
	return req
}

func (q *Chatgpt) CreateChatCompletion(req *openai.ChatCompletionRequest, typ string) (platform.ChatCompletionResponse, error) {
	res, err := q.client.CreateChatCompletion(
		context.Background(),
		*req,
	)
	if err == nil {
		return &ResultWrapper{
			resp: &res,
		}, nil
	}
	return nil, err
}

func (q *Chatgpt) AddResponseToMessage(req []openai.ChatCompletionMessage, resp platform.ChatCompletionResponse) []openai.ChatCompletionMessage {
	return req
}

func init() {
	platform.RegisterPlatform("chatgpt", &Chatgpt{})
}
