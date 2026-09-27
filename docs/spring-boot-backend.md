# Spring Boot Backend Forensics & Internal Mechanics

## 1. Application Architecture & Configuration

The backend is built with Spring Boot 4.0.2 running on Java 21, located in `backend/`.

### Maven Configuration (`pom.xml`)
- **Parent**: `spring-boot-starter-parent:4.0.2`
- **Primary Dependencies**:
  - `spring-boot-starter-webmvc`: Embedded Tomcat, REST controllers.
  - `spring-boot-starter-websocket`: STOMP message broker support.
  - `spring-boot-starter-security`: Authentication filter chain, BCrypt encoder.
  - `spring-boot-starter-data-jpa`: Hibernate ORM, repository abstraction.
  - `spring-boot-starter-validation`: Jakarta Bean Validation (`@Valid`, `@NotBlank`, `@Min`, `@Max`).
  - `spring-boot-starter-actuator`: Spring Boot health and metrics endpoints.
  - `org.postgresql:postgresql`: PostgreSQL JDBC driver.
  - `io.jsonwebtoken:jjwt-api:0.11.5`: JWT parsing and signing.
  - `org.projectlombok:lombok`: Code generation (`@Data`, `@Builder`, `@RequiredArgsConstructor`).

---

## 2. Ingress & Filter Pipeline

The Spring Security filter chain is configured in `SecurityConfig.java`:

```text
HTTP Request
     │
     ▼
[RateLimitFilter] (OncePerRequestFilter)
     │ - Checks if path in [/agent/**, /ws/**] -> If true, PASS THROUGH (NO LIMIT)
     │ - Checks if "X-Load-Tester: true" or User-Agent "monitor-loadtester/*" -> PASS THROUGH
     │ - Otherwise: key = (IP + "|" + URI), window = 60s, max = 120 reqs
     │ - If count > 120 -> HTTP 429 Too Many Requests
     ▼
[JwtFilter] (OncePerRequestFilter)
     │ - Checks if path in [/auth/**, /agent/**, /ws/**] -> If true, PASS THROUGH
     │ - Checks Header: "Authorization: Bearer <token>"
     │ - Validates HMAC SHA-256 signature and expiration
     │ - Extracts companyId (UUID) from Subject claim
     │ - Binds UsernamePasswordAuthenticationToken(companyId, null, List.of())
     │   to SecurityContextHolder
     ▼
[UsernamePasswordAuthenticationFilter]
     ▼
[Spring Security Authorization Manager]
     │ - OPTIONS /** -> permitAll()
     │ - /auth/**    -> permitAll()
     │ - /agent/**   -> permitAll()
     │ - /ws/**      -> permitAll()
     │ - anyRequest()-> authenticated()
     ▼
[DispatcherServlet] -> Target Controller
```

---

## 3. Controller Forensics

### 1. `AuthController` (`/auth`)
- **`POST /auth/register`**: Validates `RegisterRequest` (`name`, `email`, `password`). Checks email uniqueness (`companyRepository.findByEmail`). Generates random UUID `apiToken`, hashes password with `BCryptPasswordEncoder`, creates `Company` entity, and returns `AuthResponse` (`id`, `name`, `email`, `apiToken`).
- **`POST /auth/login`**: Validates `LoginRequest` (`email`, `password`). Verifies BCrypt hash. Invokes `jwtService.generateToken(company.getId())`. Returns `{"token": "<jwt>"}`.

### 2. `CompanyController` (`/company`)
- **`GET /company/me`**: Extracts `companyId` from `SecurityContextHolder`. Queries `Company` by ID and returns profile with `apiToken`.

### 3. `DeviceController` (`/devices`)
- **`GET /devices`**: Extracts `companyId` from `SecurityContextHolder`. Fetches all devices belonging to the company via `deviceRepository.findByCompany(company)`.
- **`GET /devices/{deviceId}/metrics`**: Fetches latest 50 metrics via `metricRepository.findTop50ByDeviceOrderByCreatedAtDesc(device)`.
  - ⚠️ *Security Finding*: Missing company validation. Any authenticated tenant can read metrics for any device if they know its UUID.
- **`GET /devices/{deviceId}/metrics-detail`**: Validates `device.getCompany().getId().equals(companyId)`. Returns cached details if cached within 5000ms (`DETAIL_CACHE_TTL_MS`), otherwise queries `metricDetailRepository.findTop20ByDeviceOrderByCreatedAtDesc(device)`.

### 4. `AgentController` (`/agent`)
- **`POST /agent/register`**: Registers a device using `AgentRegisterRequest` (`token`, `hostname`, `ipAddress`, `os`). Upserts device record (finds by hostname + company, or creates new). Returns `DeviceResponse`.
- **`POST /agent/metrics`**: Ingests single metric with `x-agent-token` header validation.
- **`POST /agent/metrics/batch`**: Ingests list of metrics with `x-agent-token` validation.
- **`POST /agent/metrics-detail`**: Ingests single detail snapshot with `x-agent-token` validation.
- **`POST /agent/metrics-detail/batch`**: Ingests list of detail snapshots with `x-agent-token` validation.

---

## 4. WebSocket & STOMP Architecture

Configured in `WebSocketConfig.java`:
- **Endpoint**: `/ws` (with optional SockJS fallback).
- **Prefixes**:
  - Inbound application mapping: `/app`.
  - Simple broker destination: `/topic`.
- **Transport Limits**:
  - Message size limit: 4MB (`4 * 1024 * 1024` bytes).
  - Send buffer size limit: 4MB.
  - Send time limit: 30,000ms.

### STOMP Channel Interceptor (`WebSocketAuthInterceptor.java`)
Interceps frames on `clientInboundChannel`:
1. On `StompCommand.CONNECT`:
   - Checks `Authorization: Bearer <jwt>`. If present, validates and extracts `companyId`.
   - Checks `x-agent-token: <token>`. If present, queries `companyRepository.findByApiToken(token)` to resolve `companyId`.
   - If neither resolves, throws `IllegalArgumentException("Unauthorized")`.
   - Attaches `UsernamePasswordAuthenticationToken(companyId, ...)` to `accessor.setUser(...)`.
2. On `SEND` / `SUBSCRIBE`: Passes through without subscription authorization checks.

### WebSocket Controllers
1. **`AgentWebSocketController`**:
   - `@MessageMapping("/agent/metrics")`: Receives single metric, validates company ownership, calls `agentService.saveMetric()`.
   - `@MessageMapping("/agent/metrics-batch")`: Validates that all items in batch belong to the same device ID, verifies tenant ownership, and calls `agentService.saveMetricsBatch()`.
   - `@MessageMapping("/agent/metrics-detail")`: Receives detail snapshot, validates company ownership, calls `agentService.saveMetricDetail()`.
   - `@MessageMapping("/agent/metrics-detail-batch")`: Validates same device ID and ownership, calls `agentService.saveMetricDetailsBatch()`.
2. **`CommandWebSocketController`**:
   - `@MessageMapping("/command/{deviceId}")`: Verifies device ownership against caller's `companyId`. Relays payload to `/topic/agent/{deviceId}`.
   - `@MessageMapping("/command-result")`: Verifies device ownership against caller's `companyId`. Relays payload to `/topic/command-result/{deviceId}`.

---

## 5. Storage Engine & TimescaleDB Integration (`MetricsStorageService.java`)

On `ApplicationReadyEvent`, `MetricsStorageService` initializes:
1. Executes `CREATE EXTENSION IF NOT EXISTS timescaledb`.
2. Normalizes timestamps: alters `created_at` and `last_seen_at` columns across `company`, `device`, `metric`, and `metric_detail` to `TIMESTAMPTZ`.
3. Drops primary keys on `metric` and `metric_detail` to permit hypertable partitioning.
4. Creates hypertables:
   - `SELECT create_hypertable('metric', 'created_at', if_not_exists => TRUE, migrate_data => TRUE)`
   - `SELECT create_hypertable('metric_detail', 'created_at', if_not_exists => TRUE, migrate_data => TRUE)`
5. Creates compound composite indexes:
   - `idx_metric_device_created_at ON metric (device_id, created_at DESC)`
   - `idx_metric_detail_device_created_at ON metric_detail (device_id, created_at DESC)`
6. Adds native TimescaleDB chunk retention policies:
   - `SELECT add_retention_policy('metric', INTERVAL '${retentionDays} days', if_not_exists => TRUE)`
   - `SELECT add_retention_policy('metric_detail', INTERVAL '${detailRetentionDays} days', if_not_exists => TRUE)`
7. **Relational Fallback**:
   - If TimescaleDB extension is unavailable, logs a warning and leaves `timescaleEnabled = false`.
   - A `@Scheduled(cron = "0 30 2 * * *")` fallback deletes records exceeding retention thresholds using SQL `DELETE`.

---

## 6. Device Status Scheduler (`DeviceStatusScheduler.java`)

- **Schedule**: `@Scheduled(fixedRate = 30000)` (every 30 seconds).
- **Execution**:
  - Calls `deviceRepository.findAll()`.
  - For each device: if `device.getLastSeenAt().isBefore(Instant.now().minusSeconds(30))` and status != `OFFLINE`:
    - Updates `device.setStatus(DeviceStatus.OFFLINE)`.
    - Persists update via `deviceRepository.save(device)`.
    - Broadcasts to `/topic/device-status/{deviceId}` with `DeviceStatus.OFFLINE`.\n