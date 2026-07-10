package telemetry

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"loadtester/internal/api"
	"loadtester/internal/utils"
)

type Profile struct {
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ipAddress"`
	OS           string `json:"os"`
	CPU          string `json:"cpu"`
	RAM          string `json:"ram"`
	Disk         string `json:"disk"`
	Architecture string `json:"architecture"`
	MACAddress   string `json:"macAddress"`
	Timezone     string `json:"timezone"`
	Version      string `json:"version"`
}

type Generator struct {
	rng      *rand.Rand
	profile  Profile
	cpu      float64
	memory   float64
	disk     float64
	netIn    float64
	netOut   float64
	temp     float64
	uptime   float64
}

func NewGenerator(seed int64, profile Profile) *Generator {
	rng := rand.New(rand.NewSource(seed))
	return &Generator{
		rng:     rng,
		profile: profile,
		cpu:     20 + rng.Float64()*25,
		memory:  30 + rng.Float64()*30,
		disk:    40 + rng.Float64()*25,
		netIn:   rng.Float64() * 1000,
		netOut:  rng.Float64() * 1000,
		temp:    35 + rng.Float64()*8,
		uptime:  rng.Float64() * 3600,
	}
}

func (g *Generator) NextMetric(deviceID string) api.MetricRequest {
	g.cpu = clamp(g.cpu+noise(g.rng, 4.5), 1, 100)
	g.memory = clamp(g.memory+noise(g.rng, 3.0), 1, 100)
	g.disk = clamp(g.disk+noise(g.rng, 1.5), 1, 100)
	g.netIn = math.Max(0, g.netIn+math.Abs(noise(g.rng, 200)))
	g.netOut = math.Max(0, g.netOut+math.Abs(noise(g.rng, 220)))
	g.temp = clamp(g.temp+noise(g.rng, 0.7), 25, 95)
	g.uptime += 5 + g.rng.Float64()*25
	return api.MetricRequest{
		DeviceID:    deviceID,
		CPUUsage:    round2(g.cpu),
		MemoryUsage: round2(g.memory),
		DiskUsage:   round2(g.disk),
		NetworkIn:   round2(g.netIn),
		NetworkOut:  round2(g.netOut),
	}
}

func (g *Generator) NextDetail(deviceID string) api.MetricDetailRequest {
	collectedAt := time.Now().UTC()
	processes := make([]map[string]any, 0, 8)
	for index := 0; index < 5+g.rng.Intn(7); index++ {
		processes = append(processes, map[string]any{
			"pid":              1000 + g.rng.Intn(60000),
			"name":             fmt.Sprintf("proc-%03d", index+1),
			"exe":              "/usr/bin/proc",
			"cmdline":          fmt.Sprintf("/usr/bin/proc --worker=%d", index+1),
			"username":         "monitor",
			"status":           []string{"running", "sleeping", "idle"}[index%3],
			"ppid":             1 + g.rng.Intn(5000),
			"createTime":       collectedAt.Add(-time.Duration(5+g.rng.Intn(1200)) * time.Second).Format(time.RFC3339),
			"isRunning":        true,
			"threads":          1 + g.rng.Intn(16),
			"cpuPercent":       round2(math.Abs(noise(g.rng, 6))),
			"memoryRssBytes":   int64(1024 * 1024 * (32 + g.rng.Intn(2048))),
			"memoryVmsBytes":   int64(1024 * 1024 * (64 + g.rng.Intn(4096))),
			"memoryPercent":    round2(1 + g.rng.Float64()*12),
			"ioReadBytes":      int64(g.rng.Intn(5_000_000)),
			"ioWriteBytes":     int64(g.rng.Intn(5_000_000)),
		})
	}

	connections := make([]map[string]any, 0, 4)
	for index := 0; index < g.rng.Intn(4); index++ {
		connections = append(connections, map[string]any{
			"pid":    1000 + g.rng.Intn(60000),
			"family": []string{"ipv4", "ipv6"}[index%2],
			"type":   []string{"tcp", "udp"}[index%2],
			"status": []string{"ESTABLISHED", "TIME_WAIT", "CLOSE_WAIT"}[index%3],
			"local":  fmt.Sprintf("%s:%d", g.profile.IPAddress, 10000+g.rng.Intn(40000)),
			"remote": fmt.Sprintf("172.16.%d.%d:%d", g.rng.Intn(255), g.rng.Intn(255), 1000+g.rng.Intn(50000)),
		})
	}

	memory := map[string]any{
		"total":        int64(8 * 1024 * 1024 * 1024),
		"available":    int64(3 * 1024 * 1024 * 1024),
		"used":         int64(5 * 1024 * 1024 * 1024),
		"usedPercent":  round2(g.memory),
		"free":         int64(2 * 1024 * 1024 * 1024),
		"cached":       int64(512 * 1024 * 1024),
		"buffers":      int64(128 * 1024 * 1024),
		"active":       int64(2 * 1024 * 1024 * 1024),
		"inactive":     int64(1 * 1024 * 1024 * 1024),
		"shared":       int64(256 * 1024 * 1024),
		"slab":         int64(64 * 1024 * 1024),
		"pageTables":   int64(16 * 1024 * 1024),
		"swapCached":   int64(8 * 1024 * 1024),
		"sreclaimable": int64(4 * 1024 * 1024),
		"sunreclaim":   int64(4 * 1024 * 1024),
		"swapTotal":    int64(2 * 1024 * 1024 * 1024),
		"swapUsed":     int64(256 * 1024 * 1024),
	}

	services := []map[string]any{
		{"name": "monitor-agent", "status": "running"},
		{"name": "postgres", "status": "running"},
		{"name": "dashboard", "status": "running"},
	}

	logs := map[string]any{
		"agent": []string{
			fmt.Sprintf("[%s] heartbeat ok", collectedAt.Format(time.RFC3339)),
			fmt.Sprintf("[%s] telemetry flushed", collectedAt.Add(-5*time.Second).Format(time.RFC3339)),
		},
		"system": []string{fmt.Sprintf("%s system log tail", g.profile.OS)},
	}

	details := map[string]any{
		"collectedAt": collectedAt.Format(time.RFC3339),
		"processes":    processes,
		"connections":  connections,
		"memory":       memory,
		"services":     services,
		"logs":         logs,
		"os":           g.profile.OS,
	}

	return api.MetricDetailRequest{DeviceID: deviceID, Details: details}
}

func (g *Generator) ExecuteCommand(deviceID string, request api.CommandRequest) []api.CommandResult {
	started := time.Now().UTC()
	result := api.CommandResult{
		DeviceID:  deviceID,
		CommandID: request.CommandID,
		Type:      request.Type,
		Status:    "ok",
		StartedAt: started.Format(time.RFC3339),
	}

	switch strings.ToLower(strings.TrimSpace(request.Type)) {
	case "shell":
		result.Output = simulateOutput(g.rng, request.Payload, 2+g.rng.Intn(4))
	case "service":
		result.Output = fmt.Sprintf("simulated service action: %s", request.Payload)
	case "diagnostics":
		payload, _ := json.Marshal(map[string]any{
			"hostname":     g.profile.Hostname,
			"os":           g.profile.OS,
			"cpu":          g.profile.CPU,
			"memory":       g.profile.RAM,
			"disk":         g.profile.Disk,
			"architecture": g.profile.Architecture,
			"uptimeSeconds": int(g.uptime),
		})
		result.Output = string(payload)
	case "collect-details":
		result.Output = "detailed metrics collected"
	case "timeout":
		result.Status = "timeout"
		result.Error = "simulated timeout"
	case "failure":
		result.Status = "error"
		result.Error = "simulated command failure"
	default:
		result.Status = "error"
		result.Error = "unknown command type"
	}

	finished := time.Now().UTC().Format(time.RFC3339)
	results := chunkResult(result, finished)
	for index := range results {
		if results[index].Chunked {
			results[index].ChunkIndex = index
		}
	}
	return results
}

func chunkResult(result api.CommandResult, finished string) []api.CommandResult {
	outputChunks := splitChunks(result.Output, 12*1024)
	errorChunks := splitChunks(result.Error, 12*1024)
	if len(outputChunks) <= 1 && len(errorChunks) <= 1 {
		result.FinishedAt = finished
		return []api.CommandResult{result}
	}

	items := make([]api.CommandResult, 0, len(outputChunks)+len(errorChunks))
	if len(outputChunks) > 0 {
		for index, chunk := range outputChunks {
			item := result
			item.Output = chunk
			item.Error = ""
			item.Chunked = true
			item.ChunkType = "output"
			item.ChunkIndex = index
			item.ChunkCount = len(outputChunks)
			if index == len(outputChunks)-1 && len(errorChunks) == 0 {
				item.FinishedAt = finished
			}
			items = append(items, item)
		}
	}
	if len(errorChunks) > 0 {
		for index, chunk := range errorChunks {
			item := result
			item.Output = ""
			item.Error = chunk
			item.Chunked = true
			item.ChunkType = "error"
			item.ChunkIndex = index
			item.ChunkCount = len(errorChunks)
			if index == len(errorChunks)-1 {
				item.FinishedAt = finished
			}
			items = append(items, item)
		}
	}
	return items
}

func splitChunks(value string, size int) []string {
	if value == "" {
		return nil
	}
	if len(value) <= size || size <= 0 {
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

func simulateOutput(rng *rand.Rand, payload string, lines int) string {
	lines = utils.Clamp(lines, 1, 12)
	var builder strings.Builder
	for index := 0; index < lines; index++ {
		builder.WriteString(fmt.Sprintf("line %d: %s\n", index+1, payload))
		builder.WriteString(fmt.Sprintf("status=%s checksum=%08x\n", []string{"ok", "warn", "info"}[index%3], rng.Uint32()))
	}
	return builder.String()
}

func noise(rng *rand.Rand, sigma float64) float64 {
	return rng.NormFloat64() * sigma
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
