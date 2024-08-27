package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type AIConfig struct {
	Token          string            `yaml:"token"`
	Instructions   map[string]string `yaml:"instructions"`
	Templates      map[string]string `yaml:"templates"`
	client         *openai.Client
	Sync           bool              `yaml:"sync"`
	MaxTokens      int               `yaml:"max_tokens"`
	AskAiToAnalyse bool              `yaml:"ask_ai_to_analyse_result"`
	Url            map[string]string `yaml:"url"`
	Oauth2         *Oauth2Client     `yaml:"oauth2"`
	Platform       string            `yaml:"platform"`
	platform       Platform
}

func (c *AIConfig) Init() error {
	c.platform = platforms[c.Platform]
	if c.platform == nil {
		return fmt.Errorf("platform %s not supported", c.Platform)
	}
	return c.platform.Init(c)
}

func (c *AIConfig) GetModel(typ string) string {
	return c.platform.GetModel(typ)
}

func (c *AIConfig) AddFunctionsToMessage(functions []openai.FunctionDefinition, fc *openai.FunctionCall, req *openai.ChatCompletionRequest) *openai.ChatCompletionRequest {
	return c.platform.AddFunctionsToMessage(functions, fc, req)
}

func (c *AIConfig) CreateChatCompletion(req *openai.ChatCompletionRequest, typ string) (ChatCompletionResponse, error) {
	return c.platform.CreateChatCompletion(req, typ)
}

func (c *AIConfig) AddResponseToMessage(messages []openai.ChatCompletionMessage, resp ChatCompletionResponse) []openai.ChatCompletionMessage {
	return c.platform.AddResponseToMessage(messages, resp)
}

func (c *AIConfig) CreateChatStream(req *openai.ChatCompletionRequest, typ string, fn func(string)) error {
	return c.platform.CreateChatStream(req, typ, fn)
}

func (c *AIConfig) ToMessages(cr CompletionRequest) []openai.ChatCompletionMessage {
	return c.platform.ToMessages(cr, c.Instructions, c.Templates)
}

func (o *AIConfig) GetToken() string {
	if o.Token != "" {
		return o.Token
	}
	return o.Oauth2.GetToken()
}

type Oauth2Client struct {
	ClientId     string `yaml:"test" goblet:"client_id"`
	ClientSecret string `yaml:"secret" goblet:"client_secret"`
	Url          string `yaml:"url"`
	expireTime   time.Time
	token        string
	_url         *url.URL
}

type Oauth2Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (o *Oauth2Client) GetToken() string {
	if o.token == "" || o.expireTime.Before(time.Now()) {
		o.refreshToken()
	}
	return o.token
}

func (o *Oauth2Client) refreshToken() {
	if o._url == nil {
		o._url, _ = url.Parse(o.Url)
		var params = url.Values{}
		params.Add("grant_type", "client_credentials")
		params.Add("client_id", o.ClientId)
		params.Add("client_secret", o.ClientSecret)
		o._url.RawQuery = params.Encode()
		logrus.WithField("config", o).Infoln("refresh token url", o._url.String())
	}
	var resp, err = http.Get(o._url.String())
	if err == nil {
		var token Oauth2Token
		var bts []byte
		bts, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil {
			err = json.NewDecoder(bytes.NewBuffer(bts)).Decode(&token)
			if err == nil {
				logrus.Info("refresh token success", token)
				o.token = token.AccessToken
				o.expireTime = time.Now().Add(time.Duration(token.ExpiresIn-60) * time.Second)
			} else {
				logrus.WithField("resp", string(bts)).WithError(err).Error("refresh token failed")
			}
		}

	} else {
		logrus.WithError(err).Error("send refresh token req failed")
	}
	if err == nil {
		logrus.WithField("token", o.token).WithField("expireTime", o.expireTime).Info("refresh token success")
	}
}
