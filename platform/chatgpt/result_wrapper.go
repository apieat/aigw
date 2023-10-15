package chatgpt

import (
	"errors"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ResultWrapper struct {
	resp *openai.ChatCompletionResponse
}

func (r *ResultWrapper) GetFunctionCallArguments(*openai.FunctionCall) (*openai.FunctionCall, error) {
	if len(r.resp.Choices) == 0 {
		return nil, errors.New("create chat completion failed(empty choices)")
	}
	logrus.WithField("call", r.resp.Choices[0].Message).WithField("tokens", r.resp.Usage).Debug("call")
	var fc = r.resp.Choices[0].Message.FunctionCall
	if fc != nil {
		tryToCleanJsonError(r.resp.Choices[0].Message.FunctionCall)
	}
	return r.resp.Choices[0].Message.FunctionCall, nil
}

func (r *ResultWrapper) GetMessage() *openai.ChatCompletionMessage {
	return &r.resp.Choices[0].Message
}

var jsonLastCommaMatcher = regexp.MustCompile(`,\s*}\s*}$`)

func tryToCleanJsonError(fc *openai.FunctionCall) {
	var matched = jsonLastCommaMatcher.FindString(fc.Arguments)
	if matched != "" {
		fc.Arguments = strings.Replace(fc.Arguments, matched, "}}", 1)
	}
}
