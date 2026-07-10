# Load Tester

Production-style end-to-end load testing framework for the monitoring platform.

It reuses the existing backend contracts:

- `POST /auth/register`
- `POST /auth/login`
- `GET /company/me`
- `POST /agent/register`
- `POST /agent/metrics-detail/batch`
- STOMP `/ws` with `x-agent-token` on CONNECT for agents
- STOMP `/ws` with `Authorization: Bearer <jwt>` for optional command emission

## Run

```bash
cd loadtester
go run ./cmd/loadtester -config loadtest.example.yml
```

While the run is active, the terminal prints each generated company record as soon as it is created, including the company ID, login email, password, API token, and JWT token. It also prints a live progress line every 10 seconds with the current phase, created companies, seeded systems, registered agents, connected agents, heartbeats, telemetry, and reconnects.

Use the printed login email and password to sign in to the dashboard you start manually, then open the corresponding company/device data while the load test is running.

The company ID comes from the backend responses (`POST /auth/register` and `GET /company/me`), which is the same ID the UI shows after creation.

Run only the preflight checks:

```bash
cd loadtester
go run ./cmd/loadtester -config loadtest.example.yml -preflight-only
```

Skip preflight if you need to go straight into the load run:

```bash
cd loadtester
go run ./cmd/loadtester -config loadtest.example.yml -skip-preflight
```

## Outputs

The run writes:

- `reports/load-test-report.json`
- `reports/load-test-report.md`
- `reports/load-test-report.csv`

## Example config

See [loadtest.example.yml](loadtest.example.yml).

## Notes

- Company onboarding uses the real backend registration flow.
- Simulated agents do not execute local shell commands.
- Dashboard-side command emission is optional and uses the backend STOMP contract.
