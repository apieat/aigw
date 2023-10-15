package chatgpt

import (
	"context"

	"github.com/apieat/aigw/platform"
	"github.com/sashabaranov/go-openai"
)

type Chatgpt struct {
	client *openai.Client
}

func (q *Chatgpt) Init(config *platform.AIConfig) error {
	q.client = config.GetClient()
	return nil
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

func (q *Chatgpt) CreateChatCompletion(req *openai.ChatCompletionRequest) (platform.ChatCompletionResponse, error) {
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

func init() {
	platform.RegisterPlatform("chatgpt", &Chatgpt{})
}
