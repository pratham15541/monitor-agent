# Complete API & Protocol Reference

## 1. REST API Reference

All REST endpoints accept and return `application/json` unless otherwise noted.

---

### Authentication Endpoints (`/auth`)

#### `POST /auth/register`
- **Controller**: `AuthController.register()`
- **Auth**: Public (`permitAll()`)
- **Rate Limited**: Yes (120 req / 60s per IP+URI)
- **Request Body**:
  ```json
  {
    "name": "Acme Corp",
    "email": "ops@acme.com",
    "password": "secretpassword"
  }
  ```
- **Validation**: `@NotBlank` on name, email, password.
- **Success Response (200 OK)**:
  ```json
  {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "name": "Acme Corp",
    "email": "ops@acme.com",
    "apiToken": "7b89d44e-128a-4d43-982d-114498ec5123"
  }
  ```
- **Errors**:
  - `400 Bad Request`: Validation failure on missing fields.
  - `409 Conflict`: Email already registered.

#### `POST /auth/login`
- **Controller**: `AuthController.login()`
- **Auth**: Public (`permitAll()`)
- **Rate Limited**: Yes
- **Request Body**:
  ```json
  {
    "email": "ops@acme.com",
    "password": "secretpassword"
  }
  ```
- **Success Response (200 OK)**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```
- **Errors**:
  - `401 Unauthorized`: Bad credentials.

---

### Company Profile (`/company`)

#### `GET /company/me`
- **Controller**: `CompanyController.getProfile()`
- **Auth**: Requires `Authorization: Bearer <jwt>`
- **Rate Limited**: Yes
- **Success Response (200 OK)**:
  ```json
  {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "name": "Acme Corp",
    "email": "ops@acme.com",
    "apiToken": "7b89d44e-128a-4d43-982d-114498ec5123"
  }
  ```

---

### Device Management (`/devices`)

#### `GET /devices`
- **Controller**: `DeviceController.getDevices()`
- **Auth**: Requires `Authorization: Bearer <jwt>`
- **Rate Limited**: Yes
- **Success Response (200 OK)**:
  ```json
  [
    {
      "id": "9c3e2182-3d8b-4977-bc6d-0e42a98f411b",
      "hostname": "prod-web-01",
      "ipAddress": "192.168.1.10",
      "os": "linux",
      "status": "ONLINE",
      "lastSeenAt": "2026-09-27T12:00:00Z"
    }
  ]
  ```

#### `GET /devices/{deviceId}/metrics`
- **Controller**: `DeviceController.getMetrics()`
- **Auth**: Requires `Authorization: Bearer <jwt>`
- **Rate Limited**: Yes
- **Path Param**: `deviceId` (UUID)
- **Success Response (200 OK)**: Returns latest 50 metrics ordered descending by `createdAt`.
  ```json
  [
    {
      "id": 1042,
      "cpuUsage": 18.5,
      "memoryUsage": 62.1,
      "diskUsage": 45.0,
      "networkIn": 5421098,
      "networkOut": 8904321,
      "createdAt": "2026-09-27T12:00:00Z"
    }
  ]
  ```

#### `GET /devices/{deviceId}/metrics-detail`
- **Controller**: `DeviceController.getDetailedMetrics()`
- **Auth**: Requires `Authorization: Bearer <jwt>`
- **Rate Limited**: Yes
- **Path Param**: `deviceId` (UUID)
- **Success Response (200 OK)**: Returns latest 20 detailed snapshots (cached 5s in memory).
  ```json
  [
    {
      "id": 42,
      "detailsJson": "{\"processes\": [...], \"connections\": [...]}",
      "createdAt": "2026-09-27T12:00:00Z"
    }
  ]
  ```

---

### Agent Ingestion Endpoints (`/agent`)

#### `POST /agent/register`
- **Controller**: `AgentController.register()`
- **Auth**: Public endpoint, authenticated via body `token`
- **Rate Limited**: Exempted
- **Request Body**:
  ```json
  {
    "token": "7b89d44e-128a-4d43-982d-114498ec5123",
    "hostname": "agent-host",
    "ipAddress": "auto",
    "os": "linux"
  }
  ```
- **Success Response (200 OK)**:
  ```json
  {
    "id": "9c3e2182-3d8b-4977-bc6d-0e42a98f411b",
    "hostname": "agent-host",
    "ipAddress": "auto",
    "os": "linux",
    "status": "ONLINE",
    "lastSeenAt": "2026-09-27T12:00:00Z"
  }
  ```

#### `POST /agent/metrics`
- **Controller**: `AgentController.sendMetrics()`
- **Headers**: `x-agent-token: <company-api-token>`
- **Request Body**:
  ```json
  {
    "deviceId": "9c3e2182-3d8b-4977-bc6d-0e42a98f411b",
    "cpuUsage": 22.4,
    "memoryUsage": 48.1,
    "diskUsage": 38.9,
    "networkIn": 10240,
    "networkOut": 20480
  }
  ```
- **Success Response (200 OK)**: Empty body.

#### `POST /agent/metrics/batch`
- **Controller**: `AgentController.sendMetricsBatch()`
- **Headers**: `x-agent-token: <company-api-token>`
- **Request Body**: Array of `MetricRequest` objects.
- **Success Response (200 OK)**: Empty body.

#### `POST /agent/metrics-detail`
- **Controller**: `AgentController.sendMetricDetails()`
- **Headers**: `x-agent-token: <company-api-token>`
- **Request Body**:
  ```json
  {
    "deviceId": "9c3e2182-3d8b-4977-bc6d-0e42a98f411b",
    "details": { ... }
  }
  ```
- **Success Response (200 OK)**: Empty body.

#### `POST /agent/metrics-detail/batch`
- **Controller**: `AgentController.sendMetricDetailsBatch()`
- **Headers**: `x-agent-token: <company-api-token>`
- **Request Body**: Array of `MetricDetailRequest` objects.
- **Success Response (200 OK)**: Empty body.

---

## 2. STOMP WebSocket Protocol Reference (`/ws`)

### Application Destinations (Client -> Server)

| Destination | Controller | Payload | Description |
| :--- | :--- | :--- | :--- |
| `/app/agent/metrics` | `AgentWebSocketController` | `MetricRequest` | Ingests a single real-time metric from Go agent. |
| `/app/agent/metrics-batch` | `AgentWebSocketController` | `List<MetricRequest>` | Ingests a batch of real-time metrics (default 10 items). |
| `/app/agent/metrics-detail` | `AgentWebSocketController` | `MetricDetailRequest` | Ingests a single detailed snapshot. |
| `/app/agent/metrics-detail-batch` | `AgentWebSocketController` | `List<MetricDetailRequest>` | Ingests a batch of detailed snapshots. |
| `/app/command/{deviceId}` | `CommandWebSocketController`| `CommandRequest` | Operator submits remote command (`shell`, `service`, etc.). |
| `/app/command-result` | `CommandWebSocketController`| `CommandResult` | Agent transmits chunked or complete command output. |

### Broker Topics (Server -> Client Broadcasts)

| Topic | Publisher | Subscribers | Payload | Description |
| :--- | :--- | :--- | :--- | :--- |
| `/topic/device/{deviceId}` | `AgentService` | Browser UI | `Metric` (latest item) | Real-time metric broadcast. |
| `/topic/device-status/{deviceId}` | `AgentService`, `DeviceStatusScheduler` | Browser UI | `DeviceStatus` (`ONLINE`/`OFFLINE`) | Real-time state transition broadcast. |
| `/topic/device-detail/{deviceId}` | `AgentService` | Browser UI | `MetricDetailResponse` | Fresh process/service snapshot event. |
| `/topic/command-result/{deviceId}` | `CommandWebSocketController` | Browser UI | `CommandResult` (chunked) | Real-time command stdout/stderr stream. |
| `/topic/agent/{deviceId}` | `CommandWebSocketController` | Go Agent | `CommandRequest` | Inbound command dispatch to target host. |\n