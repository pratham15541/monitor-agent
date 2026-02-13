# Monitor Tool

A full-stack monitoring platform with real-time WebSocket streaming, device registration, and metrics collection.

## Architecture

- **Backend**: Spring Boot 4.0 with STOMP/WebSocket, PostgreSQL, JWT auth
- **Frontend**: Next.js 16 with React 19, TailwindCSS, ShadcnUI components
- **Agent**: Go CLI that registers devices and streams metrics via HTTP

```
monitor-tool/
├── backend/        # Spring Boot REST + STOMP WebSocket API
├── frontend/       # Next.js dashboard
└── monitor-agent/  # Go agent for metric collection
```

## Quick Start

### 1. Backend

```bash
cd backend
./run.sh          # Linux/Mac
mvn spring-boot:run  # Windows
```

Runs on http://localhost:8080 (requires PostgreSQL)

### 2. Frontend

```bash
cd frontend
bun install
bun run dev
```

Runs on http://localhost:3000

### 3. Agent

```bash
cd monitor-agent
go build -o monitor-agent ./
./monitor-agent run
```

Register with: `./monitor-agent install --token YOUR_TOKEN --server http://localhost:8080`

## Docker

```bash
cd backend
docker compose up --build
```

Starts PostgreSQL + Backend on port 8080.

## Key Features

- **Live Metrics Streaming**: STOMP over SockJS for real-time CPU, memory, disk, network data
- **Device Status**: Auto-detect offline devices on a 30-second schedule
- **Multi-tenant**: Company API tokens for isolation
- **JWT Auth**: Token-based authentication for all endpoints
- **Responsive UI**: Tailored for desktop and tablet viewing

## Environment Variables

### Frontend

```env
NEXT_PUBLIC_API_BASE=http://localhost:8080
NEXT_PUBLIC_WS_URL=http://localhost:8080/ws
```

### Backend (Docker)

```env
SPRING_PROFILES_ACTIVE=docker
POSTGRES_DB=monitor
POSTGRES_USER=monitor
POSTGRES_PASSWORD=monitor
```

See [backend/README.md](backend/README.md), [frontend/README.md](frontend/README.md), and [monitor-agent/README.md](monitor-agent/README.md) for detailed docs.

## Development

- **Java**: 21+
- **Node**: 18+ / Bun 1.0+
- **Go**: 1.22+

## License

See individual project files for details.
