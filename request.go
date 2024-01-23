package aigw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/apieat/aigw/platform"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type CompletionRequest struct {
	Instruction string `json:"instruction"`
	Prompt      string `json:"prompt"`
	Id          string `json:"id"`
	//action mode for the request, default is "json" supported: json, text, function_call
	Mode string `json:"mode"`
	//Type for template and instructions
	Type        string            `json:"type"`
	Functions   []AllowedFunction `json:"functions"`
	Debug       bool              `json:"debug"`
	Temparature float32           `json:"temparature"`
}

type AllowedFunction struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func (c *CompletionRequest) ToMessages(instructions, templates map[string]string) []openai.ChatCompletionMessage {
	return platform.Current.ToMessages(c, instructions, templates)
}

func (c *CompletionRequest) GetInstruction() string {
	return c.Instruction
}

// GetPrompt returns the prompt
func (c *CompletionRequest) GetPrompt() string {
	return c.Prompt
}

// ToPrompt returns the prompt. If templates are provided, it will use the first template that matches the type to wrap the prompt.
func (c *CompletionRequest) ToPrompt(prompt string, templates ...map[string]string) string {
	if len(templates) > 0 && templates[0] != nil {
		if c.Type == "" {
			c.Type = "default"
		}
		temp, ok := templates[0][c.Type]
		if ok {
			return fmt.Sprintf(temp, prompt)
		} else {
			logrus.Errorln("template not found for type", c.Type)
		}
	}
	return prompt
}

// Call calls the completion endpoint
func (c *CompletionRequest) Call(url string) error {
	var bts bytes.Buffer
	err := json.NewEncoder(&bts).Encode(&c)
	if err == nil {
		var req *http.Request
		req, err = http.NewRequest(http.MethodPost, url, &bts)
		req.Header.Set("Content-Type", "application/json")
		if err == nil {
			var resp *http.Response
			resp, err = http.DefaultClient.Do(req)
			if err == nil {
				var bts []byte
				bts, err = io.ReadAll(resp.Body)
				var r Response
				if err == nil {
					err = json.Unmarshal(bts, &r)
					if err == nil && r.Success {
						return err
					} else if err == nil {
						return &r
					}
				}
			}
		}
	}
	return err
}

func (c *CompletionRequest) GetTemparature() float32 {
	if c.Temparature == 0 {
		return 1
	}
	return c.Temparature
}

type Response struct {
	Err     string `json:"error"`
	Success bool   `json:"success"`
}

func (r *Response) Error() string {
	if r.Success {
		return ""
	}
	return r.Err
}
