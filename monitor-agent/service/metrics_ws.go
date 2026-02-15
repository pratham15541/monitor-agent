package service

import (
	"encoding/json"
	"monitor-agent/config"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	metricsBatchSize    = 10
	metricsBatchMaxWait = 5 * time.Second
)

func StartMetricsWebSocketLoop(cfg *config.Config, stop <-chan struct{}) {
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

			if err := runMetricsSession(cfg, stop); err != nil {
				logrus.Error("Metrics websocket error:", err)
			}

			time.Sleep(3 * time.Second)
		}
	}()
}

func runMetricsSession(cfg *config.Config, stop <-chan struct{}) error {
	wsURL := toWebSocketURL(cfg.ServerURL)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetReadLimit(4 * 1024 * 1024)

	if stop != nil {
		go func() {
			<-stop
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			_ = conn.Close()
		}()
	}

	if err := sendStompFrame(conn, stompFrame{
		Command: "CONNECT",
		Headers: map[string]string{
			"accept-version": "1.2",
			"host":           "monitor-agent",
			"x-agent-token":  cfg.Token,
		},
	}); err != nil {
		return err
	}

	if err := waitForConnected(conn); err != nil {
		return err
	}

	var batch []map[string]interface{}
	lastFlush := time.Now()

	for {
		if stop != nil {
			select {
			case <-stop:
				return nil
			default:
			}
		}

		payload := CollectMetrics()
		payload["deviceId"] = cfg.DeviceID
		batch = append(batch, payload)

		if len(batch) >= metricsBatchSize || time.Since(lastFlush) >= metricsBatchMaxWait {
			if err := sendMetricsBatch(conn, batch); err != nil {
				return err
			}
			batch = batch[:0]
			lastFlush = time.Now()
		}

		interval := intervalForCpu(payload["cpuUsage"])
		timer := time.NewTimer(interval)
		select {
		case <-timer.C:
		case <-stop:
			timer.Stop()
			return nil
		}
	}
}

func intervalForCpu(cpuValue interface{}) time.Duration {
	cpuUsage, ok := cpuValue.(float64)
	if !ok {
		return 5 * time.Second
	}

	if cpuUsage > 90 {
		return 1 * time.Second
	}
	if cpuUsage > 70 {
		return 2 * time.Second
	}
	return 5 * time.Second
}

func sendMetricsBatch(conn *websocket.Conn, batch []map[string]interface{}) error {
	if len(batch) == 0 {
		return nil
	}

	body, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	return sendStompFrame(conn, stompFrame{
		Command: "SEND",
		Headers: map[string]string{
			"destination":  "/app/agent/metrics-batch",
			"content-type": "application/json",
		},
		Body: string(body),
	})
}
