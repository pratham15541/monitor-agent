package cmd

import (
	"fmt"
	"monitor-agent/config"

	"github.com/spf13/cobra"
)

var setTokenCmd = &cobra.Command{
	Use:   "set-token [TOKEN]",
	Short: "Set company API token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		token := args[0]

		cfg, _ := config.Load()
		cfg.Token = token
		cfg.DeviceID = "" // force re-register

		if err := config.Save(cfg); err != nil {
			fmt.Println("Failed to save config:", err)
			return
		}

		fmt.Println("Token updated successfully.")
	},
}
