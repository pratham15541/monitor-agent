package cmd

import (
	"fmt"
	"monitor-agent/config"
	"monitor-agent/service"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install agent service",
	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		server, _ := cmd.Flags().GetString("server")

		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Failed to load config:", err)
			return
		}

		if token == "" {
			fmt.Println("Token is required. Use --token.")
			return
		}

		cfg.Token = token
		if server != "" {
			cfg.ServerURL = server
		}
		cfg.DeviceID = ""

		if err := config.Save(cfg); err != nil {
			fmt.Println("Failed to save config:", err)
			return
		}

		fmt.Println("Installing service...")
		if err := service.ControlService("install"); err != nil {
			fmt.Println("Service install failed:", err)
			return
		}

		fmt.Println("Service installed. Use 'monitor-agent start' to run it.")
	},
}

func init() {
	installCmd.Flags().String("token", "", "Company API token")
	installCmd.Flags().String("server", "", "Backend server URL")
}
