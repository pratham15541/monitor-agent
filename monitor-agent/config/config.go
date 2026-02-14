package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	ServerURL string `json:"serverUrl"`
	Token     string `json:"token"`
	DeviceID  string `json:"deviceId"`
}

func getConfigPath() string {
	if override := os.Getenv("MONITOR_AGENT_CONFIG"); override != "" {
		return override
	}

	if runtime.GOOS == "windows" {
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		return filepath.Join(programData, "MonitorAgent", "config.json")
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".monitor-agent.json")
}

func Load() (*Config, error) {
	path := getConfigPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{}, nil
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	return &cfg, err
}

func Save(cfg *Config) error {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	path := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Delete() error {
	path := getConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}
