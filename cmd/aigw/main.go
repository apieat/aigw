package main

import (
	"github.com/apieat/aigw/ctrl"
	"github.com/extrame/goblet"
	"github.com/extrame/goblet/plugin"
)

func main() {
	server := goblet.Organize("aigw", plugin.JSON)

	err := ctrl.AddConfig(server)

	if err != nil {
		panic(err)
	}

	server.ControlBy(&ctrl.Completion{})

	server.Run()
}
