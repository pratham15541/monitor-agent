package cmd

import (
	"fmt"
	"monitor-agent/config"

	"github.com/spf13/cobra"
)

var setURLCmd = &cobra.Command{
	Use:   "set-url [URL]",
	Short: "Set backend server URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverURL := args[0]

		cfg, _ := config.Load()
		cfg.ServerURL = serverURL
		cfg.DeviceID = "" // force re-register

		if err := config.Save(cfg); err != nil {
			fmt.Println("Failed to save config:", err)
			return
		}

		fmt.Println("Server URL updated successfully.")
	},
}
