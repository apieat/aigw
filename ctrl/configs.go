package ctrl

import (
	"github.com/apieat/aigw/config"
	"github.com/extrame/goblet"
	"github.com/sirupsen/logrus"
)

// var openaiCfg platform.AIConfig
// var apiCfg config.ApiConfig

func AddConfig(server *goblet.Server) (*config.Config, error) {

	var config config.Config

	err := server.AddConfig("ai", &config.Platform)
	if err == nil {

		if config.Platform.Platform == "" {
			config.Platform.Platform = "chatgpt"
		}

		config.Platform.Init()
		// platform.Init(openaiCfg.Platform)

		err = server.AddConfig("api", &config.Api)
		if err == nil {
			logrus.WithField("templates", config.Platform.Templates).Infoln("openai config loaded")
			err = config.Api.Init()
			if err == nil {
				// openaiCfg = config.Api
				// apiCfg = config.Api
				return &config, nil
			}
		}
	}

	return nil, err
}
