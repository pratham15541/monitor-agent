package cmd

import (
	"fmt"
	"monitor-agent/config"
	"monitor-agent/service"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall agent service",
	Run: func(cmd *cobra.Command, args []string) {
		removeService, _ := cmd.Flags().GetBool("service")

		if removeService {
			fmt.Println("Stopping service...")
			if err := service.ControlService("stop"); err != nil {
				fmt.Println("Service stop failed:", err)
			}

			fmt.Println("Uninstalling service...")
			if err := service.ControlService("uninstall"); err != nil {
				fmt.Println("Service uninstall failed:", err)
			}
		}

		if err := config.Delete(); err != nil {
			fmt.Println("Failed to delete config:", err)
			return
		}

		fmt.Println("Uninstall complete.")
	},
}

func init() {
	uninstallCmd.Flags().Bool("service", false, "Stop and remove the service")
}
