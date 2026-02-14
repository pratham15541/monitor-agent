package service

import (
	"encoding/json"
	"monitor-agent/config"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func StartMetricsWebSocketLoop(cfg *config.Config, stop <-chan struct{}, interval time.Duration) {
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

			if err := runMetricsSession(cfg, stop, interval); err != nil {
				logrus.Error("Metrics websocket error:", err)
			}

			time.Sleep(3 * time.Second)
		}
	}()
}

func runMetricsSession(cfg *config.Config, stop <-chan struct{}, interval time.Duration) error {
	wsURL := toWebSocketURL(cfg.ServerURL)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

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

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			payload := CollectMetrics()
			payload["deviceId"] = cfg.DeviceID

			body, err := json.Marshal(payload)
			if err != nil {
				continue
			}

			if err := sendStompFrame(conn, stompFrame{
				Command: "SEND",
				Headers: map[string]string{
					"destination":  "/app/agent/metrics",
					"content-type": "application/json",
				},
				Body: string(body),
			}); err != nil {
				return err
			}
		case <-stop:
			return nil
		}
	}
}
