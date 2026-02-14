package service

import (
	"bytes"
	"encoding/json"
	"monitor-agent/config"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func StartDetailedMetricsLoop(cfg *config.Config, stop <-chan struct{}, interval time.Duration) {
	go func() {
		for {
			if stop != nil {
				select {
				case <-stop:
					return
				default:
				}
			}

			if cfg.ServerURL == "" || cfg.Token == "" {
				time.Sleep(5 * time.Second)
				continue
			}

			if cfg.DeviceID == "" {
				if err := RegisterIfNeeded(cfg); err != nil {
					logrus.Error("Registration failed:", err)
					time.Sleep(5 * time.Second)
					continue
				}
			}

			sendDetailedMetrics(cfg)

			select {
			case <-time.After(interval):
			case <-stop:
				return
			}
		}
	}()
}

func sendDetailedMetrics(cfg *config.Config) {
	details := CollectDetailedMetrics()
	payload := map[string]interface{}{
		"deviceId": cfg.DeviceID,
		"details":  details,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logrus.Error("Failed to marshal detailed metrics:", err)
		return
	}

	url := cfg.ServerURL + "/agent/metrics-detail"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		logrus.Error("Failed to create request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-agent-token", cfg.Token)

	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.Error("Failed to send detailed metrics:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Detailed metrics POST returned %d", resp.StatusCode)
		return
	}

	logrus.Debug("Detailed metrics sent successfully")
}
