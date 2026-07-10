package load

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"loadtester/internal/api"
	"loadtester/internal/device"
	"loadtester/internal/stomp"
	"loadtester/internal/telemetry"
)

type PreflightReport struct {
	BackendReachable     bool          `json:"backendReachable"`
	AuthRegistrationOK   bool          `json:"authRegistrationOk"`
	AuthLoginOK          bool          `json:"authLoginOk"`
	AuthProfileOK        bool          `json:"authProfileOk"`
	AgentRegisterOK      bool          `json:"agentRegisterOk"`
	AgentWSConnected     bool          `json:"agentWsConnected"`
	DashboardWSConnected bool          `json:"dashboardWsConnected"`
	MetricsRoundTripOK   bool          `json:"metricsRoundTripOk"`
	CommandRoundTripOK   bool          `json:"commandRoundTripOk"`
	Duration             time.Duration `json:"duration"`
	CompanyName          string        `json:"companyName"`
	DeviceHostname       string        `json:"deviceHostname"`
	Notes                []string      `json:"notes,omitempty"`
	Error                string        `json:"error,omitempty"`
}

func PrintPreflightReport(report PreflightReport) {
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("PREFLIGHT SUMMARY")
	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Printf("Backend Reachable: %t\n", report.BackendReachable)
	fmt.Printf("Auth Registration: %t\n", report.AuthRegistrationOK)
	fmt.Printf("Auth Login: %t\n", report.AuthLoginOK)
	fmt.Printf("Auth Profile: %t\n", report.AuthProfileOK)
	fmt.Printf("Agent Register: %t\n", report.AgentRegisterOK)
	fmt.Printf("Agent WS Connected: %t\n", report.AgentWSConnected)
	fmt.Printf("Dashboard WS Connected: %t\n", report.DashboardWSConnected)
	fmt.Printf("Metrics Round Trip: %t\n", report.MetricsRoundTripOK)
	fmt.Printf("Command Round Trip: %t\n", report.CommandRoundTripOK)
	fmt.Printf("Duration: %s\n", report.Duration)
	if len(report.Notes) > 0 {
		fmt.Println()
		fmt.Println("Notes")
		for _, note := range report.Notes {
			fmt.Println("-", note)
		}
	}
	fmt.Println()
}

func (r *Runner) Preflight(ctx context.Context) (PreflightReport, error) {
	started := time.Now()
	if len(r.cfg.Companies) == 0 {
		return PreflightReport{}, fmt.Errorf("no companies configured")
	}

	spec := r.cfg.Companies[0]
	tempName := spec.Name + " Preflight"
	tempEmail := fmt.Sprintf("preflight-%s-%d@test.local", strings.ToLower(strings.ReplaceAll(spec.Name, " ", "-")), time.Now().UnixNano())
	tempPassword := spec.Password
	if tempPassword == "" {
		tempPassword = "Test@123"
	}

	profile := telemetry.Profile{
		Hostname:     fmt.Sprintf("preflight-%s", uuid.NewString()[:8]),
		IPAddress:    "10.255.255.10",
		OS:           "Ubuntu 24.04",
		CPU:          "4-core",
		RAM:          "8 GB",
		Disk:         "128 GB SSD",
		Architecture: "amd64",
		MACAddress:   "02:00:00:00:00:01",
		Timezone:     "UTC",
		Version:      "preflight",
	}

	report := PreflightReport{CompanyName: tempName, DeviceHostname: profile.Hostname}
	backendProbe, _, err := r.api.RegisterCompany(ctx, api.AuthRequest{Name: tempName, Email: tempEmail, Password: tempPassword})
	if err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	report.BackendReachable = true
	report.AuthRegistrationOK = backendProbe.ID != ""

	jwt, _, err := r.api.LoginCompany(ctx, api.LoginRequest{Email: tempEmail, Password: tempPassword})
	if err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	report.AuthLoginOK = jwt != ""

	profileResp, _, err := r.api.GetCompanyMe(ctx, jwt)
	if err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	report.AuthProfileOK = profileResp.ID != uuid.Nil.String() && profileResp.Email == tempEmail

	deviceResp, _, err := r.api.RegisterDevice(ctx, profileResp.APIToken, api.AgentRegisterRequest{
		Token:     profileResp.APIToken,
		Hostname:  profile.Hostname,
		IPAddress: profile.IPAddress,
		OS:        profile.OS,
	})
	if err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	report.AgentRegisterOK = deviceResp.ID != ""

	deviceRecord := device.Record{
		CompanyID:        profileResp.ID,
		CompanyName:      profileResp.Name,
		DeviceID:         deviceResp.ID,
		Hostname:         profile.Hostname,
		IPAddress:        profile.IPAddress,
		OS:               profile.OS,
		Status:           deviceResp.Status,
		LastSeenAt:       deviceResp.LastSeenAt,
		TelemetryProfile: profile,
	}

	agentConn, err := stomp.Dial(ctx, r.cfg.WebSocketURL, map[string]string{"x-agent-token": profileResp.APIToken})
	if err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	report.AgentWSConnected = true
	defer agentConn.Close()

	dashboardConn, err := stomp.Dial(ctx, r.cfg.WebSocketURL, map[string]string{"Authorization": "Bearer " + jwt})
	if err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	report.DashboardWSConnected = true
	defer dashboardConn.Close()

	deviceTopic := "/topic/device/" + deviceRecord.DeviceID
	commandTopic := "/topic/command-result/" + deviceRecord.DeviceID
	if _, err := agentConn.Subscribe("/topic/agent/" + deviceRecord.DeviceID); err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	if _, err := dashboardConn.Subscribe(deviceTopic); err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	if _, err := dashboardConn.Subscribe(commandTopic); err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}

	metricPayload, _ := json.Marshal([]api.MetricRequest{telemetry.NewGenerator(time.Now().UnixNano(), profile).NextMetric(deviceRecord.DeviceID)})
	if err := agentConn.Send("/app/agent/metrics-batch", metricPayload, map[string]string{"content-type": "application/json"}); err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}

	metricCtx, metricCancel := context.WithTimeout(ctx, 10*time.Second)
	metricSeen := false
	if err := dashboardConn.ReadLoop(metricCtx, func(frame stomp.Frame) {
		if frame.Command == "MESSAGE" && frame.Headers["destination"] == deviceTopic && frame.Body != "" {
			metricSeen = true
			metricCancel()
		}
	}); err != nil && metricCtx.Err() == nil {
		metricCancel()
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	metricCancel()
	if !metricSeen {
		err := fmt.Errorf("timed out waiting for metric on %s", deviceTopic)
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	report.MetricsRoundTripOK = true

	commandID := uuid.NewString()
	commandRequest := api.CommandRequest{
		DeviceID:  deviceRecord.DeviceID,
		CommandID: commandID,
		Type:      "diagnostics",
		Payload:   "",
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	agentCommandErr := make(chan error, 1)
	go func() {
		gen := telemetry.NewGenerator(time.Now().UnixNano(), profile)
		err := agentConn.ReadLoop(ctxTimeout, func(frame stomp.Frame) {
			if frame.Command != "MESSAGE" || frame.Headers["destination"] != "/topic/agent/"+deviceRecord.DeviceID || frame.Body == "" {
				return
			}

			var command api.CommandRequest
			if json.Unmarshal([]byte(frame.Body), &command) != nil || command.CommandID != commandID {
				return
			}

			for _, result := range gen.ExecuteCommand(deviceRecord.DeviceID, command) {
				data, err := json.Marshal(result)
				if err != nil {
					agentCommandErr <- err
					return
				}
				if err := agentConn.Send("/app/command-result", data, map[string]string{"content-type": "application/json"}); err != nil {
					agentCommandErr <- err
					return
				}
			}
		})
		if err != nil && ctxTimeout.Err() == nil {
			agentCommandErr <- err
		}
	}()

	commandData, _ := json.Marshal(commandRequest)
	if err := dashboardConn.Send("/app/command/"+deviceRecord.DeviceID, commandData, map[string]string{"content-type": "application/json"}); err != nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}

	commandSeen := false
	if err := dashboardConn.ReadLoop(ctxTimeout, func(frame stomp.Frame) {
		if frame.Command != "MESSAGE" || frame.Body == "" {
			return
		}
		var result api.CommandResult
		if json.Unmarshal([]byte(frame.Body), &result) != nil {
			return
		}
		if result.CommandID == commandID {
			commandSeen = true
			cancel()
		}
	}); err != nil && ctxTimeout.Err() == nil {
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	select {
	case err := <-agentCommandErr:
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	default:
	}
	if !commandSeen {
		err := fmt.Errorf("timed out waiting for command result on %s", commandTopic)
		report.Duration = time.Since(started)
		report.Error = err.Error()
		return report, err
	}
	report.CommandRoundTripOK = commandSeen
	report.Duration = time.Since(started)
	return report, nil
}
