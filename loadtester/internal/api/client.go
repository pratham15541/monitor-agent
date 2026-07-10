package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	LoadTesterHeader    = "X-Load-Tester"
	LoadTesterUserAgent = "monitor-loadtester/1.0"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: timeout}}
}

func (c *Client) RegisterCompany(ctx context.Context, req AuthRequest) (AuthResponse, int, error) {
	var resp AuthResponse
	status, err := c.doJSON(ctx, http.MethodPost, "/auth/register", req, nil, &resp)
	return resp, status, err
}

func (c *Client) LoginCompany(ctx context.Context, req LoginRequest) (string, int, error) {
	var resp LoginResponse
	status, err := c.doJSON(ctx, http.MethodPost, "/auth/login", req, nil, &resp)
	return resp.Token, status, err
}

func (c *Client) GetCompanyMe(ctx context.Context, jwt string) (AuthResponse, int, error) {
	var resp AuthResponse
	status, err := c.doJSON(ctx, http.MethodGet, "/company/me", nil, map[string]string{"Authorization": "Bearer " + jwt}, &resp)
	return resp, status, err
}

func (c *Client) RegisterDevice(ctx context.Context, apiToken string, req AgentRegisterRequest) (DeviceResponse, int, error) {
	var resp DeviceResponse
	status, err := c.doJSON(ctx, http.MethodPost, "/agent/register", req, map[string]string{"x-agent-token": apiToken}, &resp)
	return resp, status, err
}

func (c *Client) PostDetailedMetrics(ctx context.Context, apiToken string, payload []MetricDetailRequest) (int, error) {
	return c.doJSON(ctx, http.MethodPost, "/agent/metrics-detail/batch", payload, map[string]string{"x-agent-token": apiToken}, nil)
}

func (c *Client) GetActuatorMetric(ctx context.Context, metricName string) (ActuatorMetricResponse, int, error) {
	var resp ActuatorMetricResponse
	status, err := c.doJSON(ctx, http.MethodGet, "/actuator/metrics/"+url.PathEscape(metricName), nil, nil, &resp)
	return resp, status, err
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, headers map[string]string, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", LoadTesterUserAgent)
	req.Header.Set(LoadTesterHeader, "true")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if len(payload) == 0 {
			return resp.StatusCode, fmt.Errorf("%s %s returned %s", method, path, resp.Status)
		}
		return resp.StatusCode, fmt.Errorf("%s %s returned %s: %s", method, path, resp.Status, strings.TrimSpace(string(payload)))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, err
		}
	}
	return resp.StatusCode, nil
}
