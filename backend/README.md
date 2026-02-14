# Monitor Tool Backend

Spring Boot backend for the monitoring platform. Provides REST APIs, STOMP/WebSocket
streams, device status tracking, and command routing.

## Requirements

- Java 21
- Maven 3.9+
- PostgreSQL 16 (local or via Docker)

## Run locally

```bash
mvn spring-boot:run
```

Backend runs on http://localhost:8080 by default.

## Run with Docker

```bash
docker compose up --build
```

Starts PostgreSQL and the backend on port 8080.

## REST APIs

Auth

- POST /auth/register
- POST /auth/login

Devices

- GET /devices
- GET /devices/{deviceId}/metrics
- GET /devices/{deviceId}/metrics-detail

Agent

- POST /agent/register
- POST /agent/metrics
- POST /agent/metrics-detail

## WebSocket (STOMP)

- Endpoint: /ws (SockJS enabled)
- Topics:
  - /topic/device/{deviceId} (live metrics)
  - /topic/device-status/{deviceId} (ONLINE/OFFLINE)
  - /topic/command-result/{deviceId}
  - /topic/agent/{deviceId} (commands to agent)
- App destinations:
  - /app/agent/metrics
  - /app/agent/metrics-detail
  - /app/command/{deviceId}
  - /app/command-result

Authentication headers:

- UI: Authorization: Bearer <jwt>
- Agent: x-agent-token: <api token>

## Background Jobs

- Offline detection runs every 30 seconds and broadcasts status changes.

## Configuration

Default settings live in src/main/resources/application.yml and
src/main/resources/application.properties.

```env
JWT_SECRET=...
JWT_ISSUER=monitor-tool
JWT_EXP_MINUTES=60
CORS_ALLOWED_ORIGINS=http://localhost:3000
RATE_LIMIT_WINDOW=60
RATE_LIMIT_MAX=120
```
