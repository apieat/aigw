package ctrl

import (
	"context"
	"encoding/json"

	"github.com/apieat/aigw"
	"github.com/apieat/aigw/errors"
	"github.com/extrame/goblet"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Completion struct {
	goblet.SingleController
}

func (c *Completion) Post(ctx *goblet.Context, arg aigw.CompletionRequest) error {

	client := openaiCfg.GetClient()

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

		var req = openai.ChatCompletionRequest{
			Model:        openai.GPT3Dot5Turbo0613,
			Functions:    functions,
			MaxTokens:    openaiCfg.MaxTokens,
			Messages:     arg.ToMessages(openaiCfg.Instructions, openaiCfg.Templates),
			Temperature:  arg.GetTemparature(),
			FunctionCall: fc,
		}

		if openaiCfg.Sync {
			return handleCallback(&req, arg.Id, client, openaiCfg.AskAiToAnalyse, ctx)
		} else {
			go handleCallback(&req, arg.Id, client, false, nil)
		}
		logrus.Debug("completion finished")
		return nil
	} else {
		ctx.AddRespond("functions", functions)
		ctx.AddRespond("prompt", arg.ToPrompt(arg.Prompt, openaiCfg.Templates))
		return nil
	}
}

func handleCallback(req *openai.ChatCompletionRequest, id string, client *openai.Client, aiToAnalyse bool, ctx *goblet.Context) (err error) {
	var apiResp json.RawMessage
	var pName, mName string
	retry := 0
	for retry < 3 {
		resp, err := client.CreateChatCompletion(
			context.Background(),
			*req,
		)
		if err != nil {
			logrus.WithError(err).Errorln("create chat completion failed")
			retry++
			continue
		}
		logrus.WithField("call", resp.Choices[0].Message).WithField("tokens", resp.Usage).Debug("call")
		var fc = resp.Choices[0].Message.FunctionCall
		if fc != nil {
			pName, mName, apiResp, err = apiCfg.Call(id, resp.Choices[0].Message.FunctionCall)
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
						req.Messages = append(req.Messages, openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleSystem,
							Content: errMessage.Reason,
						},
						)
					}
					continue
				}
				if aiToAnalyse {
					var analyseResp openai.ChatCompletionResponse
					analyseResp, err = client.CreateChatCompletion(
						context.Background(),
						openai.ChatCompletionRequest{
							Model:     openai.GPT3Dot5Turbo0613,
							Functions: apiCfg.GetFunctionByName(pName, mName),
							MaxTokens: openaiCfg.MaxTokens,
							Messages: []openai.ChatCompletionMessage{
								{
									Role:    openai.ChatMessageRoleFunction,
									Name:    resp.Choices[0].Message.FunctionCall.Name,
									Content: string(apiResp),
								},
							},
						},
					)
					if err == nil && ctx != nil {
						ctx.Respond(analyseResp.Choices[0].Message)
					}
				}
			}
			logrus.WithField("id", id).WithError(err).Errorln("function call finish")
			return err
		} else {
			raw, _ := json.Marshal(resp.Choices[0].Message)
			logrus.WithField("id", id).Errorln("no function call is responded in", string(raw))
			return errors.NoFunctionCall(string(raw))
		}
	}
	return errors.ErrorTooManyRetry
}
