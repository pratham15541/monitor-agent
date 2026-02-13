package cmd

import (
	"fmt"
	"monitor-agent/config"

	"github.com/spf13/cobra"
)

var deregisterCmd = &cobra.Command{
	Use:   "deregister",
	Short: "Clear the stored device ID",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Failed to load config:", err)
			return
		}

		cfg.DeviceID = ""
		if err := config.Save(cfg); err != nil {
			fmt.Println("Failed to save config:", err)
			return
		}

		fmt.Println("Device ID cleared.")
	},
}
