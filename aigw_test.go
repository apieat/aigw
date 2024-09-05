package aigw

import (
	"testing"

	"github.com/apieat/aigw/config"
	"github.com/apieat/aigw/model"
	"github.com/apieat/aigw/stream"
	"github.com/sirupsen/logrus"

	_ "github.com/apieat/aigw/platform/aistudio"
)

func TestMain(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	cfg, err := config.Init("./config.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(cfg)
	var testRequest = &CompletionRequest{
		Prompt: "一款宠物医院管理系统，可以管理宠物的基本信息，疫苗接种情况，病历信息，以及宠物的体检信息。可以为宠物主人提供宠物的健康管理服务。",
		Type:   "generate_software",
		Functions: []model.AllowedFunction{
			{
				Path:   "/api/project/fill",
				Method: "POST",
			},
		},
		Stream: true,
	}
	_, err = testRequest.SendStream(cfg, func(s *stream.Stat) {
		t.Log(string(s.Json))
	})
	if err != nil {
		t.Error(err)
	}

}
