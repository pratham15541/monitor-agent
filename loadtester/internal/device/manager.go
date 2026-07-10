package device

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sort"
	"strings"
	"sync"

	"loadtester/internal/api"
	"loadtester/internal/company"
	"loadtester/internal/config"
	"loadtester/internal/telemetry"
)

type Record struct {
	CompanyID        string            `json:"companyId"`
	CompanyName      string            `json:"companyName"`
	DeviceID         string            `json:"deviceId"`
	Hostname         string            `json:"hostname"`
	IPAddress        string            `json:"ipAddress"`
	OS               string            `json:"os"`
	Status           string            `json:"status"`
	LastSeenAt       string            `json:"lastSeenAt,omitempty"`
	Architecture     string            `json:"architecture"`
	CPU              string            `json:"cpu"`
	RAM              string            `json:"ram"`
	Disk             string            `json:"disk"`
	MACAddress       string            `json:"macAddress"`
	Timezone         string            `json:"timezone"`
	Version          string            `json:"version"`
	TelemetryProfile telemetry.Profile `json:"-"`
}

type Manager struct {
	client *api.Client
	pool   int
}

func NewManager(client *api.Client, pool int) *Manager {
	if pool <= 0 {
		pool = 20
	}
	return &Manager{client: client, pool: pool}
}

func (m *Manager) SeedCompany(ctx context.Context, cfg config.Config, record company.Record) ([]Record, error) {
	total := record.Systems
	if total <= 0 {
		return nil, nil
	}

	tasks := make(chan int)
	results := make(chan Record, total)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	workers := m.pool
	if total < workers {
		workers = total
	}

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range tasks {
				profile := buildProfile(record, index)
				registered, _, err := m.client.RegisterDevice(ctx, record.APIToken, api.AgentRegisterRequest{
					Token:     record.APIToken,
					Hostname:  profile.Hostname,
					IPAddress: profile.IPAddress,
					OS:        profile.OS,
				})
				if err != nil {
					select {
					case errCh <- fmt.Errorf("register device %s: %w", profile.Hostname, err):
					default:
					}
					return
				}

				results <- Record{
					CompanyID:        record.ID,
					CompanyName:      record.Name,
					DeviceID:         registered.ID,
					Hostname:         profile.Hostname,
					IPAddress:        profile.IPAddress,
					OS:               profile.OS,
					Status:           registered.Status,
					LastSeenAt:       registered.LastSeenAt,
					Architecture:     profile.Architecture,
					CPU:              profile.CPU,
					RAM:              profile.RAM,
					Disk:             profile.Disk,
					MACAddress:       profile.MACAddress,
					Timezone:         profile.Timezone,
					Version:          profile.Version,
					TelemetryProfile: profile,
				}
			}
		}()
	}

	go func() {
		for index := 0; index < total; index++ {
			tasks <- index
		}
		close(tasks)
		wg.Wait()
		close(results)
		close(errCh)
	}()

	devices := make([]Record, 0, total)
	for {
		select {
		case err, ok := <-errCh:
			if ok && err != nil {
				return nil, err
			}
		case result, ok := <-results:
			if !ok {
				sort.Slice(devices, func(i, j int) bool { return devices[i].Hostname < devices[j].Hostname })
				return devices, nil
			}
			devices = append(devices, result)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func buildProfile(record company.Record, index int) telemetry.Profile {
	hostname := fmt.Sprintf("device-%06d", index+1)
	ip := fmt.Sprintf("10.%d.%d.%d", (index/256)%256, index%256, (index%200)+10)
	osChoices := []string{"Windows 11", "Ubuntu 24.04", "Windows Server 2022"}
	osName := osChoices[index%len(osChoices)]
	mac := net.HardwareAddr{0x02, byte(index >> 16), byte(index >> 8), byte(index)}.String()
	version := fmt.Sprintf("%s-%s", strings.ToLower(strings.ReplaceAll(record.Name, " ", "-")), runtime.GOARCH)
	return telemetry.Profile{
		Hostname:     hostname,
		IPAddress:    ip,
		OS:           osName,
		CPU:          fmt.Sprintf("%d-core", 4+(index%12)),
		RAM:          fmt.Sprintf("%d GB", 8+(index%56)),
		Disk:         fmt.Sprintf("%d GB SSD", 128+(index%2048)),
		Architecture: runtime.GOARCH,
		MACAddress:   mac,
		Timezone:     "UTC",
		Version:      version,
	}
}
