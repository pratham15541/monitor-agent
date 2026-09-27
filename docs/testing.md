# Testing Analysis & Verification Strategy

## 1. Existing Test Inventory

### Backend Test Suite (`backend/src/test/`)
- **File**: `MontorToolApplicationTests.java`
- **Framework**: JUnit 5, `@SpringBootTest`.
- **Implementation**:
  ```java
  @SpringBootTest
  class MontorToolApplicationTests {
      @Test
      void contextLoads() {
      }
  }
  ```
- **Coverage**: Evaluates only whether the Spring ApplicationContext can be bootstrapped (requires an active PostgreSQL database connection to succeed).
- **Gaps**: **Zero** unit tests, controller mock tests, service integration tests, security authorization tests, or repository tests.

### Go Agent Test Suite (`monitor-agent/`)
- **Inventory**: **Zero** test files exist under `monitor-agent/` (no `*_test.go` files).
- **Gaps**:
  - No tests for Cobra CLI argument parsing or flag validation.
  - No unit tests for STOMP frame serialization or parsing (`parseStompFrames`).
  - No race detection tests (`go test -race`).
  - No unit tests for `CollectMetrics()`, `collectProcessMetrics()`, or command execution sanitization.

### Frontend Test Suite (`frontend/`)
- **Inventory**: **Zero** test files exist under `frontend/` (no `*.test.tsx` or `*.spec.ts` files).
- **Dependencies**: `package.json` contains no test runner (no Jest, Vitest, Playwright, or Cypress).
- **Linting**: ESLint configured via `eslint.config.mjs`.

---

## 2. Load Testing Suite (`loadtester/`)

The repository contains an independent, comprehensive load-testing tool in `loadtester/`.

```text
loadtester/
├── cmd/loadtester/main.go          # CLI entrypoint with flags: --config, --preflight-only, --skip-preflight
├── internal/
│   ├── config/config.go           # YAML configuration loader & validator
│   ├── api/client.go              # REST API client for company & device provisioning
│   ├── company/manager.go         # Company creation & authentication driver
│   ├── device/manager.go          # Concurrent device registration & token association
│   ├── telemetry/generator.go     # Synthetic metrics & detail snapshot payload generator
│   ├── stomp/client.go            # Custom concurrent STOMP WebSocket client
│   ├── metrics/metrics.go         # Collector tracking throughput, errors, p95 latencies
│   ├── load/runner.go             # Master coordinator executing ramp-up, steady-state, chaos, and teardown
│   └── report/report.go           # Generates markdown, CSV, and JSON benchmark summaries
└── loadtest.example.yml           # Reference load testing profile
```

### Key Capabilities of the Load Tester
1. **Preflight Health Checks**: Verifies backend availability, database connectivity, and WebSocket upgrade support prior to launching load routines.
2. **Deterministic Ramp-Up**: Gradually spins up virtual agents according to a configured ramp-up profile (`rampSchedule`), measuring incremental backend degradation.
3. **Chaos Injection**: Periodically terminates virtual agent WebSocket connections or injects packet stalls (`chaosLoop`) to evaluate backend session recovery.
4. **Command Latency Benchmarking**: Emits remote commands and measures round-trip time from `/app/command/{id}` to `/topic/command-result/{id}`.

---

## 3. Recommended Automated Test Plan

### Priority 1: Go Agent Race & Concurrency Tests
- Implement `metrics_ws_test.go` with mock WebSocket servers (`httptest.NewServer`) testing:
  - Concurrent `RegisterIfNeeded` execution with `go test -race ./...`.
  - Batching threshold triggers (10 items vs 5s timer).
  - Socket disconnect during batch send and verification of reconnection backoff.

### Priority 2: Spring Boot Integration Tests
- Implement `@WebMvcTest` and `@DataJpaTest` with Testcontainers (PostgreSQL + TimescaleDB):
  - Multi-tenant data isolation: verify Company A user receives 403 or empty array when requesting Company B device metrics.
  - WebSocket subscription authorization: verify Company A client cannot subscribe to Company B device topic.
  - RateLimitFilter: verify 121st request from single IP receives HTTP 429.

### Priority 3: Frontend Component & E2E Tests
- Setup Vitest + React Testing Library:
  - Verify chunked command reassembly: pass simulated out-of-order chunks and verify assembled text output.
  - Verify auto-refresh toggles and interval clearing.\n