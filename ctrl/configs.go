package ctrl

import (
	"github.com/apieat/aigw/config"
	"github.com/apieat/aigw/platform"
	"github.com/extrame/goblet"
	"github.com/sirupsen/logrus"
)

var openaiCfg platform.AIConfig
var apiCfg config.ApiConfig

func AddConfig(server *goblet.Server) error {
	err := server.AddConfig("ai", &openaiCfg)
	if err == nil {

		if openaiCfg.Platform == "" {
			openaiCfg.Platform = "chatgpt"
		}

		platform.Init(openaiCfg.Platform)

		err = server.AddConfig("api", &apiCfg)
		if err == nil {
			logrus.WithField("templates", openaiCfg.Templates).Infoln("openai config loaded")
			return apiCfg.Init()
		}
	}

	return err
}
