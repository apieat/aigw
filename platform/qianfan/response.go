package qianfan

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Result  string       `json:"result"`
	Usage   openai.Usage `json:"usage"`
}

func (r *ChatCompletionResponse) GetFunctionCallArguments(reqFc *openai.FunctionCall) (*openai.FunctionCall, error) {
	var jsonStr = strings.TrimPrefix(r.Result, "```json")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)
	return &openai.FunctionCall{
		Name:      reqFc.Name,
		Arguments: jsonStr,
	}, nil
}

func (r *ChatCompletionResponse) GetMessage() *openai.ChatCompletionMessage {
	return &openai.ChatCompletionMessage{
		Content: r.Result,
	}
}
