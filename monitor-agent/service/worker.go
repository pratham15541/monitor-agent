package service

import (
	"monitor-agent/config"
	"time"

	"github.com/sirupsen/logrus"
)

func StartWorker(stop chan struct{}) {
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Loading config")
	if cfg.Token == "" {
		logrus.Fatal("Token not set. Use 'monitor-agent install' or 'set-token'")
	}

	logrus.Info("Starting command and metrics loops...")
	go StartCommandLoop(cfg, stop)
	go StartMetricsWebSocketLoop(cfg, stop, 2*time.Second)
	go StartDetailedMetricsLoop(cfg, stop, 30*time.Second)

	// Retry registration in background (don't block service startup)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}

			if cfg.DeviceID == "" {
				logrus.Info("Attempting to register device...")
				if err := RegisterIfNeeded(cfg); err != nil {
					logrus.Warn("Registration failed, will retry:", err)
					time.Sleep(10 * time.Second)
					continue
				}
				logrus.Info("Device registered:", cfg.DeviceID)
			}

			time.Sleep(15 * time.Second)
		}
	}()

	logrus.Info("Agent startup complete")
}
