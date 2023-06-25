package config

import (
	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	Token  string `yaml:"token"`
	client *openai.Client
}

func (this *OpenAIConfig) GetClient() *openai.Client {
	if this.client == nil {
		this.client = openai.NewClient(this.Token)
	}
	return this.client
}
