package ctrl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/apieat/aigw"
	"github.com/apieat/aigw/config"
	"github.com/apieat/aigw/errors"
	"github.com/apieat/aigw/platform"
	"github.com/apieat/aigw/stream"
	"github.com/extrame/goblet"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Completion struct {
	goblet.SingleController
	Config *config.Config
}

func (c *Completion) Post(ctx *goblet.Context, arg aigw.CompletionRequest) error {

	if arg.Prompt == "" {
		return errors.ErrorEmptyPrompt
	}

	var functions = c.Config.Api.GetFunctions(arg.Functions)
	var fc *openai.FunctionCall
	if len(functions) == 1 {
		fc = &openai.FunctionCall{
			Name: functions[0].Name,
		}
	}

	logrus.WithField("prompt", arg.ToPrompt(arg.Prompt, c.Config.Platform.Templates)).WithField("instruction", arg.ToPrompt(arg.Instruction, c.Config.Platform.Instructions)).
		WithField("functions_filter", arg.Functions).
		WithField("functions", functions).
		WithField("temparature", arg.GetTemparature()).
		Debug("get completion request")

	if !arg.Debug {

		var req = &openai.ChatCompletionRequest{
			Model:       c.Config.Platform.GetModel(arg.Type),
			MaxTokens:   c.Config.Platform.MaxTokens,
			Messages:    c.Config.Platform.ToMessages(&arg), //.ToMessages(c.Config.Platform.Instructions, c.Config.Platform.Templates),
			Temperature: arg.GetTemparature(),
		}

		req = c.Config.Platform.AddFunctionsToMessage(functions, fc, req)

		if c.Config.Platform.Sync {
			if arg.Stream {
				var writer = ctx.Writer()
				_, ok := ctx.Writer().(http.Flusher)
				if !ok {
					return errors.ErrorStreamingNotSupported
				}
				return c.handleStream(req, arg.Type, c.Config.Platform.AskAiToAnalyse, writer, ctx)
			} else {
				return c.handleCallback(req, functions, fc, arg.Id, arg.Mode, arg.Type, ctx)
			}
		} else {
			go c.handleCallback(req, functions, fc, arg.Id, arg.Mode, arg.Type, nil)
		}
		logrus.Debug("completion finished")
		return nil
	} else {
		ctx.AddRespond("functions", functions)
		ctx.AddRespond("prompt", arg.ToPrompt(arg.Prompt, c.Config.Platform.Templates))
		return nil
	}
}

func (c *Completion) handleCallback(req *openai.ChatCompletionRequest, functions []openai.FunctionDefinition, fc *openai.FunctionCall, id, mode, typ string, ctx *goblet.Context) (err error) {
	var apiResp json.RawMessage
	var pName, mName string
	var originalMessages = req.Messages
	retry := 0
	for retry < 3 {
		logrus.WithField("retry", retry).WithField("req", req).WithField("type", typ).Debug("create chat completion")
		resp, err := c.Config.Platform.CreateChatCompletion(req, typ)
		if err != nil {
			retry++
			logrus.Info("create chat completion failed", err.Error())
			continue
		}
		var args *openai.FunctionCall
		args, err = resp.GetFunctionCallArguments(fc)
		if err != nil {
			logrus.WithField("resp", resp).WithError(err).Errorln("create chat completion failed")
			retry++
			continue
		}
		if args != nil {
			pName, mName, apiResp, err = c.Config.Api.Call(id, functions, args)
			if err == nil {
				var errMessage errors.Error
				err = json.Unmarshal(apiResp, &errMessage)
				if err != nil {
					logrus.WithError(err).WithField("resp", string(apiResp)).Errorln("unmarshal api response failed")
				}
				if errMessage.Code == errors.InvalidResponse {
					logrus.WithField("resp", string(apiResp)).Errorln("invalid response,retry")
					retry++
					if errMessage.Reason != "" {
						originalMessages = c.Config.Platform.AddResponseToMessage(originalMessages, resp)

						req.Messages = append(originalMessages, openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleUser,
							Content: fmt.Sprintf("last response is error,error is '%s', retry no.%d", errMessage.Reason, retry),
						},
						)
						logrus.WithField("reason", errMessage.Reason).WithField("messages", req.Messages).Errorln("add reason to messages")
					}
					continue
				}
				if c.Config.Platform.AskAiToAnalyse {
					var analyseResp platform.ChatCompletionResponse
					analyseResp, err = c.Config.Platform.CreateChatCompletion(
						&openai.ChatCompletionRequest{
							Model:     openai.GPT3Dot5Turbo0613,
							Functions: c.Config.Api.GetFunctionByName(pName, mName),
							MaxTokens: c.Config.Platform.MaxTokens,
							Messages: []openai.ChatCompletionMessage{
								{
									Role:    openai.ChatMessageRoleFunction,
									Name:    args.Name,
									Content: string(apiResp),
								},
							},
						}, typ,
					)
					if err == nil && ctx != nil {
						ctx.Respond(analyseResp.GetMessage())
					}
				}
			}
			logrus.WithField("id", id).WithError(err).Errorln("function call finish")
			return err
		} else {
			raw, _ := json.Marshal(resp)
			logrus.WithField("id", id).Errorln("no function call is responded in", string(raw))
			return errors.NoFunctionCall(string(raw))
		}
	}
	return errors.ErrorTooManyRetry
}

func (c *Completion) handleStream(req *openai.ChatCompletionRequest, typ string, aiToAnalyse bool, writer http.ResponseWriter, ctx *goblet.Context) (err error) {

	var builder stream.Builder

	logrus.WithField("req", req).WithField("type", typ).Debug("create chat completion")
	err = c.Config.Platform.CreateChatStream(req, typ, func(s string) {
		for _, c := range s {
			builder.AppendRune(c)
		}
		stat := builder.Stat()
		bts, err := json.Marshal(stat)
		if err == nil {
			wrapper, _ := FormatServerSentEvent("doing", string(bts))
			writer.Write([]byte(wrapper))
		}
		writer.(http.Flusher).Flush()
	})

	// send done
	wrapper, _ := FormatServerSentEvent("done", "")
	writer.Write([]byte(wrapper))
	writer.(http.Flusher).Flush()

	if err != nil {
		return err
	}

	return errors.ErrorTooManyRetry
}

func FormatServerSentEvent(event string, data any) (string, error) {

	buff := bytes.NewBuffer([]byte{})

	encoder := json.NewEncoder(buff)

	err := encoder.Encode(data)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("event: %s\n", event))
	sb.WriteString(fmt.Sprintf("data: %v\n\n", buff.String()))

	return sb.String(), nil
}
