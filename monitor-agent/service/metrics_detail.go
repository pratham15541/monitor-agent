package service

import (
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	maxProcessEntries    = 100
	maxConnectionEntries = 200
	maxPathLength        = 256
)

func CollectDetailedMetrics() map[string]interface{} {
	details := map[string]interface{}{
		"collectedAt": time.Now().UTC().Format(time.RFC3339),
		"processes":   collectProcessMetrics(),
		"connections": collectConnections(),
		"memory":      collectMemoryDetails(),
		"services":    collectServicesSnapshot(),
		"logs":        collectLogsSnapshot(),
		"os":          runtime.GOOS,
	}

	return details
}

func collectProcessMetrics() []map[string]interface{} {
	procs, err := process.Processes()
	if err != nil {
		return []map[string]interface{}{}
	}

	results := make([]map[string]interface{}, 0, len(procs))
	for _, proc := range procs {
		info := map[string]interface{}{
			"pid": int(proc.Pid),
		}

		if name, err := proc.Name(); err == nil {
			info["name"] = name
		}
		if exe, err := proc.Exe(); err == nil {
			info["exe"] = trimString(exe, maxPathLength)
		}
		if cmdline, err := proc.Cmdline(); err == nil {
			info["cmdline"] = trimString(cmdline, maxPathLength)
		}
		if username, err := proc.Username(); err == nil {
			info["username"] = username
		}
		if status, err := proc.Status(); err == nil && len(status) > 0 {
			info["status"] = status
		}
		if ppid, err := proc.Ppid(); err == nil {
			info["ppid"] = int(ppid)
		}
		if createTime, err := proc.CreateTime(); err == nil {
			info["createTime"] = createTime
		}
		if isRunning, err := proc.IsRunning(); err == nil {
			info["isRunning"] = isRunning
		}
		if numThreads, err := proc.NumThreads(); err == nil {
			info["threads"] = numThreads
		}
		if cpuPercent, err := proc.Percent(0); err == nil {
			info["cpuPercent"] = cpuPercent
		}
		if memInfo, err := proc.MemoryInfo(); err == nil {
			info["memoryRssBytes"] = memInfo.RSS
			info["memoryVmsBytes"] = memInfo.VMS
		}
		if memPercent, err := proc.MemoryPercent(); err == nil {
			info["memoryPercent"] = memPercent
		}
		if ioCounters, err := proc.IOCounters(); err == nil {
			info["ioReadBytes"] = ioCounters.ReadBytes
			info["ioWriteBytes"] = ioCounters.WriteBytes
		}

		results = append(results, info)
	}

	sort.Slice(results, func(i, j int) bool {
		leftCPU := getFloat64(results[i]["cpuPercent"])
		rightCPU := getFloat64(results[j]["cpuPercent"])
		if leftCPU == rightCPU {
			return getFloat64(results[i]["memoryRssBytes"]) > getFloat64(results[j]["memoryRssBytes"])
		}
		return leftCPU > rightCPU
	})

	if len(results) > maxProcessEntries {
		results = results[:maxProcessEntries]
	}

	return results
}

func collectConnections() []map[string]interface{} {
	conns, err := net.Connections("all")
	if err != nil {
		return []map[string]interface{}{}
	}

	results := make([]map[string]interface{}, 0, len(conns))
	for _, conn := range conns {
		info := map[string]interface{}{
			"pid":    conn.Pid,
			"family": conn.Family,
			"type":   conn.Type,
			"status": conn.Status,
			"local": map[string]interface{}{
				"ip":   conn.Laddr.IP,
				"port": conn.Laddr.Port,
			},
			"remote": map[string]interface{}{
				"ip":   conn.Raddr.IP,
				"port": conn.Raddr.Port,
			},
		}
		results = append(results, info)
	}

	if len(results) > maxConnectionEntries {
		results = results[:maxConnectionEntries]
	}

	return results
}

func getFloat64(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case uint64:
		return float64(typed)
	default:
		return 0
	}
}

func trimString(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}

	return value[:maxLen]
}

func collectMemoryDetails() map[string]interface{} {
	stats, err := mem.VirtualMemory()
	if err != nil {
		return map[string]interface{}{}
	}

	details := map[string]interface{}{
		"total":        stats.Total,
		"available":    stats.Available,
		"used":         stats.Used,
		"usedPercent":  stats.UsedPercent,
		"free":         stats.Free,
		"cached":       stats.Cached,
		"buffers":      stats.Buffers,
		"active":       stats.Active,
		"inactive":     stats.Inactive,
		"shared":       stats.Shared,
		"slab":         stats.Slab,
		"pageTables":   stats.PageTables,
		"swapCached":   stats.SwapCached,
		"sreclaimable": stats.Sreclaimable,
		"sunreclaim":   stats.Sunreclaim,
	}

	if swapStats, err := mem.SwapMemory(); err == nil {
		details["swapTotal"] = swapStats.Total
		details["swapUsed"] = swapStats.Used
		details["swapFree"] = swapStats.Free
		details["swapUsedPercent"] = swapStats.UsedPercent
	}

	return details
}
