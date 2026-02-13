# Monitor Tool Backend

Spring Boot backend for the monitoring platform. It exposes REST APIs, streams
metrics over STOMP/WebSocket, and marks devices offline on a schedule.

## Requirements

- Java 21
- Maven 3.9+
- PostgreSQL 16 (local or via Docker)

## Configuration

Default settings live in `src/main/resources/application.yml` and
`src/main/resources/application.properties`. The Docker profile uses the
`docker-compose.yml` service settings.

## Run locally

```bash
mvn spring-boot:run
```

Backend runs on `http://localhost:8080` by default.

## Run with Docker

```bash
docker compose up --build
```

This starts PostgreSQL and the backend on port 8080.

## WebSocket (STOMP)

- Endpoint: `/ws` (SockJS enabled)
- Topics:
  - `/topic/device/{deviceId}` for live metrics
  - `/topic/device-status/{deviceId}` for ONLINE/OFFLINE status

## Common APIs

- `GET /devices`
- `GET /devices/{deviceId}/metrics`
