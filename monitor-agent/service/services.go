package service

import (
	"runtime"
)

const servicesMaxBytes = 32 * 1024

func collectServicesSnapshot() map[string]interface{} {
	output, source := collectServiceOutput()
	return map[string]interface{}{
		"source": source,
		"output": output,
	}
}

func collectServiceOutput() (string, string) {
	switch runtime.GOOS {
	case "windows":
		return runCommandLimited(servicesMaxBytes, "sc", "query", "state=", "all"), "sc"
	case "linux":
		return runCommandLimited(servicesMaxBytes, "systemctl", "list-units", "--type=service", "--all", "--no-pager"), "systemctl"
	case "darwin":
		return runCommandLimited(servicesMaxBytes, "launchctl", "list"), "launchctl"
	default:
		return "", ""
	}
}
