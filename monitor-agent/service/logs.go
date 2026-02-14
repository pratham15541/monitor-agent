package service

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"monitor-agent/config"
)

const logMaxBytes = 16 * 1024

func collectLogsSnapshot() map[string]interface{} {
	return map[string]interface{}{
		"agent":  readAgentLogTail(logMaxBytes),
		"system": readSystemLogs(logMaxBytes),
	}
}

func readAgentLogTail(maxBytes int64) string {
	path := config.GetLogPath()
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return ""
	}

	size := stat.Size()
	if size <= 0 {
		return ""
	}

	start := size - maxBytes
	if start < 0 {
		start = 0
	}

	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return ""
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return ""
	}

	return string(data)
}

func readSystemLogs(maxBytes int64) string {
	switch runtime.GOOS {
	case "windows":
		system := runCommandLimited(maxBytes, "wevtutil", "qe", "System", "/c:50", "/f:text", "/rd:true")
		app := runCommandLimited(maxBytes, "wevtutil", "qe", "Application", "/c:50", "/f:text", "/rd:true")
		return system + "\n" + app
	case "linux":
		return runCommandLimited(maxBytes, "journalctl", "-n", "200", "--no-pager")
	case "darwin":
		return runCommandLimited(maxBytes, "log", "show", "--last", "10m", "--style", "compact")
	default:
		return ""
	}
}

func runCommandLimited(maxBytes int64, name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	_ = cmd.Run()

	data := output.Bytes()
	if int64(len(data)) <= maxBytes {
		return string(data)
	}

	return string(data[len(data)-int(maxBytes):])
}
