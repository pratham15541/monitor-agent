package main

import (
	"monitor-agent/cmd"
	"monitor-agent/service"
)

func main() {
	service.InitLogger()
	cmd.Execute()
}
