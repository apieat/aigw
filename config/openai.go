package config

import (
	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	Token     string            `yaml:"token"`
	Templates map[string]string `yaml:"templates"`
	client    *openai.Client
	Sync      bool `yaml:"sync"`
	MaxTokens int  `yaml:"max_tokens"`
}

func (o *OpenAIConfig) GetClient() *openai.Client {
	if o.client == nil {
		o.client = openai.NewClient(o.Token)
	}
	return o.client
}
