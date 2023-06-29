package ctrl

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type CompletionRequest struct {
	Prompt string `json:"prompt"`
	Id     string `json:"id"`
	Type   string `json:"type"`
}

//ToPrompt returns the prompt. If templates are provided, it will use the first template that matches the type to wrap the prompt.
func (c *CompletionRequest) ToPrompt(templates ...map[string]string) string {
	if len(templates) > 0 && templates[0] != nil {
		if c.Type == "" {
			c.Type = "default"
		}
		temp, ok := templates[0][c.Type]
		if ok {
			return fmt.Sprintf(temp, c.Prompt)
		} else {
			logrus.Errorln("template not found for type", c.Type)
		}
	}
	return c.Prompt
}
