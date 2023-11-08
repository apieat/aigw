package qianfan

import (
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ChatCompletionResponseWrapper struct {
	LogId     string                 `json:"logId"`
	ErrorCode int                    `json:"errorCode"`
	ErrorMsg  string                 `json:"errorMsg"`
	Result    ChatCompletionResponse `json:"result"`
}

type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Result  string       `json:"result"`
	Usage   openai.Usage `json:"usage"`
}

func (r *ChatCompletionResponse) GetFunctionCallArguments(reqFc *openai.FunctionCall) (*openai.FunctionCall, error) {
	logrus.Info("get function call arguments", reqFc, r)
	var jsonStr string
	if strings.Contains(r.Result, "```json") {
		_, jsonStr, _ = strings.Cut(r.Result, "```json")
		jsonStr, _, _ = strings.Cut(jsonStr, "```")
	} else {
		jsonStr = r.Result
	}
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	// jsonStr = tryToCleanJsonError(strings.TrimSpace(jsonStr))
	return &openai.FunctionCall{
		Name:      reqFc.Name,
		Arguments: jsonStr,
	}, nil
}

var jsonArrayBodyCommaMatcher = regexp.MustCompile(`^{\s+"body":\s*\[\s*{`)
var jsonArrayBodyEndCommaMatcher = regexp.MustCompile(`}\s*\]\s*}$`)

func tryToCleanJsonError(jsonStr string) string {
	var matched = jsonArrayBodyCommaMatcher.FindString(jsonStr)
	if matched != "" {
		jsonStr = strings.Replace(jsonStr, matched, `{"body":{`, 1)
		matched = jsonArrayBodyEndCommaMatcher.FindString(jsonStr)
		if matched != "" {
			jsonStr = strings.Replace(jsonStr, matched, "}}", 1)
		}
	}
	return jsonStr
}

func (r *ChatCompletionResponse) GetMessage() *openai.ChatCompletionMessage {
	return &openai.ChatCompletionMessage{
		Content: r.Result,
	}
}
