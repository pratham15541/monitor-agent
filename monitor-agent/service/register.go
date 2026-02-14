package service

import (
	"encoding/json"
	"fmt"
	"monitor-agent/config"
	"os"
	"runtime"
)

func RegisterIfNeeded(cfg *config.Config) error {

	if cfg.DeviceID != "" {
		return nil
	}

	payload := map[string]string{
		"token":     cfg.Token,
		"hostname":  getHostname(),
		"ipAddress": "auto",
		"os":        getOS(),
	}

	resp, err := postJSON(cfg.ServerURL+"/agent/register", payload)
	if err != nil {
		fmt.Println("Register error:", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("register failed with status %s", resp.Status)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("register decode failed: %w", err)
	}
	if result.ID == "" {
		return fmt.Errorf("register response missing id")
	}

	cfg.DeviceID = result.ID
	config.Save(cfg)

	fmt.Println("Registered device:", cfg.DeviceID)
	return nil
}

func getHostname() string {
	h, _ := os.Hostname()
	return h
}

func getOS() string {
	return runtime.GOOS
}
