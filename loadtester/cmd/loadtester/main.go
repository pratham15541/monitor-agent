package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"loadtester/internal/config"
	"loadtester/internal/load"
)

func main() {
	var configPath string
	var preflightOnly bool
	var skipPreflight bool
	flag.StringVar(&configPath, "config", "", "Path to load tester YAML config")
	flag.BoolVar(&preflightOnly, "preflight-only", false, "Run preflight checks only and exit")
	flag.BoolVar(&skipPreflight, "skip-preflight", false, "Skip preflight checks before the load run")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config:", err)
		os.Exit(1)
	}

	runner, err := load.NewRunner(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create runner:", err)
		os.Exit(1)
	}

	if !skipPreflight {
		preflightReport, err := runner.Preflight(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "preflight failed:", err)
			os.Exit(1)
		}
		load.PrintPreflightReport(preflightReport)
		if preflightOnly {
			return
		}
	}

	if err := runner.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "load test failed:", err)
		os.Exit(1)
	}
}
