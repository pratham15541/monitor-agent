package service

import (
	"context"
	"encoding/json"
	"monitor-agent/config"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	appservice "github.com/kardianos/service"
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
	Chunked    bool   `json:"chunked"`
	ChunkType  string `json:"chunkType"`
	ChunkIndex int    `json:"chunkIndex"`
	ChunkCount int    `json:"chunkCount"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

const (
	commandTimeout       = 30 * time.Second
	maxCommandChunkBytes = 12 * 1024
)

var blockedCommandPattern = regexp.MustCompile(
	`(?i)\b(rm|remove-item|removeitem|del|erase|rmdir|rd|format)\b`,
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

	logrus.WithFields(logrus.Fields{
		"deviceId": cfg.DeviceID,
		"wsUrl":    wsURL,
	}).Info("Command websocket connected")

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

	logrus.WithFields(logrus.Fields{
		"deviceId": cfg.DeviceID,
		"topic":    "/topic/agent/" + cfg.DeviceID,
	}).Info("Command websocket subscribed")

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

			results := executeCommand(cfg, request)
			for _, result := range results {
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
}

func executeCommand(cfg *config.Config, request CommandRequest) []CommandResult {
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
		logCommandResult(request, status, errText, output)
	case "service":
		output, errText, status := runServiceAction(request.Payload)
		result.Output = output
		result.Error = errText
		result.Status = status
		logCommandResult(request, status, errText, output)
	case "diagnostics":
		result.Output = buildDiagnostics(cfg)
	case "collect-details":
		if err := sendDetailedMetricsNow(cfg); err != nil {
			result.Status = "error"
			result.Error = err.Error()
		} else {
			result.Output = "detailed metrics collected"
		}
	default:
		result.Status = "error"
		result.Error = "unknown command type"
	}

	finishedAt := time.Now().Format(time.RFC3339)
	return chunkCommandResult(result, finishedAt)
}

func runShellCommand(command string) (string, string, string) {
	if strings.TrimSpace(command) == "" {
		return "", "empty command", "error"
	}

	if blockedCommandPattern.MatchString(command) {
		return "", "blocked command detected", "error"
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Prefer PowerShell to support ls/pwd/dir aliases in service mode.
		// Force UTF-8 output to avoid null-padded strings in service context.
		psCommand :=
			"$OutputEncoding=[System.Text.UTF8Encoding]::new();" +
				"[Console]::OutputEncoding=[System.Text.UTF8Encoding]::new();" +
				command
		cmd = exec.CommandContext(ctx, "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
			"-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCommand)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	cleanOutput := sanitizeOutput(string(output))
	if ctx.Err() == context.DeadlineExceeded {
		return cleanOutput, "command timed out", "timeout"
	}

	if err != nil {
		return cleanOutput, err.Error(), "error"
	}

	return cleanOutput, "", "ok"
}

func sanitizeOutput(value string) string {
	return strings.ReplaceAll(value, "\x00", "")
}

func chunkCommandResult(result CommandResult, finishedAt string) []CommandResult {
	outputChunks := splitChunks(result.Output, maxCommandChunkBytes)
	errorChunks := splitChunks(result.Error, maxCommandChunkBytes)

	if len(outputChunks) <= 1 && len(errorChunks) <= 1 {
		result.FinishedAt = finishedAt
		return []CommandResult{result}
	}

	results := make([]CommandResult, 0, len(outputChunks)+len(errorChunks))
	if len(outputChunks) > 0 {
		for i, chunk := range outputChunks {
			item := result
			item.Output = chunk
			item.Error = ""
			item.Chunked = true
			item.ChunkType = "output"
			item.ChunkIndex = i
			item.ChunkCount = len(outputChunks)
			if i == len(outputChunks)-1 && len(errorChunks) == 0 {
				item.FinishedAt = finishedAt
				item.Status = result.Status
			} else {
				item.Status = "stream"
			}
			results = append(results, item)
		}
	}

	if len(errorChunks) > 0 {
		for i, chunk := range errorChunks {
			item := result
			item.Output = ""
			item.Error = chunk
			item.Chunked = true
			item.ChunkType = "error"
			item.ChunkIndex = i
			item.ChunkCount = len(errorChunks)
			if i == len(errorChunks)-1 {
				item.FinishedAt = finishedAt
				item.Status = result.Status
			} else {
				item.Status = "stream"
			}
			results = append(results, item)
		}
	}

	return results
}

func splitChunks(value string, size int) []string {
	if value == "" {
		return nil
	}
	if size <= 0 || len(value) <= size {
		return []string{value}
	}

	chunks := make([]string, 0, (len(value)+size-1)/size)
	for start := 0; start < len(value); start += size {
		end := start + size
		if end > len(value) {
			end = len(value)
		}
		chunks = append(chunks, value[start:end])
	}
	return chunks
}

func runServiceAction(action string) (string, string, string) {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		return "", "empty service action", "error"
	}

	if !appservice.Interactive() {
		switch action {
		case "start":
			return "service already running", "", "ok"
		case "stop", "restart":
			scheduleServiceAction(action)
			return "service action scheduled", "", "ok"
		}
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

func scheduleServiceAction(action string) {
	go func() {
		time.Sleep(800 * time.Millisecond)
		if err := ControlService(action); err != nil {
			logrus.Warn("Scheduled service action failed:", err)
		}
	}()
}

func logCommandResult(request CommandRequest, status, errText, output string) {
	snippet := trimLogSnippet(output)
	if snippet == "" {
		snippet = trimLogSnippet(errText)
	}

	logrus.WithFields(logrus.Fields{
		"commandType": request.Type,
		"commandId":   request.CommandID,
		"status":      status,
		"payload":     trimLogSnippet(request.Payload),
		"snippet":     snippet,
	}).Info("Remote command executed")
}

func trimLogSnippet(value string) string {
	const maxLen = 400
	if value == "" {
		return ""
	}
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen]
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

func sendDetailedMetricsNow(cfg *config.Config) error {
	payload := collectDetailedMetricsPayload(cfg)
	if payload == nil {
		return nil
	}

	return sendDetailedMetricsBatch(cfg, []map[string]interface{}{payload})
}
