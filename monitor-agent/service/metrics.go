package service

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"fmt"
	"monitor-agent/config"
	"time"
	"runtime"
)

func StartMetricsLoop(cfg *config.Config) {

	fmt.Println("Starting metrics loop...")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := SendMetrics(cfg)
		if err != nil {
			fmt.Println("Metric send error:", err)
		}
	}
}


func CollectMetrics() map[string]interface{} {

	var cpuUsage float64 = 0
	var memoryUsage float64 = 0
	var diskUsage float64 = 0
	var netIn uint64 = 0
	var netOut uint64 = 0

	// CPU
	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// Memory
	if memStats, err := mem.VirtualMemory(); err == nil {
		memoryUsage = memStats.UsedPercent
	}

	// Disk
	var diskPath string
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	} else {
		diskPath = "/"
	}

	if diskStats, err := disk.Usage(diskPath); err == nil {
		diskUsage = diskStats.UsedPercent
	}

	// Network
	if netStats, err := net.IOCounters(false); err == nil && len(netStats) > 0 {
		netIn = netStats[0].BytesRecv
		netOut = netStats[0].BytesSent
	}

	return map[string]interface{}{
		"cpuUsage":    cpuUsage,
		"memoryUsage": memoryUsage,
		"diskUsage":   diskUsage,
		"networkIn":   netIn,
		"networkOut":  netOut,
	}
}

func SendMetrics(cfg *config.Config) error {

	metrics := CollectMetrics()

	payload := map[string]interface{}{
		"deviceId":    cfg.DeviceID,
		"cpuUsage":    metrics["cpuUsage"],
		"memoryUsage": metrics["memoryUsage"],
		"diskUsage":   metrics["diskUsage"],
		"networkIn":   metrics["networkIn"],
		"networkOut":  metrics["networkOut"],
	}

	resp, err := postJSON(cfg.ServerURL+"/agent/metrics", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Metrics sent:", resp.Status)
	return nil
}
