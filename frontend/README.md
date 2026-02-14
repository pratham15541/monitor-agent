# Monitor Tool Frontend

Next.js dashboard for the monitoring platform. It provides authentication flows,
device inventory, live metrics, and deep-dive system insights.

## Features

- Company registration and login flows
- Dashboard with device search, auto-refresh, and status counts
- Device detail view with live metrics charts
- Remote commands with result history
- Detailed metrics for processes, connections, services, and logs
- Company profile screen for API token management

## Getting Started

```bash
bun install
bun dev
```

The app runs on http://localhost:3000

You can also use npm, yarn, or pnpm:

```bash
npm install
npm run dev
```

## Environment Variables

```env
NEXT_PUBLIC_API_BASE=http://localhost:8080
NEXT_PUBLIC_WS_URL=http://localhost:8080/ws
```

## Notes

- JWT tokens are stored in localStorage.
- WebSocket auth uses the Authorization header on STOMP CONNECT.
