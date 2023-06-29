package ctrl

import (
	"github.com/apieat/aigw/config"
	"github.com/extrame/goblet"
	"github.com/sirupsen/logrus"
)

var openaiCfg config.OpenAIConfig
var apiCfg config.ApiConfig

func AddConfig(server *goblet.Server) error {
	err := server.AddConfig("openai", &openaiCfg)
	if err == nil {
		err = server.AddConfig("api", &apiCfg)
		if err == nil {
			logrus.WithField("templates", openaiCfg.Templates).Infoln("openai config loaded")
			return apiCfg.Init()
		}
	}

	return err
}
