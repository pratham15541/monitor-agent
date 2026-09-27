# Next.js Frontend Forensics & Architecture

## 1. Application Layout & Tech Stack

The frontend is a modern web application located in `frontend/`, built with:
- **Framework**: Next.js 16.1.6 (App Router)
- **UI Runtime**: React 19.2.3, TypeScript 5, Tailwind CSS v4, Radix UI primitives, `lucide-react`
- **Real-Time Client**: `@stomp/stompjs` 7.1.0 and `sockjs-client` 1.6.1
- **HTTP Client**: Custom fetch wrapper (`lib/api.ts`) and `axios` 1.13.5

```text
frontend/
├── app/
│   ├── (auth)/
│   │   ├── login/page.tsx         # Company sign-in
│   │   └── register/page.tsx      # Company registration & API token issue
│   ├── (dashboard)/
│   │   ├── layout.tsx             # Shell wrapper, sidebar, authentication guard
│   │   ├── dashboard/page.tsx     # Fleet overview, search, status tiles
│   │   ├── devices/[deviceId]/
│   │   │   └── page.tsx           # Live metrics, charts, diagnostics, commands
│   │   ├── company/page.tsx       # Company profile & API token management
│   │   └── agents/page.tsx        # Manual agent registration form
│   ├── layout.tsx                 # Root layout & global stylesheet
│   └── globals.css                # Tailwind CSS imports & theme vars
├── components/
│   ├── app/
│   │   ├── AppShell.tsx           # Responsive layout structure
│   │   ├── Sidebar.tsx            # Navigation drawer
│   │   ├── TopBar.tsx             # Breadcrumbs & sign-out
│   │   ├── DeviceCard.tsx         # Grid card for device status
│   │   ├── MetricChart.tsx        # SVG-based live metric polyline charts
│   │   ├── ProcessListView.tsx    # Interactive process table with sorting
│   │   ├── ConnectionsView.tsx    # Open socket list (local/remote IP:port)
│   │   ├── ServicesView.tsx       # OS service list output viewer
│   │   ├── LogsView.tsx           # Dual-tab log viewer (agent + system)
│   │   └── StatTile.tsx           # Stat counter card
│   └── ui/                        # Radix-ui wrappers (Button, Card, Table, Sheet, etc.)
└── lib/
    ├── api.ts                     # fetchJson wrapper with automatic Bearer JWT injection
    ├── auth.ts                    # LocalStorage token & profile persistence
    ├── format.ts                  # Date/time, byte, and number formatters
    └── types.ts                   # Full TypeScript domain contracts
```

---

## 2. Authentication & Client Session Management (`lib/auth.ts` & `lib/api.ts`)

- **Token Storage**:
  - `monitor.jwt`: Stores the raw JWT string in `window.localStorage`.
  - `monitor.company`: Stores the JSON-serialized `CompanyProfile` (`id`, `name`, `email`, `apiToken`).
- **Request Interceptor (`lib/api.ts: fetchJson()`)**:
  - Automatically reads `getToken()`.
  - Injects `Authorization: Bearer <token>` into HTTP headers.
  - Automatically parses JSON response, or handles 204 No Content.
  - Throws typed errors containing backend error response text on non-2xx status codes.

---

## 3. Real-Time WebSocket/STOMP Engine (`devices/[deviceId]/page.tsx`)

When an operator navigates to `/devices/{deviceId}`:

1. **Initial Hydration**:
   Executes `Promise.all`:
   - `GET /devices`: Resolves hostname, OS, IP address, and status.
   - `GET /devices/{deviceId}/metrics`: Fetches initial 50 historical metrics.
   - `GET /devices/{deviceId}/metrics-detail`: Fetches latest 20 snapshots (cached locally in `detailedMetricsCache` for 10s).
2. **WebSocket Activation (`@stomp/stompjs`)**:
   - Factory initializes `new SockJS(normalizeSockJsUrl(process.env.NEXT_PUBLIC_WS_URL))`.
   - Passes `connectHeaders: { Authorization: "Bearer <token>" }`.
   - On successful STOMP connection, creates four independent topic subscriptions:
     - `/topic/device/{deviceId}`: Ingestion stream of live metrics. Appends to `metrics` state (capped at `MAX_METRICS = 60`).
     - `/topic/device-status/{deviceId}`: Ingestion stream of status transitions (`ONLINE` / `OFFLINE`).
     - `/topic/device-detail/{deviceId}`: Ingestion stream of deep diagnostic snapshots.
     - `/topic/command-result/{deviceId}`: Inbound stream of remote command output chunks.
3. **Automatic Reconnection & Teardown**:
   - `reconnectDelay: 5000` ms.
   - On component unmount (`useEffect` cleanup), calls `client.deactivate()` and nullifies the client ref.

---

## 4. Multi-Chunk Command Reassembly Engine

In `devices/[deviceId]/page.tsx`:

When the agent sends large command outputs (e.g. `systemctl list-units` or full directory trees), the agent chunks the output into <=12KB segments. The frontend reassembles these asynchronously:

```typescript
function upsertCommandResult(prev: CommandResult[], incoming: CommandResult, buffers: Map<string, CommandChunkBuffer>) {
  if (incoming.chunked && incoming.chunkType) {
    const buffer = buffers.get(incoming.commandId) ?? { outputParts: new Map(), errorParts: new Map() };
    if (incoming.chunkType === "output") {
      buffer.outputParts.set(incoming.chunkIndex ?? 0, incoming.output ?? "");
      if (incoming.chunkCount) buffer.outputCount = incoming.chunkCount;
    } else {
      buffer.errorParts.set(incoming.chunkIndex ?? 0, incoming.error ?? "");
      if (incoming.chunkCount) buffer.errorCount = incoming.chunkCount;
    }
    buffers.set(incoming.commandId, buffer);

    const assembledOutput = assembleChunks(buffer.outputParts, buffer.outputCount);
    const assembledError = assembleChunks(buffer.errorParts, buffer.errorCount);
    
    const merged: CommandResult = {
      ...incoming,
      output: assembledOutput || undefined,
      error: assembledError || undefined,
      status: incoming.finishedAt ? incoming.status : "stream",
    };

    const next = prev.map((item) => item.commandId === incoming.commandId ? { ...item, ...merged } : item);
    if (!next.some((item) => item.commandId === incoming.commandId)) {
      next.unshift(merged);
    }
    if (incoming.finishedAt) {
      buffers.delete(incoming.commandId);
    }
    return next.slice(0, 20);
  }
}
```

---

## 5. View Components & Diagnostic Visualizers

### 1. `MetricChart.tsx`
- Renders responsive pure SVG charts with dual polyline layers (e.g., CPU% and Memory%, or Network In and Network Out).
- Automatically calculates coordinate scales, min/max domains, grid lines, and tooltips without heavy charting libraries.

### 2. `ProcessListView.tsx`
- Renders process table: PID, Process Name, User, Status, CPU %, Memory RSS (formatted in MB/GB), Threads, and Full Commandline.
- Supports client-side sorting by CPU, Memory, or PID.

### 3. `ConnectionsView.tsx`
- Displays active sockets: PID, Protocol Family (IPv4/IPv6), Socket Type (TCP/UDP), Status (LISTEN, ESTABLISHED, TIME_WAIT), Local Address, and Remote Address.
- Includes a client-side search filter across IP, port, and PID.

### 4. `ServicesView.tsx` & `LogsView.tsx`
- `ServicesView`: Displays formatted service manager status text with copy-to-clipboard functionality.
- `LogsView`: Split tab displaying the tail of `agent.log` and the OS system event log (`journalctl` / `wevtutil`).\n