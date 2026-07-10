package api

type AuthRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	APIToken string `json:"apiToken"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type DeviceResponse struct {
	ID         string `json:"id"`
	Hostname   string `json:"hostname"`
	IPAddress  string `json:"ipAddress"`
	OS         string `json:"os"`
	Status     string `json:"status"`
	LastSeenAt string `json:"lastSeenAt"`
}

type AgentRegisterRequest struct {
	Token     string `json:"token"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ipAddress"`
	OS        string `json:"os"`
}

type MetricRequest struct {
	DeviceID    string  `json:"deviceId"`
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	DiskUsage   float64 `json:"diskUsage"`
	NetworkIn   float64 `json:"networkIn"`
	NetworkOut  float64 `json:"networkOut"`
}

type MetricDetailRequest struct {
	DeviceID string                 `json:"deviceId"`
	Details  map[string]interface{} `json:"details"`
}

type CommandRequest struct {
	DeviceID  string `json:"deviceId"`
	CommandID string `json:"commandId"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
}

type CommandResult struct {
	DeviceID   string `json:"deviceId"`
	CommandID  string `json:"commandId"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Output     string `json:"output"`
	Error      string `json:"error"`
	Chunked    bool   `json:"chunked"`
	ChunkType  string `json:"chunkType"`
	ChunkIndex int    `json:"chunkIndex"`
	ChunkCount int    `json:"chunkCount"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

type ActuatorMetricResponse struct {
	Name         string `json:"name"`
	Measurements []struct {
		Statistic string  `json:"statistic"`
		Value     float64 `json:"value"`
	} `json:"measurements"`
}
