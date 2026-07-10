package load

import (
	"context"
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"loadtester/internal/api"
	"loadtester/internal/company"
	"loadtester/internal/device"
	"loadtester/internal/metrics"
	"loadtester/internal/stomp"
	"loadtester/internal/telemetry"
	"loadtester/internal/utils"
)

type AgentConfig struct {
	WebSocketURL       string
	HeartbeatInterval  time.Duration
	TelemetryInterval  time.Duration
	CommandInterval    time.Duration
	RemoteCommands     bool
	RequestTimeout     time.Duration
	BackendURL         string
}

type AgentSimulator struct {
	company company.Record
	device  device.Record
	client  *api.Client
	metrics *metrics.Collector
	cfg     AgentConfig
	gen     *telemetry.Generator
	rng     *rand.Rand

	mu          sync.Mutex
	metricsConn *stomp.Client
	commandConn *stomp.Client
	closed      chan struct{}
	stopOnce    sync.Once
	connected   bool
}

func NewAgentSimulator(companyRecord company.Record, deviceRecord device.Record, client *api.Client, collector *metrics.Collector, cfg AgentConfig) *AgentSimulator {
	seed := time.Now().UnixNano() + int64(len(companyRecord.ID)+len(deviceRecord.DeviceID))
	return &AgentSimulator{
		company: companyRecord,
		device:  deviceRecord,
		client:  client,
		metrics: collector,
		cfg:     cfg,
		rng:     rand.New(rand.NewSource(seed)),
		gen:     telemetry.NewGenerator(seed, deviceRecord.TelemetryProfile),
		closed:  make(chan struct{}),
	}
}

func (a *AgentSimulator) Start(ctx context.Context) {
	go a.runMetricsLoop(ctx)
	go a.runDetailedTelemetryLoop(ctx)
	go a.runCommandLoop(ctx)
}

func (a *AgentSimulator) Stop() {
	a.stopOnce.Do(func() {
		close(a.closed)
		a.closeConnections()
	})
}

func (a *AgentSimulator) Disconnect() {
	a.closeConnections()
}

func (a *AgentSimulator) runMetricsLoop(ctx context.Context) {
	backoff := 500 * time.Millisecond
	for {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-a.closed:
			return
		default:
		}

		start := time.Now()
		conn, err := stomp.Dial(ctx, a.cfg.WebSocketURL, map[string]string{"x-agent-token": a.company.APIToken})
		if err != nil {
			a.metrics.IncReconnectAttempts()
			time.Sleep(backoff)
			backoff = utils.Clamp(backoff*2, 500*time.Millisecond, 15*time.Second)
			continue
		}
		backoff = 500 * time.Millisecond

		a.mu.Lock()
		a.metricsConn = conn
		a.mu.Unlock()

		a.metrics.IncReconnectSuccess()
		a.metrics.RecordConnection(time.Since(start))
		if !a.markConnected(true) {
			a.metrics.AddConnected(1)
		}

		if err := a.metricsSession(ctx, conn); err != nil {
			_ = conn.Close()
			a.markConnected(false)
			a.metrics.AddConnected(-1)
			time.Sleep(backoff)
			continue
		}
		return
	}
}

func (a *AgentSimulator) metricsSession(ctx context.Context, conn *stomp.Client) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-a.closed:
			return context.Canceled
		default:
		}

		start := time.Now()
		batch := make([]api.MetricRequest, 0, 10)
		for len(batch) < 10 {
			batch = append(batch, a.gen.NextMetric(a.device.DeviceID))
		}
		body, err := json.Marshal(batch)
		if err != nil {
			return err
		}
		if err := conn.Send("/app/agent/metrics-batch", body, map[string]string{"content-type": "application/json"}); err != nil {
			return err
		}
		a.metrics.IncHeartbeatsSent(1)
		a.metrics.IncTelemetryMessagesSent(int64(len(batch)))
		a.metrics.RecordTelemetry(time.Since(start))

		select {
		case <-time.After(a.cfg.HeartbeatInterval):
		case <-ctx.Done():
			return ctx.Err()
		case <-a.closed:
			return context.Canceled
		}
	}
}

func (a *AgentSimulator) runDetailedTelemetryLoop(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.TelemetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.closed:
			return
		case <-ticker.C:
			start := time.Now()
			payload := []api.MetricDetailRequest{a.gen.NextDetail(a.device.DeviceID)}
			if _, err := a.client.PostDetailedMetrics(ctx, a.company.APIToken, payload); err != nil {
				continue
			}
			a.metrics.RecordTelemetry(time.Since(start))
			a.metrics.IncTelemetryMessagesSent(int64(len(payload)))
		}
	}
}

func (a *AgentSimulator) runCommandLoop(ctx context.Context) {
	if !a.cfg.RemoteCommands {
		<-ctx.Done()
		return
	}
	backoff := 500 * time.Millisecond
	for {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-a.closed:
			return
		default:
		}

		conn, err := stomp.Dial(ctx, a.cfg.WebSocketURL, map[string]string{"x-agent-token": a.company.APIToken})
		if err != nil {
			time.Sleep(backoff)
			backoff = utils.Clamp(backoff*2, 500*time.Millisecond, 15*time.Second)
			continue
		}
		backoff = 500 * time.Millisecond

		a.mu.Lock()
		a.commandConn = conn
		a.mu.Unlock()

		if _, err := conn.Subscribe("/topic/agent/" + a.device.DeviceID); err != nil {
			_ = conn.Close()
			time.Sleep(backoff)
			continue
		}

		err = conn.ReadLoop(ctx, func(frame stomp.Frame) {
			if frame.Command != "MESSAGE" || frame.Body == "" {
				return
			}
			a.metrics.IncCommandsReceived()
			var command api.CommandRequest
			if json.Unmarshal([]byte(frame.Body), &command) != nil {
				a.metrics.IncCommandsFailed()
				return
			}
			results := a.gen.ExecuteCommand(a.device.DeviceID, command)
			for _, result := range results {
				data, err := json.Marshal(result)
				if err != nil {
					a.metrics.IncCommandsFailed()
					continue
				}
				if sendErr := conn.Send("/app/command-result", data, map[string]string{"content-type": "application/json"}); sendErr != nil {
					a.metrics.IncCommandsFailed()
					continue
				}
				a.metrics.IncCommandsCompleted()
			}
		})
		_ = conn.Close()
		if err != nil {
			time.Sleep(backoff)
			continue
		}
		return
	}
}

func (a *AgentSimulator) markConnected(value bool) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	previous := a.connected
	a.connected = value
	return previous
}

func (a *AgentSimulator) closeConnections() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.metricsConn != nil {
		_ = a.metricsConn.Close()
		a.metricsConn = nil
	}
	if a.commandConn != nil {
		_ = a.commandConn.Close()
		a.commandConn = nil
	}
	a.connected = false
}
