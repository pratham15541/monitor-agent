package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"loadtester/internal/company"
	"loadtester/internal/config"
	"loadtester/internal/device"
	"loadtester/internal/metrics"
)

type Analysis struct {
	Status          string   `json:"status"`
	Findings        []string `json:"findings"`
	Recommendations []string `json:"recommendations"`
}

type Result struct {
	StartedAt time.Time              `json:"startedAt"`
	FinishedAt time.Time             `json:"finishedAt"`
	Config    config.Config          `json:"config"`
	Companies []company.Record       `json:"companies"`
	Devices   []device.Record        `json:"devices"`
	Metrics   metrics.Snapshot       `json:"metrics"`
	Backend   metrics.BackendSnapshot `json:"backend"`
	Analysis  Analysis               `json:"analysis"`
}

func WriteAll(dir string, result Result) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "load-test-report.json"), result); err != nil {
		return err
	}
	if err := writeMarkdown(filepath.Join(dir, "load-test-report.md"), result); err != nil {
		return err
	}
	if err := writeCSV(filepath.Join(dir, "load-test-report.csv"), result); err != nil {
		return err
	}
	return nil
}

func Analyze(result Result) Analysis {
	analysis := Analysis{Status: "PASS"}
	metricsSnapshot := result.Metrics

	if metricsSnapshot.AuthenticationFailure > 0 {
		analysis.Status = "WARN"
		analysis.Findings = append(analysis.Findings, fmt.Sprintf("authentication failures: %d", metricsSnapshot.AuthenticationFailure))
		analysis.Recommendations = append(analysis.Recommendations, "Inspect auth throughput and password/token provisioning")
	}
	if metricsSnapshot.RegistrationFailure > 0 {
		analysis.Status = "WARN"
		analysis.Findings = append(analysis.Findings, fmt.Sprintf("registration failures: %d", metricsSnapshot.RegistrationFailure))
		analysis.Recommendations = append(analysis.Recommendations, "Increase registration worker capacity or check backend validation pressure")
	}
	if metricsSnapshot.P95 > 2*time.Second {
		analysis.Status = "WARN"
		analysis.Findings = append(analysis.Findings, fmt.Sprintf("telemetry p95 is high: %s", metricsSnapshot.P95))
		analysis.Recommendations = append(analysis.Recommendations, "Increase websocket buffer sizes or reduce batch frequency")
	}
	if metricsSnapshot.CommandsFailed > 0 {
		analysis.Status = "WARN"
		analysis.Findings = append(analysis.Findings, fmt.Sprintf("command failures: %d", metricsSnapshot.CommandsFailed))
	}
	if len(result.Backend.Error) > 0 {
		analysis.Status = "WARN"
		analysis.Findings = append(analysis.Findings, fmt.Sprintf("backend metrics unavailable: %s", result.Backend.Error))
	}
	if len(metricsSnapshot.LoadTesterSamples) >= 2 {
		first := metricsSnapshot.LoadTesterSamples[0]
		last := metricsSnapshot.LoadTesterSamples[len(metricsSnapshot.LoadTesterSamples)-1]
		if last.MemoryBytes > first.MemoryBytes*12/10 {
			analysis.Status = "WARN"
			analysis.Findings = append(analysis.Findings, "load tester memory increased materially during the run")
			analysis.Recommendations = append(analysis.Recommendations, "Review goroutine lifetimes and connection shutdown paths")
		}
	}
	if len(metricsSnapshot.BackendSamples) >= 2 {
		first := metricsSnapshot.BackendSamples[0]
		last := metricsSnapshot.BackendSamples[len(metricsSnapshot.BackendSamples)-1]
		if first.Available && last.Available && last.MemoryUsedBytes > first.MemoryUsedBytes*12/10 {
			analysis.Status = "WARN"
			analysis.Findings = append(analysis.Findings, "backend memory increased materially during the run")
			analysis.Recommendations = append(analysis.Recommendations, "Watch database write pressure and websocket fan-out allocations")
		}
	}
	if metricsSnapshot.ConnectedAgents < metricsSnapshot.AgentsRegistered {
		analysis.Status = "WARN"
		analysis.Findings = append(analysis.Findings, "not all agents stayed connected through the test window")
	}
	if analysis.Status == "PASS" {
		analysis.Findings = append(analysis.Findings, "authentication stable")
		analysis.Findings = append(analysis.Findings, "registration stable")
		analysis.Findings = append(analysis.Findings, "no connection storms detected")
		analysis.Findings = append(analysis.Findings, "backend sustained the configured agent set")
	}
	return analysis
}

func writeJSON(path string, result Result) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func writeMarkdown(path string, result Result) error {
	var builder strings.Builder
	builder.WriteString("==========================================\n\n")
	builder.WriteString("LOAD TEST SUMMARY\n\n")
	builder.WriteString("==========================================\n\n")
	builder.WriteString(fmt.Sprintf("Started: %s\n", result.StartedAt.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("Finished: %s\n", result.FinishedAt.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("Duration: %s\n\n", result.Metrics.Duration))
	builder.WriteString(fmt.Sprintf("Companies: %d\n", result.Metrics.CompaniesCreated))
	builder.WriteString(fmt.Sprintf("Systems: %d\n", result.Metrics.SystemsCreated))
	builder.WriteString(fmt.Sprintf("Registered: %d\n", result.Metrics.AgentsRegistered))
	builder.WriteString(fmt.Sprintf("Connected: %d\n", result.Metrics.ConnectedAgents))
	builder.WriteString(fmt.Sprintf("Authentication Success: %d\n", result.Metrics.AuthenticationSuccess))
	builder.WriteString(fmt.Sprintf("Authentication Failure: %d\n", result.Metrics.AuthenticationFailure))
	builder.WriteString(fmt.Sprintf("Peak Concurrent Connections: %d\n", result.Metrics.PeakConcurrentConnections))
	builder.WriteString(fmt.Sprintf("Heartbeats: %d\n", result.Metrics.HeartbeatsSent))
	builder.WriteString(fmt.Sprintf("Telemetry Messages: %d\n", result.Metrics.TelemetryMessagesSent))
	builder.WriteString(fmt.Sprintf("Telemetry/sec: %.2f\n", result.Metrics.TelemetryPerSecond))
	builder.WriteString(fmt.Sprintf("Commands Executed: %d\n", result.Metrics.CommandsCompleted))
	builder.WriteString(fmt.Sprintf("Reconnect Attempts: %d\n", result.Metrics.ReconnectAttempts))
	builder.WriteString(fmt.Sprintf("Reconnect Success: %d\n", result.Metrics.ReconnectSuccess))
	builder.WriteString(fmt.Sprintf("Average Registration Time: %s\n", result.Metrics.AverageRegistrationTime))
	builder.WriteString(fmt.Sprintf("Average Connection Time: %s\n", result.Metrics.AverageConnectionTime))
	builder.WriteString(fmt.Sprintf("Average Latency: %s\n", result.Metrics.AverageTelemetryLatency))
	builder.WriteString(fmt.Sprintf("P95: %s\n", result.Metrics.P95))
	builder.WriteString(fmt.Sprintf("P99: %s\n", result.Metrics.P99))
	builder.WriteString(fmt.Sprintf("Backend CPU: %.2f\n", result.Backend.CPUUsage))
	builder.WriteString(fmt.Sprintf("Backend Memory: %.2f\n", result.Backend.MemoryUsedBytes))
	builder.WriteString(fmt.Sprintf("Load Tester Memory Samples: %d\n\n", len(result.Metrics.LoadTesterSamples)))
	builder.WriteString("Generated Test Data\n\n")
	for _, company := range result.Companies {
		builder.WriteString(fmt.Sprintf("Company ID: %s\n", company.ID))
		builder.WriteString(fmt.Sprintf("Company: %s\n", company.Name))
		builder.WriteString(fmt.Sprintf("Email: %s\n", company.Email))
		builder.WriteString(fmt.Sprintf("Password: %s\n", company.Password))
		builder.WriteString(fmt.Sprintf("Systems: %d\n", company.Systems))
		builder.WriteString(fmt.Sprintf("Connected: %d\n", company.Connected))
		if company.JWT != "" {
			builder.WriteString(fmt.Sprintf("JWT Token: %s\n", company.JWT))
		}
		builder.WriteString("----------------------------------------\n")
	}
	builder.WriteString("\nAnalysis\n\n")
	builder.WriteString(result.Analysis.Status + "\n")
	for _, finding := range result.Analysis.Findings {
		builder.WriteString("- " + finding + "\n")
	}
	if len(result.Analysis.Recommendations) > 0 {
		builder.WriteString("\nRecommendations\n\n")
		for _, recommendation := range result.Analysis.Recommendations {
			builder.WriteString("- " + recommendation + "\n")
		}
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func writeCSV(path string, result Result) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"kind", "id", "name", "email", "password", "apiToken", "jwt", "systems", "connected", "status"}); err != nil {
		return err
	}
	for _, company := range result.Companies {
		if err := writer.Write([]string{"company", company.ID, company.Name, company.Email, company.Password, company.APIToken, company.JWT, fmt.Sprintf("%d", company.Systems), fmt.Sprintf("%d", company.Connected), result.Analysis.Status}); err != nil {
			return err
		}
	}
	return writer.Write([]string{"summary", "", "", "", "", "", "", fmt.Sprintf("%d", result.Metrics.SystemsCreated), fmt.Sprintf("%d", result.Metrics.ConnectedAgents), result.Analysis.Status})
}

func SortCompanies(companies []company.Record) {
	sort.Slice(companies, func(i, j int) bool { return companies[i].Name < companies[j].Name })
}
