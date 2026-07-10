package load

import (
	"context"
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"loadtester/internal/api"
	"loadtester/internal/device"
	"loadtester/internal/metrics"
	"loadtester/internal/stomp"
	"loadtester/internal/utils"
)

type commandDriver struct {
	client   *stomp.Client
	devices  []device.Record
	metrics  *metrics.Collector
	rng      *rand.Rand
	tracked  map[string]chan api.CommandResult
	mu       sync.Mutex
	closed   chan struct{}
	stopOnce sync.Once
}

func newCommandDriver(ctx context.Context, jwt string, devices []device.Record, websocketURL string, collector *metrics.Collector, sampleSize int) (*commandDriver, error) {
	if len(devices) == 0 || jwt == "" {
		return nil, nil
	}
	if sampleSize > 0 && len(devices) > sampleSize {
		devices = append([]device.Record(nil), devices[:sampleSize]...)
	}
	conn, err := stomp.Dial(ctx, websocketURL, map[string]string{"Authorization": "Bearer " + jwt})
	if err != nil {
		return nil, err
	}
	driver := &commandDriver{
		client:  conn,
		devices: devices,
		metrics: collector,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		tracked: map[string]chan api.CommandResult{},
		closed:  make(chan struct{}),
	}
	for _, deviceRecord := range devices {
		if _, err := conn.Subscribe("/topic/command-result/" + deviceRecord.DeviceID); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}
	go driver.readLoop(ctx)
	return driver, nil
}

func (d *commandDriver) readLoop(ctx context.Context) {
	_ = d.client.ReadLoop(ctx, func(frame stomp.Frame) {
		if frame.Command != "MESSAGE" || frame.Body == "" {
			return
		}
		var result api.CommandResult
		if err := json.Unmarshal([]byte(frame.Body), &result); err != nil {
			return
		}
		d.mu.Lock()
		ch, ok := d.tracked[result.CommandID]
		d.mu.Unlock()
		if ok {
			select {
			case ch <- result:
			default:
			}
		}
	})
}

func (d *commandDriver) Run(ctx context.Context, interval time.Duration) {
	if d == nil || len(d.devices) == 0 || interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.closed:
			return
		case <-ticker.C:
			d.issueCommand(ctx)
		}
	}
}

func (d *commandDriver) issueCommand(ctx context.Context) {
	deviceRecord := d.devices[d.rng.Intn(len(d.devices))]
	types := []string{"diagnostics", "collect-details", "shell", "service", "timeout", "failure"}
	commandType := types[d.rng.Intn(len(types))]
	payload := commandPayloadForType(commandType, d.rng)
	commandID := utils.ShortID("cmd")
	request := api.CommandRequest{
		DeviceID:  deviceRecord.DeviceID,
		CommandID: commandID,
		Type:      commandType,
		Payload:   payload,
	}
	responseCh := make(chan api.CommandResult, 1)
	d.mu.Lock()
	d.tracked[commandID] = responseCh
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		delete(d.tracked, commandID)
		d.mu.Unlock()
	}()
	data, _ := json.Marshal(request)
	if err := d.client.Send("/app/command/"+deviceRecord.DeviceID, data, map[string]string{"content-type": "application/json"}); err != nil {
		return
	}
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
		return
	case <-responseCh:
		return
	}
}

func (d *commandDriver) Stop() {
	d.stopOnce.Do(func() {
		close(d.closed)
		if d.client != nil {
			_ = d.client.Close()
		}
	})
}

func commandPayloadForType(commandType string, rng *rand.Rand) string {
	switch commandType {
	case "shell":
		return "echo simulated command"
	case "service":
		return []string{"start", "stop", "restart"}[rng.Intn(3)]
	case "timeout":
		return "sleep"
	case "failure":
		return "invalid"
	default:
		return ""
	}
}
