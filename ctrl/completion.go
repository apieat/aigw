package ctrl

import (
	"encoding/json"
	"fmt"

	"github.com/apieat/aigw"
	"github.com/apieat/aigw/errors"
	"github.com/apieat/aigw/platform"
	"github.com/extrame/goblet"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Completion struct {
	goblet.SingleController
}

func (c *Completion) Init(server *goblet.Server) error {
	return platform.Current.Init(&openaiCfg)
}

func (c *Completion) Post(ctx *goblet.Context, arg aigw.CompletionRequest) error {

	if arg.Prompt == "" {
		return errors.ErrorEmptyPrompt
	}

	var functions = apiCfg.GetFunctions(arg.Functions)
	var fc *openai.FunctionCall
	if len(functions) == 1 {
		fc = &openai.FunctionCall{
			Name: functions[0].Name,
		}
	}

	logrus.WithField("prompt", arg.ToPrompt(arg.Prompt, openaiCfg.Templates)).WithField("instruction", arg.ToPrompt(arg.Instruction, openaiCfg.Instructions)).
		WithField("functions_filter", arg.Functions).
		WithField("functions", functions).
		WithField("temparature", arg.GetTemparature()).
		Debug("get completion request")

	if !arg.Debug {

		var req = &openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo0613,
			MaxTokens:   openaiCfg.MaxTokens,
			Messages:    arg.ToMessages(openaiCfg.Instructions, openaiCfg.Templates),
			Temperature: arg.GetTemparature(),
		}

		req = platform.Current.AddFunctionsToMessage(functions, fc, req)

		if openaiCfg.Sync {
			return handleCallback(req, functions, fc, arg.Id, openaiCfg.AskAiToAnalyse, ctx)
		} else {
			go handleCallback(req, functions, fc, arg.Id, false, nil)
		}
		logrus.Debug("completion finished")
		return nil
	} else {
		ctx.AddRespond("functions", functions)
		ctx.AddRespond("prompt", arg.ToPrompt(arg.Prompt, openaiCfg.Templates))
		return nil
	}
}

func handleCallback(req *openai.ChatCompletionRequest, functions []openai.FunctionDefinition, fc *openai.FunctionCall, id string, aiToAnalyse bool, ctx *goblet.Context) (err error) {
	var apiResp json.RawMessage
	var pName, mName string
	var originalMessages = req.Messages
	retry := 0
	for retry < 3 {
		logrus.WithField("retry", retry).WithField("req", req).Debug("create chat completion")
		resp, err := platform.Current.CreateChatCompletion(req)
		if err != nil {
			retry++
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
			pName, mName, apiResp, err = apiCfg.Call(id, functions, args)
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
						originalMessages = platform.Current.AddResponseToMessage(originalMessages, resp)

						req.Messages = append(originalMessages, openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleUser,
							Content: fmt.Sprintf("last response is error,error is '%s', retry no.%d", errMessage.Reason, retry),
						},
						)
						logrus.WithField("reason", errMessage.Reason).WithField("messages", req.Messages).Errorln("add reason to messages")
					}
					continue
				}
				if aiToAnalyse {
					var analyseResp platform.ChatCompletionResponse
					analyseResp, err = platform.Current.CreateChatCompletion(
						&openai.ChatCompletionRequest{
							Model:     openai.GPT3Dot5Turbo0613,
							Functions: apiCfg.GetFunctionByName(pName, mName),
							MaxTokens: openaiCfg.MaxTokens,
							Messages: []openai.ChatCompletionMessage{
								{
									Role:    openai.ChatMessageRoleFunction,
									Name:    args.Name,
									Content: string(apiResp),
								},
							},
						},
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
