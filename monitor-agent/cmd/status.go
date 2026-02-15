package cmd

import (
	"fmt"
	"monitor-agent/config"
	"monitor-agent/service"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show agent and service status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Failed to load config:", err)
			return
		}

		fmt.Println("Config path:", config.ConfigPath())
		fmt.Println("Log path:", config.GetLogPath())
		fmt.Println("Server:", cfg.ServerURL)
		fmt.Println("Token set:", cfg.Token != "")
		fmt.Println("Device ID:", cfg.DeviceID)

		status, err := service.GetServiceStatus()
		if err != nil {
			fmt.Println("Service status:", status, "(", err, ")")
			return
		}

		fmt.Println("Service status:", status)
	},
}
