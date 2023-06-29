package ctrl

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/extrame/goblet"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Completion struct {
	goblet.SingleController
}

func (c *Completion) Post(ctx *goblet.Context, arg CompletionRequest) error {
	client := openaiCfg.GetClient()

	if arg.Prompt == "" {
		return errors.New("prompt is empty")
	}
	logrus.WithField("prompt", arg.ToPrompt(openaiCfg.Templates)).WithField("functions", apiCfg.GetFunctions()).Debug("get completion request")

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo0613,
			Functions: apiCfg.GetFunctions(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: arg.ToPrompt(openaiCfg.Templates),
				},
			},
		},
	)
	if err == nil {
		var apiResp json.RawMessage
		var pName, mName string
		logrus.WithField("call", resp.Choices[0].Message).Debug("call")
		var fc = resp.Choices[0].Message.FunctionCall
		if fc != nil {
			pName, mName, apiResp, err = apiCfg.Call(arg.Id, resp.Choices[0].Message.FunctionCall)
			if err == nil {
				resp, err = client.CreateChatCompletion(
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
				if err == nil {
					ctx.AddRespond(resp.Choices[0].Message.Content)
				}
			}
		} else {
			raw, _ := json.Marshal(resp.Choices[0].Message)
			return errors.New("no function call is responded in" + string(raw))
		}
	}
	return err
}
