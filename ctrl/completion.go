package ctrl

import (
	"context"
	"encoding/json"

	"github.com/extrame/goblet"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Completion struct {
	goblet.SingleController
}

func (this *Completion) Post(ctx *goblet.Context, arg struct {
	Prompt string `json:"prompt"`
	Id     string `json:"id"`
}) error {
	client := openaiCfg.GetClient()
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo0613,
			Functions: apiCfg.GetFunctions(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: arg.Prompt,
				},
			},
		},
	)
	if err == nil {
		var apiResp json.RawMessage
		var pName, mName string
		if debugMode {
			logrus.WithField("call", resp.Choices[0].Message.FunctionCall).Info("call")
		}
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
	}
	return err
}
