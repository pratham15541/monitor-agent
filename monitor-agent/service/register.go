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
	fmt.Println("Register error:", err)
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	cfg.DeviceID = result["id"].(string)
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
