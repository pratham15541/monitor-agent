package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"monitor-agent/config"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	detailBatchSize    = 1
	detailBatchMaxWait = 30 * time.Second
)

func StartDetailedMetricsLoop(cfg *config.Config, stop <-chan struct{}, interval time.Duration) {
	go func() {
		var batch []map[string]interface{}
		lastFlush := time.Now()

		for {
			if stop != nil {
				select {
				case <-stop:
					_ = sendDetailedMetricsBatch(cfg, batch)
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

			payload := collectDetailedMetricsPayload(cfg)
			if payload != nil {
				batch = append(batch, payload)
			}

			if len(batch) >= detailBatchSize || time.Since(lastFlush) >= detailBatchMaxWait {
				if err := sendDetailedMetricsBatch(cfg, batch); err != nil {
					logrus.Error("Failed to send detailed metrics batch:", err)
				} else {
					batch = batch[:0]
					lastFlush = time.Now()
				}
			}

			select {
			case <-time.After(interval):
			case <-stop:
				_ = sendDetailedMetricsBatch(cfg, batch)
				return
			}
		}
	}()
}

func collectDetailedMetricsPayload(cfg *config.Config) map[string]interface{} {
	details := CollectDetailedMetrics()
	return map[string]interface{}{
		"deviceId": cfg.DeviceID,
		"details":  details,
	}
}

func sendDetailedMetricsBatch(cfg *config.Config, batch []map[string]interface{}) error {
	if len(batch) == 0 {
		return nil
	}

	body, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	url := cfg.ServerURL + "/agent/metrics-detail/batch"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-agent-token", cfg.Token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("detailed metrics batch POST returned %d", resp.StatusCode)
	}

	logrus.Debug("Detailed metrics batch sent successfully")
	return nil
}
