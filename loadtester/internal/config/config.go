package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type CompanyConfig struct {
	Name       string `yaml:"name"`
	Systems    int    `yaml:"systems"`
	AdminName  string `yaml:"admin_name"`
	AdminEmail string `yaml:"admin_email"`
	Password   string `yaml:"password"`
}

type Config struct {
	BackendURL                 string          `yaml:"backend_url"`
	WebSocketURL               string          `yaml:"ws_url"`
	Duration                   time.Duration   `yaml:"duration"`
	RampUp                     time.Duration   `yaml:"ramp_up"`
	HeartbeatInterval          time.Duration   `yaml:"heartbeat_interval"`
	TelemetryInterval          time.Duration   `yaml:"telemetry_interval"`
	CommandInterval            time.Duration   `yaml:"command_interval"`
	ChaosInterval              time.Duration   `yaml:"chaos_interval"`
	RandomDisconnectPercent    float64         `yaml:"random_disconnect_percent"`
	RemoteCommands             bool            `yaml:"remote_commands"`
	ReportDir                  string          `yaml:"report_dir"`
	MaxConcurrentRegistrations int             `yaml:"max_concurrent_registrations"`
	MaxConcurrentAgents        int             `yaml:"max_concurrent_agents"`
	CommandSampleSize          int             `yaml:"command_sample_size"`
	BackendMetricsInterval     time.Duration   `yaml:"backend_metrics_interval"`
	RequestTimeout             time.Duration   `yaml:"request_timeout"`
	Companies                  []CompanyConfig `yaml:"companies"`
}

func Load(path string) (Config, error) {
	cfg := defaultConfig()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse config: %w", err)
		}
	}
	cfg.Normalize()
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		BackendURL:                 "http://localhost:8080",
		Duration:                   20 * time.Minute,
		RampUp:                     6 * time.Minute,
		HeartbeatInterval:          30 * time.Second,
		TelemetryInterval:          5 * time.Second,
		CommandInterval:            30 * time.Second,
		ChaosInterval:              time.Minute,
		RandomDisconnectPercent:    5,
		RemoteCommands:             true,
		ReportDir:                  "reports",
		MaxConcurrentRegistrations: 50,
		MaxConcurrentAgents:        100,
		CommandSampleSize:          10,
		BackendMetricsInterval:     15 * time.Second,
		RequestTimeout:             15 * time.Second,
		Companies: []CompanyConfig{
			{Name: "Company A", Systems: 50},
			{Name: "Company B", Systems: 100},
			{Name: "Company C", Systems: 350},
		},
	}
}

func (c *Config) Normalize() {
	if strings.TrimSpace(c.BackendURL) == "" {
		c.BackendURL = "http://localhost:8080"
	}
	if strings.TrimSpace(c.WebSocketURL) == "" {
		c.WebSocketURL = deriveWebSocketURL(c.BackendURL)
	}
	if c.Duration <= 0 {
		c.Duration = 20 * time.Minute
	}
	if c.RampUp <= 0 {
		c.RampUp = 6 * time.Minute
	}
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = 30 * time.Second
	}
	if c.TelemetryInterval <= 0 {
		c.TelemetryInterval = 5 * time.Second
	}
	if c.CommandInterval <= 0 {
		c.CommandInterval = 30 * time.Second
	}
	if c.ChaosInterval <= 0 {
		c.ChaosInterval = time.Minute
	}
	if c.RandomDisconnectPercent <= 0 {
		c.RandomDisconnectPercent = 5
	}
	if c.ReportDir == "" {
		c.ReportDir = "reports"
	}
	if c.MaxConcurrentRegistrations <= 0 {
		c.MaxConcurrentRegistrations = 50
	}
	if c.MaxConcurrentAgents <= 0 {
		c.MaxConcurrentAgents = 100
	}
	if c.CommandSampleSize <= 0 {
		c.CommandSampleSize = 10
	}
	if c.BackendMetricsInterval <= 0 {
		c.BackendMetricsInterval = 15 * time.Second
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 15 * time.Second
	}
	for i := range c.Companies {
		if c.Companies[i].Systems <= 0 {
			c.Companies[i].Systems = 1
		}
		if strings.TrimSpace(c.Companies[i].Name) == "" {
			c.Companies[i].Name = fmt.Sprintf("Company %d", i+1)
		}
		if strings.TrimSpace(c.Companies[i].Password) == "" {
			c.Companies[i].Password = "Test@123"
		}
		if strings.TrimSpace(c.Companies[i].AdminName) == "" {
			c.Companies[i].AdminName = c.Companies[i].Name + " Admin"
		}
		if strings.TrimSpace(c.Companies[i].AdminEmail) == "" {
			c.Companies[i].AdminEmail = defaultEmail(c.Companies[i].Name)
		}
	}
}

func EnsureDirectory(path string) error {
	if path == "" {
		return nil
	}
	return os.MkdirAll(filepath.Clean(path), 0o755)
}

func deriveWebSocketURL(baseURL string) string {
	if strings.HasPrefix(baseURL, "https://") {
		return strings.TrimRight(strings.Replace(baseURL, "https://", "wss://", 1), "/") + "/ws"
	}
	return strings.TrimRight(strings.Replace(baseURL, "http://", "ws://", 1), "/") + "/ws"
}

func defaultEmail(name string) string {
	slug := strings.ToLower(strings.NewReplacer(" ", "", "/", "", "_", "", "-", "").Replace(name))
	return slug + "@test.local"
}
