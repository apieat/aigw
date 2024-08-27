package zhipu

import (
	"errors"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ChatCompletionResponse openai.ChatCompletionResponse

func (r *ChatCompletionResponse) GetFunctionCallArguments(reqFc *openai.FunctionCall) (*openai.FunctionCall, error) {
	logrus.Info("get function call arguments", reqFc, r)
	var jsonStr string
	if len(r.Choices) == 0 {
		return nil, errors.New("no choices")
	}
	var message = r.Choices[0].Message.Content
	if strings.Contains(message, "```json") {
		_, jsonStr, _ = strings.Cut(message, "```json")
		jsonStr, _, _ = strings.Cut(jsonStr, "```")
	} else {
		jsonStr = message
	}
	if strings.Contains(jsonStr, "/*") {
		var after string
		jsonStr, after, _ = strings.Cut(jsonStr, "/*")
		_, after, _ = strings.Cut(after, "*/")
		jsonStr += after
	}
	jsonStr = findLineBreakAfterComments(jsonStr)
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\r", "\n")
	logrus.Info("jsonStr fixed", jsonStr)
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
		offset := strings.Index(jsonStr, "//")
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
	return &r.Choices[0].Message
}
