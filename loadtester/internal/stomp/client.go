package stomp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	loadTesterHeader    = "X-Load-Tester"
	loadTesterUserAgent = "monitor-loadtester/1.0"
)

type Frame struct {
	Command string
	Headers map[string]string
	Body    string
}

type Client struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	closed   atomic.Bool
	subCount atomic.Int64
}

func Dial(ctx context.Context, rawURL string, headers map[string]string) (*Client, error) {
	wsURL := normalizeWebSocketURL(rawURL)
	handshakeHeaders := http.Header{}
	handshakeHeaders.Set("User-Agent", loadTesterUserAgent)
	handshakeHeaders.Set(loadTesterHeader, "true")
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, handshakeHeaders)
	if err != nil {
		return nil, err
	}

	client := &Client{conn: conn}
	conn.SetReadLimit(8 * 1024 * 1024)
	if err := client.send(Frame{Command: "CONNECT", Headers: connectHeaders(wsURL, headers)}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	if err := client.waitForConnected(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = conn.SetReadDeadline(time.Time{})
	return client, nil
}

func (c *Client) Subscribe(destination string) (string, error) {
	id := fmt.Sprintf("sub-%d", c.subCount.Add(1))
	return id, c.send(Frame{Command: "SUBSCRIBE", Headers: map[string]string{"id": id, "destination": destination}})
}

func (c *Client) Send(destination string, body []byte, extraHeaders map[string]string) error {
	headers := map[string]string{"destination": destination}
	for key, value := range extraHeaders {
		headers[key] = value
	}
	return c.send(Frame{Command: "SEND", Headers: headers, Body: string(body)})
}

func (c *Client) ReadLoop(ctx context.Context, handler func(Frame)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		for _, frame := range parseFrames(string(payload)) {
			if handler != nil {
				handler(frame)
			}
		}
	}
}

func (c *Client) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) send(frame Frame) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed.Load() {
		return fmt.Errorf("stomp client closed")
	}

	var builder strings.Builder
	builder.WriteString(frame.Command)
	builder.WriteByte('\n')
	for key, value := range frame.Headers {
		builder.WriteString(key)
		builder.WriteByte(':')
		builder.WriteString(value)
		builder.WriteByte('\n')
	}
	builder.WriteByte('\n')
	builder.WriteString(frame.Body)
	builder.WriteByte(0)

	return c.conn.WriteMessage(websocket.TextMessage, []byte(builder.String()))
}

func (c *Client) waitForConnected() error {
	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		for _, frame := range parseFrames(string(payload)) {
			switch frame.Command {
			case "CONNECTED":
				return nil
			case "ERROR":
				return fmt.Errorf("stomp error: %s", frame.Body)
			}
		}
	}
}

func parseFrames(payload string) []Frame {
	parts := bytes.Split([]byte(payload), []byte{0})
	frames := make([]Frame, 0, len(parts))
	for _, part := range parts {
		frame, ok := parseFrame(string(part))
		if ok {
			frames = append(frames, frame)
		}
	}
	return frames
}

func parseFrame(payload string) (Frame, bool) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return Frame{}, false
	}
	segments := strings.SplitN(payload, "\n\n", 2)
	if len(segments) == 0 {
		return Frame{}, false
	}
	headers := strings.Split(segments[0], "\n")
	if len(headers) == 0 {
		return Frame{}, false
	}
	frame := Frame{Command: strings.TrimSpace(headers[0]), Headers: map[string]string{}}
	for _, header := range headers[1:] {
		if idx := strings.Index(header, ":"); idx > 0 {
			frame.Headers[strings.TrimSpace(header[:idx])] = strings.TrimSpace(header[idx+1:])
		}
	}
	if len(segments) > 1 {
		frame.Body = strings.TrimRight(segments[1], "\x00")
	}
	return frame, true
}

func normalizeWebSocketURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if parsed.Scheme == "http" {
		parsed.Scheme = "ws"
	}
	if parsed.Scheme == "https" {
		parsed.Scheme = "wss"
	}
	return parsed.String()
}

func connectHeaders(rawURL string, headers map[string]string) map[string]string {
	merged := map[string]string{
		"accept-version": "1.2,1.1,1.0",
		"heart-beat":     "0,0",
	}

	if parsed, err := url.Parse(rawURL); err == nil && parsed.Host != "" {
		merged["host"] = parsed.Host
	}

	for key, value := range headers {
		merged[key] = value
	}
	return merged
}
