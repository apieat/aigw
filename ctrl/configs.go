package ctrl

import (
	"github.com/apieat/aigw/config"
	"github.com/extrame/goblet"
)

var openaiCfg config.OpenAIConfig
var apiCfg config.ApiConfig
var debugMode bool

func AddConfig(server *goblet.Server) error {
	server.AddConfig("openai", &openaiCfg)
	server.AddConfig("api", &apiCfg)

	server.Debug(func() {
		debugMode = true
	})

	return apiCfg.Init()

}
