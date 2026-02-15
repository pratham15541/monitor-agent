package main

import (
	"monitor-agent/cmd"
	"monitor-agent/service"

	appservice "github.com/kardianos/service"
)

func main() {
	service.InitLogger()
	if !appservice.Interactive() {
		service.RunService()
		return
	}
	cmd.Execute()
}
