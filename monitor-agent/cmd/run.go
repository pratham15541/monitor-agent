package cmd

import (
	"fmt"
	"monitor-agent/config"
	"monitor-agent/service"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run agent in foreground",
	Run: func(cmd *cobra.Command, args []string) {
		runAgent()
	},
}

func runAgent() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	fmt.Println("Starting agent...")
	fmt.Println("Server:", cfg.ServerURL)
	fmt.Println("Token:", cfg.Token)

	if err := service.RegisterIfNeeded(cfg); err != nil {
		fmt.Println("Register failed:", err)
		return
	}

	service.StartMetricsLoop(cfg)
}
