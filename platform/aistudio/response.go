package aistudio

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
	ID           string       `json:"id"`
	Object       string       `json:"object"`
	Created      int64        `json:"created,string"`
	Model        string       `json:"model"`
	Result       string       `json:"result"`
	Usage        openai.Usage `json:"usage"`
	SentenceId   string       `json:"sentence_id"`
	IsEnd        bool         `json:"is_end"`
	IsTruncated  bool         `json:"is_truncated"`
	FinishReason string       `json:"finish_reason"`
	NeedClear    bool         `json:"need_clear_history"`
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
	jsonStr = findLineBreakAfterComments(jsonStr)
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\r", "\\n")
	logrus.Info("jsonStr", jsonStr)
	return &openai.FunctionCall{
		Name:      reqFc.Name,
		Arguments: jsonStr,
	}, nil
}

var jsonArrayBodyCommaMatcher = regexp.MustCompile(`^{\s+"body":\s*\[\s*{`)
var jsonArrayBodyEndCommaMatcher = regexp.MustCompile(`}\s*\]\s*}$`)

func findLineBreakAfterComments(jsonStr string) string {
	nextComment := strings.Index(jsonStr, "//")
	for nextComment != -1 {
		nextLineBreak := strings.Index(jsonStr[nextComment:], "\n")
		if nextLineBreak != -1 {
			jsonStr = jsonStr[:nextComment] + "\r" + jsonStr[nextComment+nextLineBreak+1:]
		}
		offset := strings.Index(jsonStr[nextComment+2:], "//")
		if offset != -1 {
			nextComment = nextComment + 2 + offset
		} else {
			nextComment = -1
		}
	}
	return jsonStr
}

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
