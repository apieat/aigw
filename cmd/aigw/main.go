package main

import (
	"github.com/apieat/aigw/ctrl"
	"github.com/extrame/goblet"
	"github.com/extrame/goblet/plugin"
	"github.com/pkg/errors"

	_ "github.com/apieat/aigw/platform/aistudio"
	_ "github.com/apieat/aigw/platform/chatgpt"
	_ "github.com/apieat/aigw/platform/qianfan"
	_ "github.com/apieat/aigw/platform/zhipu"
)

func main() {
	server := goblet.Organize("aigw", plugin.JSON)

	config, err := ctrl.AddConfig(server)

	if err != nil {
		panic(errors.Wrap(err, "add config failed"))
	}

	server.ControlBy(&ctrl.Completion{
		Config: config,
	})

	server.Run()
}
