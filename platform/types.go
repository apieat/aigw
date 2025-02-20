package platform

type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
}

type ChatCompletionStreamChoice struct {
	Index        int                              `json:"index"`
	Delta        *ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason *string                          `json:"finish_reason"`
}

type ChatCompletionStreamChoiceDelta struct {
	Content       string      `json:"content,omitempty"`
	ReasonContent string      `json:"reason_content,omitempty"`
	Role          string      `json:"role,omitempty"`
	FunctionCall  []*ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name string `json:"name"`
		Args string `json:"arguments"`
	} `json:"function"`
}
