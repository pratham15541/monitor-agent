package metrics

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type BackendSnapshot struct {
	Available       bool      `json:"available"`
	CPUUsage        float64   `json:"cpuUsage"`
	MemoryUsedBytes float64   `json:"memoryUsedBytes"`
	Error           string    `json:"error,omitempty"`
	CollectedAt     time.Time `json:"collectedAt"`
}

type ProcessSample struct {
	CPUPercent  float64   `json:"cpuPercent"`
	MemoryBytes uint64    `json:"memoryBytes"`
	Goroutines  int       `json:"goroutines"`
	CollectedAt time.Time `json:"collectedAt"`
}

type Snapshot struct {
	CompaniesCreated          int64             `json:"companiesCreated"`
	UsersCreated              int64             `json:"usersCreated"`
	SystemsCreated            int64             `json:"systemsCreated"`
	AgentsRegistered          int64             `json:"agentsRegistered"`
	AuthenticationSuccess     int64             `json:"authenticationSuccess"`
	AuthenticationFailure     int64             `json:"authenticationFailure"`
	RegistrationFailure       int64             `json:"registrationFailure"`
	ConnectedAgents           int64             `json:"connectedAgents"`
	PeakConcurrentConnections int64             `json:"peakConcurrentConnections"`
	ReconnectAttempts         int64             `json:"reconnectAttempts"`
	ReconnectSuccess          int64             `json:"reconnectSuccess"`
	HeartbeatsSent            int64             `json:"heartbeatsSent"`
	TelemetryMessagesSent     int64             `json:"telemetryMessagesSent"`
	TelemetryPerSecond        float64           `json:"telemetryPerSecond"`
	CommandsReceived          int64             `json:"commandsReceived"`
	CommandsCompleted         int64             `json:"commandsCompleted"`
	CommandsFailed            int64             `json:"commandsFailed"`
	AverageRegistrationTime   time.Duration     `json:"averageRegistrationTime"`
	AverageConnectionTime     time.Duration     `json:"averageConnectionTime"`
	AverageTelemetryLatency   time.Duration     `json:"averageTelemetryLatency"`
	P50                       time.Duration     `json:"p50"`
	P95                       time.Duration     `json:"p95"`
	P99                       time.Duration     `json:"p99"`
	BackendResponseTime       time.Duration     `json:"backendResponseTime"`
	Duration                  time.Duration     `json:"duration"`
	BackendSamples            []BackendSnapshot `json:"backendSamples"`
	LoadTesterSamples         []ProcessSample    `json:"loadTesterSamples"`
}

type Collector struct {
	startedAt time.Time

	companiesCreated atomic.Int64
	usersCreated     atomic.Int64
	systemsCreated   atomic.Int64
	agentsRegistered atomic.Int64
	authSuccess      atomic.Int64
	authFailure      atomic.Int64
	registrationFail atomic.Int64
	connectedAgents  atomic.Int64
	peakConnections  atomic.Int64
	reconnectAttempts atomic.Int64
	reconnectSuccess  atomic.Int64
	heartbeatsSent   atomic.Int64
	telemetrySent    atomic.Int64
	commandsReceived atomic.Int64
	commandsCompleted atomic.Int64
	commandsFailed   atomic.Int64

	registrationMu sync.Mutex
	registrations   []time.Duration
	connectionMu    sync.Mutex
	connections     []time.Duration
	telemetryMu     sync.Mutex
	telemetry       []time.Duration
	backendMu       sync.Mutex
	backendTimes    []time.Duration
	backendSampleMu sync.Mutex
	backendSamples  []BackendSnapshot
	loadSampleMu    sync.Mutex
	loadSamples     []ProcessSample
}

func NewCollector() *Collector {
	return &Collector{startedAt: time.Now()}
}

func (c *Collector) IncCompaniesCreated() { c.companiesCreated.Add(1) }
func (c *Collector) IncUsersCreated() { c.usersCreated.Add(1) }
func (c *Collector) IncSystemsCreated(n int64) { c.systemsCreated.Add(n) }
func (c *Collector) IncAgentsRegistered() { c.agentsRegistered.Add(1) }
func (c *Collector) IncAuthenticationSuccess() { c.authSuccess.Add(1) }
func (c *Collector) IncAuthenticationFailure() { c.authFailure.Add(1) }
func (c *Collector) IncRegistrationFailure() { c.registrationFail.Add(1) }
func (c *Collector) IncReconnectAttempts() { c.reconnectAttempts.Add(1) }
func (c *Collector) IncReconnectSuccess() { c.reconnectSuccess.Add(1) }
func (c *Collector) IncHeartbeatsSent(n int64) { c.heartbeatsSent.Add(n) }
func (c *Collector) IncTelemetryMessagesSent(n int64) { c.telemetrySent.Add(n) }
func (c *Collector) IncCommandsReceived() { c.commandsReceived.Add(1) }
func (c *Collector) IncCommandsCompleted() { c.commandsCompleted.Add(1) }
func (c *Collector) IncCommandsFailed() { c.commandsFailed.Add(1) }

func (c *Collector) AddConnected(delta int64) int64 {
	value := c.connectedAgents.Add(delta)
	c.updatePeak(value)
	return value
}

func (c *Collector) RecordRegistration(duration time.Duration) {
	c.registrationMu.Lock()
	c.registrations = append(c.registrations, duration)
	c.registrationMu.Unlock()
}

func (c *Collector) RecordConnection(duration time.Duration) {
	c.connectionMu.Lock()
	c.connections = append(c.connections, duration)
	c.connectionMu.Unlock()
}

func (c *Collector) RecordTelemetry(duration time.Duration) {
	c.telemetryMu.Lock()
	c.telemetry = append(c.telemetry, duration)
	c.telemetryMu.Unlock()
}

func (c *Collector) RecordBackendResponse(duration time.Duration) {
	c.backendMu.Lock()
	c.backendTimes = append(c.backendTimes, duration)
	c.backendMu.Unlock()
}

func (c *Collector) AddBackendSnapshot(snapshot BackendSnapshot) {
	c.backendSampleMu.Lock()
	c.backendSamples = append(c.backendSamples, snapshot)
	c.backendSampleMu.Unlock()
}

func (c *Collector) AddLoadSample(sample ProcessSample) {
	c.loadSampleMu.Lock()
	c.loadSamples = append(c.loadSamples, sample)
	c.loadSampleMu.Unlock()
}

func (c *Collector) updatePeak(value int64) {
	for {
		current := c.peakConnections.Load()
		if value <= current {
			return
		}
		if c.peakConnections.CompareAndSwap(current, value) {
			return
		}
	}
}

func (c *Collector) Snapshot() Snapshot {
	registrations := append([]time.Duration(nil), c.registrations...)
	connections := append([]time.Duration(nil), c.connections...)
	telemetry := append([]time.Duration(nil), c.telemetry...)
	backendTimes := append([]time.Duration(nil), c.backendTimes...)
	backendSamples := append([]BackendSnapshot(nil), c.backendSamples...)
	loadSamples := append([]ProcessSample(nil), c.loadSamples...)
	elapsed := time.Since(c.startedAt).Seconds()
	telemetryPerSecond := 0.0
	if elapsed > 0 {
		telemetryPerSecond = float64(c.telemetrySent.Load()) / elapsed
	}
	return Snapshot{
		CompaniesCreated:          c.companiesCreated.Load(),
		UsersCreated:              c.usersCreated.Load(),
		SystemsCreated:            c.systemsCreated.Load(),
		AgentsRegistered:          c.agentsRegistered.Load(),
		AuthenticationSuccess:     c.authSuccess.Load(),
		AuthenticationFailure:     c.authFailure.Load(),
		RegistrationFailure:       c.registrationFail.Load(),
		ConnectedAgents:           c.connectedAgents.Load(),
		PeakConcurrentConnections: c.peakConnections.Load(),
		ReconnectAttempts:         c.reconnectAttempts.Load(),
		ReconnectSuccess:          c.reconnectSuccess.Load(),
		HeartbeatsSent:            c.heartbeatsSent.Load(),
		TelemetryMessagesSent:     c.telemetrySent.Load(),
		TelemetryPerSecond:        telemetryPerSecond,
		CommandsReceived:          c.commandsReceived.Load(),
		CommandsCompleted:         c.commandsCompleted.Load(),
		CommandsFailed:            c.commandsFailed.Load(),
		AverageRegistrationTime:   average(registrations),
		AverageConnectionTime:     average(connections),
		AverageTelemetryLatency:   average(telemetry),
		P50:                       percentile(telemetry, 50),
		P95:                       percentile(telemetry, 95),
		P99:                       percentile(telemetry, 99),
		BackendResponseTime:       average(backendTimes),
		Duration:                  time.Since(c.startedAt),
		BackendSamples:            backendSamples,
		LoadTesterSamples:         loadSamples,
	}
}

func average(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}
	var total time.Duration
	for _, value := range values {
		total += value
	}
	return total / time.Duration(len(values))
}

func percentile(values []time.Duration, value float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]time.Duration(nil), values...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	index := int(math.Ceil((value/100.0)*float64(len(ordered)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(ordered) {
		index = len(ordered) - 1
	}
	return ordered[index]
}
