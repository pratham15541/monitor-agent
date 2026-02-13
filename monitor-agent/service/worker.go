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

	for {

		select {
		case <-stop:
			logrus.Info("Worker stopped")
			return

		default:

			if cfg.Token == "" {
				logrus.Warn("Token not set")
				time.Sleep(10 * time.Second)
				continue
			}

			err := RegisterIfNeeded(cfg)
			if err != nil {
				logrus.Error("Registration failed:", err)
				time.Sleep(5 * time.Second)
				continue
			}

			metric := CollectMetrics()
			metric["deviceId"] = cfg.DeviceID

			_, err = postJSON(cfg.ServerURL+"/agent/metrics", metric)
			if err != nil {
				logrus.Error("Metric send failed:", err)
			}

			time.Sleep(10 * time.Second)
		}
	}
}
