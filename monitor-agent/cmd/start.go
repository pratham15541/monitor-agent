package cmd

import (
	"fmt"
	"monitor-agent/service"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start agent service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting service...")
		if err := service.ControlService("start"); err != nil {
			fmt.Println("Service start failed:", err)
			return
		}

		fmt.Println("Service started.")
	},
}
