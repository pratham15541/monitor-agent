package service

import (
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

type stompFrame struct {
	Command string
	Headers map[string]string
	Body    string
}

func waitForConnected(conn *websocket.Conn) error {
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		frames := parseStompFrames(string(payload))
		for _, frame := range frames {
			if frame.Command == "CONNECTED" {
				return nil
			}
			if frame.Command == "ERROR" {
				return fmt.Errorf("stomp error: %s", frame.Body)
			}
		}
	}
}

func parseStompFrames(payload string) []stompFrame {
	parts := strings.Split(payload, "\x00")
	frames := make([]stompFrame, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		frame, ok := parseStompFrame(part)
		if ok {
			frames = append(frames, frame)
		}
	}
	return frames
}

func parseStompFrame(payload string) (stompFrame, bool) {
	sections := strings.SplitN(payload, "\n\n", 2)
	if len(sections) == 0 {
		return stompFrame{}, false
	}

	headers := make(map[string]string)
	headerLines := strings.Split(sections[0], "\n")
	if len(headerLines) == 0 {
		return stompFrame{}, false
	}

	command := strings.TrimSpace(headerLines[0])
	for _, line := range headerLines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		pair := strings.SplitN(line, ":", 2)
		if len(pair) == 2 {
			headers[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
		}
	}

	body := ""
	if len(sections) == 2 {
		body = sections[1]
	}

	return stompFrame{Command: command, Headers: headers, Body: body}, true
}

func sendStompFrame(conn *websocket.Conn, frame stompFrame) error {
	var builder strings.Builder
	builder.WriteString(frame.Command)
	builder.WriteString("\n")
	for key, value := range frame.Headers {
		builder.WriteString(key)
		builder.WriteString(":")
		builder.WriteString(value)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
	builder.WriteString(frame.Body)
	builder.WriteByte(0)
	return conn.WriteMessage(websocket.TextMessage, []byte(builder.String()))
}

func toWebSocketURL(serverURL string) string {
	base := strings.TrimRight(serverURL, "/")
	if strings.HasPrefix(base, "https://") {
		return "wss://" + strings.TrimPrefix(base, "https://") + "/ws"
	}
	if strings.HasPrefix(base, "http://") {
		return "ws://" + strings.TrimPrefix(base, "http://") + "/ws"
	}
	if strings.HasPrefix(base, "ws://") || strings.HasPrefix(base, "wss://") {
		return base + "/ws"
	}
	return "ws://" + base + "/ws"
}
