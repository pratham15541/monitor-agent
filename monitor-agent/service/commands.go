package service

import (
	"context"
	"encoding/json"
	"monitor-agent/config"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type CommandRequest struct {
	DeviceID  string `json:"deviceId"`
	CommandID string `json:"commandId"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
}

type CommandResult struct {
	DeviceID   string `json:"deviceId"`
	CommandID  string `json:"commandId"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Output     string `json:"output"`
	Error      string `json:"error"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

const (
	commandTimeout        = 30 * time.Second
	maxCommandOutputBytes = 16 * 1024
)

func StartCommandLoop(cfg *config.Config, stop <-chan struct{}) {
	go func() {
		for {
			if stop != nil {
				select {
				case <-stop:
					return
				default:
				}
			}

			if cfg.ServerURL == "" || cfg.Token == "" || cfg.DeviceID == "" {
				time.Sleep(5 * time.Second)
				continue
			}

			if err := runCommandSession(cfg, stop); err != nil {
				logrus.Error("Command websocket error:", err)
			}

			time.Sleep(3 * time.Second)
		}
	}()
}

func runCommandSession(cfg *config.Config, stop <-chan struct{}) error {
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

	subscriptionID := "agent-" + cfg.DeviceID
	if err := sendStompFrame(conn, stompFrame{
		Command: "SUBSCRIBE",
		Headers: map[string]string{
			"id":          subscriptionID,
			"destination": "/topic/agent/" + cfg.DeviceID,
		},
	}); err != nil {
		return err
	}

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		frames := parseStompFrames(string(payload))
		for _, frame := range frames {
			if frame.Command != "MESSAGE" {
				continue
			}

			var request CommandRequest
			if err := json.Unmarshal([]byte(frame.Body), &request); err != nil {
				continue
			}

			if request.DeviceID != "" && request.DeviceID != cfg.DeviceID {
				continue
			}

			result := executeCommand(cfg, request)
			body, err := json.Marshal(result)
			if err != nil {
				continue
			}

			_ = sendStompFrame(conn, stompFrame{
				Command: "SEND",
				Headers: map[string]string{
					"destination":  "/app/command-result",
					"content-type": "application/json",
				},
				Body: string(body),
			})
		}
	}
}

func executeCommand(cfg *config.Config, request CommandRequest) CommandResult {
	started := time.Now()
	result := CommandResult{
		DeviceID:  cfg.DeviceID,
		CommandID: request.CommandID,
		Type:      request.Type,
		Status:    "ok",
		StartedAt: started.Format(time.RFC3339),
	}

	switch request.Type {
	case "shell":
		output, errText, status := runShellCommand(request.Payload)
		result.Output = output
		result.Error = errText
		result.Status = status
	case "service":
		output, errText, status := runServiceAction(request.Payload)
		result.Output = output
		result.Error = errText
		result.Status = status
	case "diagnostics":
		result.Output = buildDiagnostics(cfg)
	default:
		result.Status = "error"
		result.Error = "unknown command type"
	}

	result.Output = trimOutput(result.Output)
	result.Error = trimOutput(result.Error)

	result.FinishedAt = time.Now().Format(time.RFC3339)
	return result
}

func trimOutput(value string) string {
	if len(value) <= maxCommandOutputBytes {
		return value
	}

	return value[len(value)-maxCommandOutputBytes:]
}

func runShellCommand(command string) (string, string, string) {
	if strings.TrimSpace(command) == "" {
		return "", "empty command", "error"
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), "command timed out", "timeout"
	}

	if err != nil {
		return string(output), err.Error(), "error"
	}

	return string(output), "", "ok"
}

func runServiceAction(action string) (string, string, string) {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		return "", "empty service action", "error"
	}

	switch action {
	case "start", "stop", "restart":
		if action == "restart" {
			if err := ControlService("stop"); err != nil {
				return "", err.Error(), "error"
			}
			if err := ControlService("start"); err != nil {
				return "", err.Error(), "error"
			}
			return "service restarted", "", "ok"
		}

		if err := ControlService(action); err != nil {
			return "", err.Error(), "error"
		}
		return "service action completed", "", "ok"
	default:
		return "", "unsupported service action", "error"
	}
}

func buildDiagnostics(cfg *config.Config) string {
	metrics := CollectMetrics()
	payload := map[string]interface{}{
		"deviceId":  cfg.DeviceID,
		"hostname":  getHostname(),
		"os":        getOS(),
		"goVersion": runtime.Version(),
		"metrics":   metrics,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}
