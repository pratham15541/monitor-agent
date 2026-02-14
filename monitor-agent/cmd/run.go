package cmd

import (
	"fmt"
	"monitor-agent/config"
	"monitor-agent/service"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	stop := make(chan struct{})
	service.StartCommandLoop(cfg, stop)
	service.StartMetricsWebSocketLoop(cfg, stop, 5*time.Second)
	service.StartDetailedMetricsLoop(cfg, stop, 30*time.Second)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nShutting down...")
	close(stop)
	time.Sleep(1 * time.Second)
	fmt.Println("Agent stopped.")
}
