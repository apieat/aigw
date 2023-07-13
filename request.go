package aigw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

type CompletionRequest struct {
	Prompt    string            `json:"prompt"`
	Id        string            `json:"id"`
	Type      string            `json:"type"`
	Functions []AllowedFunction `json:"functions"`
	Debug     bool              `json:"debug"`
}

type AllowedFunction struct {
	Path   string `json:"path"`
	Method string `json:"method"`
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

//Call calls the completion endpoint
func (c *CompletionRequest) Call(url string, funcs ...AllowedFunction) error {
	var bts bytes.Buffer
	c.Functions = []AllowedFunction{
		{
			Path:   "/api/project/fill",
			Method: "POST",
		},
	}
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
