package main

import (
	"github.com/apieat/aigw/ctrl"
	"github.com/extrame/goblet"
	"github.com/extrame/goblet/plugin"
	"github.com/pkg/errors"
)

func main() {
	server := goblet.Organize("aigw", plugin.JSON)

	err := ctrl.AddConfig(server)

	if err != nil {
		panic(errors.Wrap(err, "add config failed"))
	}

	server.ControlBy(&ctrl.Completion{})

	server.Run()
}
