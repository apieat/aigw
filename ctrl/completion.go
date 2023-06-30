package ctrl

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/apieat/aigw"
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
		return errors.New("prompt is empty")
	}
	logrus.WithField("prompt", arg.ToPrompt(openaiCfg.Templates)).WithField("functions", apiCfg.GetFunctions(arg.Functions)).Debug("get completion request")

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo0613,
			Functions: apiCfg.GetFunctions(arg.Functions),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: arg.ToPrompt(openaiCfg.Templates),
				},
			},
			Temperature: 0.9,
		},
	)
	if err == nil {
		if openaiCfg.Sync {
			return handleCallback(&resp, arg.Id, client, ctx)
		} else {
			go handleCallback(&resp, arg.Id, client, nil)
		}
	}
	logrus.Debug("completion finished")
	return err
}

func handleCallback(resp *openai.ChatCompletionResponse, id string, client *openai.Client, ctx *goblet.Context) (err error) {
	var apiResp json.RawMessage
	var pName, mName string
	logrus.WithField("call", resp.Choices[0].Message).Debug("call")
	var fc = resp.Choices[0].Message.FunctionCall
	if fc != nil {
		pName, mName, apiResp, err = apiCfg.Call(id, resp.Choices[0].Message.FunctionCall)
		if err == nil {
			var analyseResp openai.ChatCompletionResponse
			analyseResp, err = client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:     openai.GPT3Dot5Turbo0613,
					Functions: apiCfg.GetFunctionByName(pName, mName),
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
		logrus.WithField("id", id).WithError(err).Errorln("function call error")
		return err
	} else {
		raw, _ := json.Marshal(resp.Choices[0].Message)
		logrus.WithField("id", id).Errorln("no function call is responded in", string(raw))
		return errors.New("no function call is responded in" + string(raw))
	}
}
