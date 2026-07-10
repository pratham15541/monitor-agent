package load

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"

	"loadtester/internal/api"
	"loadtester/internal/company"
	"loadtester/internal/config"
	"loadtester/internal/device"
	"loadtester/internal/metrics"
	"loadtester/internal/report"
	"loadtester/internal/utils"
)

type Runner struct {
	cfg       config.Config
	api       *api.Client
	companies *company.Manager
	devices   *device.Manager
	metrics   *metrics.Collector
	startedAt time.Time

	mu        sync.Mutex
	agents    []*AgentSimulator
	commands  *commandDriver
	phaseMu   sync.RWMutex
	phase     string
	backendMu sync.Mutex
	backend   metrics.BackendSnapshot
	rand      *rand.Rand
}

func NewRunner(cfg config.Config) (*Runner, error) {
	if err := config.EnsureDirectory(cfg.ReportDir); err != nil {
		return nil, err
	}
	client := api.NewClient(cfg.BackendURL, cfg.RequestTimeout)
	collector := metrics.NewCollector()
	return &Runner{
		cfg:       cfg,
		api:       client,
		companies: company.NewManager(client),
		devices:   device.NewManager(client, cfg.MaxConcurrentRegistrations),
		metrics:   collector,
		startedAt: time.Now(),
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (r *Runner) Run(ctx context.Context) error {
	r.setPhase("Creating companies")
	companyRecords, err := r.companies.CreateAll(ctx, r.cfg.Companies)
	if err != nil {
		r.metrics.IncAuthenticationFailure()
		return err
	}
	for range companyRecords {
		r.metrics.IncCompaniesCreated()
		r.metrics.IncUsersCreated()
		r.metrics.IncAuthenticationSuccess()
	}

	r.setPhase("Seeding systems")
	deviceRecords := make([]device.Record, 0, totalSystems(r.cfg.Companies))
	for index, record := range companyRecords {
		seeded, err := r.devices.SeedCompany(ctx, r.cfg, record)
		if err != nil {
			r.metrics.IncRegistrationFailure()
			return err
		}
		r.metrics.IncSystemsCreated(int64(len(seeded)))
		for range seeded {
			r.metrics.IncAgentsRegistered()
		}
		for i := range seeded {
			seeded[i].CompanyName = record.Name
			seeded[i].CompanyID = record.ID
			deviceRecords = append(deviceRecords, seeded[i])
		}
		companyRecords[index].Connected = 0
	}

	rampCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	r.setPhase("Running agents")
	activeDevices := r.selectActiveDevices(deviceRecords)
	go r.progressLoop(rampCtx, len(companyRecords), len(deviceRecords))
	go r.launchAgents(rampCtx, companyRecords, activeDevices)
	if r.cfg.RemoteCommands && len(companyRecords) > 0 {
		driver, err := newCommandDriver(rampCtx, companyRecords[0].JWT, activeDevices, r.cfg.WebSocketURL, r.metrics, r.cfg.CommandSampleSize)
		if err == nil {
			r.mu.Lock()
			r.commands = driver
			r.mu.Unlock()
			go driver.Run(rampCtx, r.cfg.CommandInterval)
		}
	}
	go r.sampleBackendLoop(rampCtx)
	go r.sampleLoadTesterLoop(rampCtx)
	go r.chaosLoop(rampCtx)

	runCtx, finish := context.WithTimeout(rampCtx, r.cfg.Duration)
	defer finish()
	<-runCtx.Done()
	r.setPhase("Finalizing")
	preStopSnapshot := r.metrics.Snapshot()
	r.stopAgents()
	r.stopCommands()
	return r.finish(companyRecords, deviceRecords, preStopSnapshot)
}

func (r *Runner) progressLoop(ctx context.Context, companyCount, systemCount int) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := r.metrics.Snapshot()
			fmt.Printf("[progress] phase=%s companies=%d/%d systems=%d/%d registered=%d connected=%d heartbeats=%d telemetry=%d reconnects=%d p95=%s\n",
				r.currentPhase(),
				snapshot.CompaniesCreated, companyCount,
				snapshot.SystemsCreated, systemCount,
				snapshot.AgentsRegistered,
				snapshot.ConnectedAgents,
				snapshot.HeartbeatsSent,
				snapshot.TelemetryMessagesSent,
				snapshot.ReconnectAttempts,
				snapshot.P95,
			)
		}
	}
}

func (r *Runner) setPhase(phase string) {
	r.phaseMu.Lock()
	r.phase = phase
	r.phaseMu.Unlock()
}

func (r *Runner) currentPhase() string {
	r.phaseMu.RLock()
	defer r.phaseMu.RUnlock()
	if r.phase == "" {
		return "starting"
	}
	return r.phase
}

func (r *Runner) launchAgents(ctx context.Context, companies []company.Record, devices []device.Record) {
	sort.Slice(devices, func(i, j int) bool { return devices[i].DeviceID < devices[j].DeviceID })
	total := len(devices)
	if total == 0 {
		return
	}
	schedule := rampSchedule(total, r.cfg.RampUp)
	start := time.Now()
	nextIndex := 0
	for nextIndex < total {
		select {
		case <-ctx.Done():
			return
		default:
		}

		desired := schedule(time.Since(start))
		for nextIndex < desired && nextIndex < total {
			companyRecord := findCompany(companies, devices[nextIndex].CompanyID)
			if companyRecord == nil {
				return
			}
			simulator := NewAgentSimulator(*companyRecord, devices[nextIndex], r.api, r.metrics, AgentConfig{
				WebSocketURL:      r.cfg.WebSocketURL,
				HeartbeatInterval: r.cfg.HeartbeatInterval,
				TelemetryInterval: r.cfg.TelemetryInterval,
				CommandInterval:   r.cfg.CommandInterval,
				RemoteCommands:    r.cfg.RemoteCommands,
				RequestTimeout:    r.cfg.RequestTimeout,
				BackendURL:        r.cfg.BackendURL,
			})
			simulator.Start(ctx)
			r.mu.Lock()
			r.agents = append(r.agents, simulator)
			r.mu.Unlock()
			nextIndex++
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (r *Runner) chaosLoop(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.ChaosInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.triggerChaos()
		}
	}
}

func (r *Runner) triggerChaos() {
	r.mu.Lock()
	agents := append([]*AgentSimulator(nil), r.agents...)
	r.mu.Unlock()
	if len(agents) == 0 {
		return
	}
	numberToDrop := int(float64(len(agents)) * (r.cfg.RandomDisconnectPercent / 100.0))
	if numberToDrop <= 0 {
		numberToDrop = 1
	}
	if numberToDrop > len(agents) {
		numberToDrop = len(agents)
	}
	r.rand.Shuffle(len(agents), func(i, j int) { agents[i], agents[j] = agents[j], agents[i] })
	for index := 0; index < numberToDrop; index++ {
		r.metrics.IncReconnectAttempts()
		agents[index].Disconnect()
	}
}

func (r *Runner) selectActiveDevices(devices []device.Record) []device.Record {
	limit := r.cfg.MaxConcurrentAgents
	if limit <= 0 || limit >= len(devices) {
		return devices
	}

	selected := append([]device.Record(nil), devices...)
	r.rand.Shuffle(len(selected), func(i, j int) { selected[i], selected[j] = selected[j], selected[i] })
	return selected[:limit]
}

func (r *Runner) sampleBackendLoop(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.BackendMetricsInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.sampleBackend(ctx)
		}
	}
}

func (r *Runner) sampleBackend(ctx context.Context) {
	var sample metrics.BackendSnapshot
	for _, metricName := range []string{"process.cpu.usage", "jvm.memory.used"} {
		start := time.Now()
		resp, status, err := r.api.GetActuatorMetric(ctx, metricName)
		r.metrics.RecordBackendResponse(time.Since(start))
		if err != nil || status < 200 || status >= 300 {
			sample.Available = false
			sample.Error = fmt.Sprintf("%s unavailable: %v", metricName, err)
			continue
		}
		sample.Available = true
		sample.CollectedAt = time.Now()
		if metricName == "process.cpu.usage" && len(resp.Measurements) > 0 {
			sample.CPUUsage = resp.Measurements[0].Value
		}
		if metricName == "jvm.memory.used" {
			var total float64
			for _, measurement := range resp.Measurements {
				total += measurement.Value
			}
			sample.MemoryUsedBytes = total
		}
	}
	r.backendMu.Lock()
	r.backend = sample
	r.backendMu.Unlock()
	r.metrics.AddBackendSnapshot(sample)
}

func (r *Runner) sampleLoadTesterLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpuPercent, err := proc.CPUPercent()
			if err != nil {
				continue
			}
			memInfo, err := proc.MemoryInfo()
			if err != nil {
				continue
			}
			r.metrics.AddLoadSample(metrics.ProcessSample{
				CPUPercent:  cpuPercent,
				MemoryBytes: memInfo.RSS,
				Goroutines:  runtime.NumGoroutine(),
				CollectedAt: time.Now(),
			})
		}
	}
}

func (r *Runner) stopAgents() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, agent := range r.agents {
		agent.Stop()
	}
}

func (r *Runner) stopCommands() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.commands != nil {
		r.commands.Stop()
		r.commands = nil
	}
}

func (r *Runner) finish(companies []company.Record, devices []device.Record, metricsSnapshot metrics.Snapshot) error {
	backend := metrics.BackendSnapshot{}
	r.backendMu.Lock()
	backend = r.backend
	r.backendMu.Unlock()
	metricsSnapshot.Duration = time.Since(r.startedAt)
	connectedByCompany := r.connectedAgentCounts()
	for index := range companies {
		companies[index].Connected = connectedByCompany[companies[index].ID]
	}
	result := report.Result{
		StartedAt:  r.startedAt,
		FinishedAt: time.Now(),
		Config:     r.cfg,
		Companies:  companies,
		Devices:    devices,
		Metrics:    metricsSnapshot,
		Backend:    backend,
	}
	result.Analysis = report.Analyze(result)
	if err := report.WriteAll(r.cfg.ReportDir, result); err != nil {
		return err
	}
	printSummary(result)
	return nil
}

func (r *Runner) connectedAgentCounts() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()

	counts := make(map[string]int, len(r.agents))
	for _, agent := range r.agents {
		counts[agent.company.ID]++
	}
	return counts
}

func totalSystems(specs []config.CompanyConfig) int {
	total := 0
	for _, spec := range specs {
		total += spec.Systems
	}
	return total
}

func findCompany(companies []company.Record, id string) *company.Record {
	for index := range companies {
		if companies[index].ID == id {
			return &companies[index]
		}
	}
	return nil
}

func rampSchedule(total int, rampUp time.Duration) func(time.Duration) int {
	if rampUp <= 0 {
		return func(time.Duration) int { return total }
	}
	return func(elapsed time.Duration) int {
		if elapsed >= rampUp {
			return total
		}
		fraction := float64(elapsed) / float64(rampUp)
		return utils.Clamp(int(fraction*float64(total)), 0, total)
	}
}

func printSummary(result report.Result) {
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("LOAD TEST SUMMARY")
	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Printf("Companies: %d\n", result.Metrics.CompaniesCreated)
	fmt.Printf("Systems: %d\n", result.Metrics.SystemsCreated)
	fmt.Printf("Registered: %d\n", result.Metrics.AgentsRegistered)
	fmt.Printf("Connected: %d\n", result.Metrics.ConnectedAgents)
	fmt.Printf("Authentication Success: %d\n", result.Metrics.AuthenticationSuccess)
	fmt.Printf("Authentication Failure: %d\n", result.Metrics.AuthenticationFailure)
	fmt.Printf("Peak Concurrent Connections: %d\n", result.Metrics.PeakConcurrentConnections)
	fmt.Printf("Heartbeats: %d\n", result.Metrics.HeartbeatsSent)
	fmt.Printf("Telemetry Messages: %d\n", result.Metrics.TelemetryMessagesSent)
	fmt.Printf("Telemetry/sec: %.2f\n", result.Metrics.TelemetryPerSecond)
	fmt.Printf("Commands Executed: %d\n", result.Metrics.CommandsCompleted)
	fmt.Printf("Reconnect Attempts: %d\n", result.Metrics.ReconnectAttempts)
	fmt.Printf("Reconnect Success: %d\n", result.Metrics.ReconnectSuccess)
	fmt.Printf("Average Registration Time: %s\n", result.Metrics.AverageRegistrationTime)
	fmt.Printf("Average Connection Time: %s\n", result.Metrics.AverageConnectionTime)
	fmt.Printf("Average Latency: %s\n", result.Metrics.AverageTelemetryLatency)
	fmt.Printf("P95: %s\n", result.Metrics.P95)
	fmt.Printf("P99: %s\n", result.Metrics.P99)
	fmt.Printf("Backend CPU: %.2f\n", result.Backend.CPUUsage)
	fmt.Printf("Backend Memory: %.2f\n", result.Backend.MemoryUsedBytes)
	fmt.Printf("Load Tester Memory Samples: %d\n", len(result.Metrics.LoadTesterSamples))
	fmt.Printf("Duration: %s\n", result.Metrics.Duration)
	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("GENERATED TEST DATA")
	fmt.Println()
	for _, companyRecord := range result.Companies {
		fmt.Printf("Company ID: %s\n", companyRecord.ID)
		fmt.Printf("Company: %s\n", companyRecord.Name)
		fmt.Printf("Email: %s\n", companyRecord.Email)
		fmt.Printf("Password: %s\n", companyRecord.Password)
		fmt.Printf("Systems: %d\n", companyRecord.Systems)
		fmt.Printf("Connected: %d\n", companyRecord.Connected)
		if companyRecord.JWT != "" {
			fmt.Printf("JWT Token: %s\n", companyRecord.JWT)
		}
		fmt.Println("----------------------------------------")
	}
	fmt.Println()
	fmt.Println("Analysis")
	fmt.Println(result.Analysis.Status)
	for _, finding := range result.Analysis.Findings {
		fmt.Println("-", finding)
	}
	if len(result.Analysis.Recommendations) > 0 {
		fmt.Println()
		fmt.Println("Recommendations")
		for _, recommendation := range result.Analysis.Recommendations {
			fmt.Println("-", recommendation)
		}
	}
}
