package config

import (
	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	Token          string            `yaml:"token"`
	Instructions   map[string]string `yaml:"instructions"`
	Templates      map[string]string `yaml:"templates"`
	client         *openai.Client
	Sync           bool `yaml:"sync"`
	MaxTokens      int  `yaml:"max_tokens"`
	AskAiToAnalyse bool `yaml:"ask_ai_to_analyse_result"`
}

func (o *OpenAIConfig) GetClient() *openai.Client {
	if o.client == nil {
		o.client = openai.NewClient(o.Token)
	}
	return o.client
}
