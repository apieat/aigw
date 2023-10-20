package platform

import "github.com/sashabaranov/go-openai"

var platforms = make(map[string]Platform)
var Current Platform

type Platform interface {
	ToMessages(c CompletionRequest, instructions, templates map[string]string) []openai.ChatCompletionMessage
	AddFunctionsToMessage(functions []openai.FunctionDefinition, fc *openai.FunctionCall, req *openai.ChatCompletionRequest) *openai.ChatCompletionRequest
	CreateChatCompletion(req *openai.ChatCompletionRequest) (ChatCompletionResponse, error)
	//将系统响应加入到消息列表中，部分平台在多次提交时需要将上次的响应加入到消息列表中
	AddResponseToMessage(messages []openai.ChatCompletionMessage, resp ChatCompletionResponse) []openai.ChatCompletionMessage
	Init(cfg *AIConfig) error
}

type CompletionRequest interface {
	ToPrompt(prompt string, templates ...map[string]string) string
	GetInstruction() string
	GetPrompt() string
}

type ChatCompletionResponse interface {
	GetFunctionCallArguments(*openai.FunctionCall) (*openai.FunctionCall, error)
	GetMessage() *openai.ChatCompletionMessage
}

func RegisterPlatform(name string, platform Platform) {
	platforms[name] = platform
}

func Init(name string) {
	Current = platforms[name]
}
