package cmd

import (
	"fmt"
	"monitor-agent/service"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop agent service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping service...")
		if err := service.ControlService("stop"); err != nil {
			fmt.Println("Service stop failed:", err)
			return
		}

		fmt.Println("Service stopped.")
	},
}
