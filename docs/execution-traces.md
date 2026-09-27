# End-to-End Execution Traces (Code-Forensic Verification)

This document provides concrete, code-verified execution traces through every critical operation in the **Monitor Agent** ecosystem. Each trace specifies exact file paths, line numbers, function calls, network frames, and database interactions.

---

## Trace 1: Metrics Collection, Batching, Transmission & Live UI Display

This trace shows the entire lifecycle of a real-time metric from the OS kernel counters to the rendered SVG chart on the Next.js frontend.

```text
[Operating System Kernel]
  │ (CPU counters, /proc/meminfo, root drive IO, socket counters)
  ▼
[Go Agent: service/metrics.go - CollectMetrics()]
  │ Calls gopsutil/v3/cpu.Percent(0, false)
  │ Calls gopsutil/v3/mem.VirtualMemory()
  │ Calls gopsutil/v3/disk.Usage("C:\\" or "/")
  │ Calls gopsutil/v3/net.IOCounters(false)
  ▼
[Go Agent: service/metrics_ws.go - runMetricsSession()]
  │ payload["deviceId"] = cfg.DeviceID
  │ batch = append(batch, payload)
  │ Condition Check: len(batch) >= 10 || time.Since(lastFlush) >= 5s
  ▼
[Go Agent: service/metrics_ws.go - sendMetricsBatch()]
  │ json.Marshal(batch)
  │ Constructs STOMP Frame:
  │   COMMAND: SEND
  │   destination: /app/agent/metrics-batch
  │   content-type: application/json
  │   Body: [{"deviceId":"...","cpuUsage":12.5,"memoryUsage":45.2,...}]
  ▼
[Network: WebSocket Connection 1 (Agent -> Backend)]
  │ STOMP frame transmitted over gorilla/websocket client connection
  ▼
[Backend: config/WebSocketConfig.java + config/WebSocketAuthInterceptor.java]
  │ Client inbound channel routes STOMP message
  │ Session principal pre-authenticated from CONNECT frame (UUID companyId)
  ▼
[Backend: controller/AgentWebSocketController.java - receiveMetricsBatch()]
  │ Destination: @MessageMapping("/agent/metrics-batch")
  │ Validates: requests != null && !requests.isEmpty()
  │ Validates: requests.get(0).getDeviceId() == all requests in batch
  │ Looks up: deviceRepository.findById(deviceId)
  │ Validates: device.getCompany().getId().equals(companyId)
  ▼
[Backend: service/AgentService.java - saveMetricsBatch()]
  │ @Transactional boundary begins
  │ Converts DTOs to Entities: buildMetrics(device, requests, Instant.now())
  │ NOTE: Every metric in the batch is assigned identical timestamp `Instant.now()`
  │ Persists: metricRepository.saveAll(metrics) -> INSERT INTO metric (...)
  │ Updates Device:
  │   device.setLastSeenAt(latestCreatedAt)
  │   if device.status != ONLINE -> device.setStatus(ONLINE)
  │ Persists Device: deviceRepository.save(device)
  │ Broadcasts Latest Metric:
  │   messagingTemplate.convertAndSend("/topic/device/" + deviceId, latestMetric)
  │ @Transactional commits
  ▼
[Backend: Spring SimpleBroker]
  │ Broadcasts STOMP MESSAGE frame to subscribers of "/topic/device/{deviceId}"
  ▼
[Network: WebSocket Connection 2 (Backend -> Browser SockJS)]
  │ STOMP frame transmitted over SockJS WebSocket connection
  ▼
[Frontend: app/(dashboard)/devices/[deviceId]/page.tsx - client.subscribe()]
  │ Handler receives message.body:
  │ JSON.parse(message.body) as Metric
  │ State update: setMetrics((prev) => [parsed, ...prev].slice(0, 60))
  │ State update: setDevice((current) => current ? { ...current, lastSeenAt: parsed.createdAt } : current)
  ▼
[Frontend: components/app/MetricChart.tsx]
  │ React re-renders with new metrics array
  │ Recomputes chartMetrics (reversed chronological order)
  │ SVG polylines and StatTiles update in real-time
```

---

## Trace 2: Detailed System Snapshot Pipeline (Process, Sockets, Services, Logs)

This trace details the collection and ingestion of heavy diagnostics snapshots.

```text
[OS Subsystems]
  │ Process table (PID, CPU%, RSS, threads, cmdline)
  │ Socket table (all listening and established TCP/UDP sockets)
  │ Memory breakdown (cached, buffers, swap, slab, active/inactive)
  │ Service manager: `sc query state= all` (Win) | `systemctl list-units` (Linux)
  │ Log tail: 16KB tail of agent.log + `wevtutil` / `journalctl -n 200`
  ▼
[Go Agent: service/metrics_detail.go - CollectDetailedMetrics()]
  │ Runs collectProcessMetrics() -> sorted by CPU% desc, then RSS bytes desc
  │ Runs collectConnections() -> capped at maxConnectionEntries = 200
  │ Runs collectMemoryDetails() -> virtual memory + swap details
  │ Runs collectServicesSnapshot() -> executes CLI command with 5s timeout
  │ Runs collectLogsSnapshot() -> stateful file seek on agent.log + OS command
  ▼
[Go Agent: service/metrics_detail_ws.go - StartDetailedMetricsLoop()]
  │ payload = collectDetailedMetricsPayload(cfg)
  │ batch = append(batch, payload)
  │ Condition Check: len(batch) >= 1 || time.Since(lastFlush) >= 30s
  ▼
[Go Agent: service/metrics_detail_ws.go - sendDetailedMetricsBatch()]
  │ Serializes JSON body
  │ Constructs HTTP Request:
  │   POST /agent/metrics-detail/batch
  │   Content-Type: application/json
  │   x-agent-token: <cfg.Token>
  │ Executes via httpClient.Do(req) (10s client timeout)
  ▼
[Backend: config/SecurityConfig.java]
  │ Request matches /agent/** -> permitAll() (Bypasses JwtFilter)
  ▼
[Backend: controller/AgentController.java - sendMetricDetailsBatch()]
  │ Extracts header: @RequestHeader("x-agent-token") String agentToken
  │ Delegates to agentService.saveMetricDetailsBatch(requests, agentToken)
  ▼
[Backend: service/AgentService.java - saveMetricDetailsBatch()]
  │ Validates agentToken against companyRepository.findByApiToken(agentToken)
  │ Validates device belongs to company: device.getCompany().getId().equals(company.getId())
  │ Serializes request.getDetails() map to JSON String via objectMapper
  │ Persists: metricDetailRepository.saveAll(details) -> INSERT INTO metric_detail (...)
  │ Constructs MetricDetailResponse (id, detailsJson, createdAt)
  │ Broadcasts via WebSocket:
  │   messagingTemplate.convertAndSend("/topic/device-detail/" + deviceId, response)
  ▼
[Backend: SimpleBroker]
  │ Pushes STOMP frame to "/topic/device-detail/{deviceId}"
  ▼
[Frontend: app/(dashboard)/devices/[deviceId]/page.tsx]
  │ WebSocket subscriber receives MetricDetailResponse
  │ Stores in detailedMetrics state: setDetailedMetrics([parsed, ...prev].slice(0, 20))
  │ If user visits "Detailed" tab:
  │   Parses detailsJson -> ProcessListView, ConnectionsView, ServicesView, LogsView
```

---

## Trace 3: Remote Command Execution & Streaming Output Assembly

This trace tracks a remote shell command dispatched from the UI, routed across two STOMP connections, executed by the Go agent, chunked, transmitted, and reassembled in browser state.

```text
[Operator Browser UI: DeviceDetailPage]
  │ Operator types "systemctl status nginx" and presses Enter / Run
  │ stompRef.current.publish({
  │   destination: "/app/command/" + deviceId,
  │   body: JSON.stringify({
  │     deviceId: "9c3e...",
  │     commandId: "5f8a...",
  │     type: "shell",
  │     payload: "systemctl status nginx"
  │   })
  │ })
  ▼
[Network: WebSocket Connection (Browser -> Backend)]
  │ STOMP SEND frame received by Spring WebSocket inbound channel
  ▼
[Backend: controller/CommandWebSocketController.java - sendCommand()]
  │ @MessageMapping("/command/{deviceId}")
  │ Resolves companyId from Principal
  │ Validates: device.getCompany().getId().equals(companyId)
  │ Generates commandId if empty: UUID.randomUUID().toString()
  │ Relays to destination:
  │   messagingTemplate.convertAndSend("/topic/agent/" + deviceId, request)
  ▼
[Backend: SimpleBroker]
  │ Routes message to subscriber of "/topic/agent/{deviceId}"
  ▼
[Network: WebSocket Connection (Backend -> Agent Command Loop)]
  │ STOMP MESSAGE frame received by Agent gorilla/websocket conn
  ▼
[Go Agent: service/commands.go - runCommandSession()]
  │ Reads frame: Command == "MESSAGE"
  │ Unmarshals JSON to CommandRequest
  │ Validates: request.DeviceID == cfg.DeviceID
  ▼
[Go Agent: service/commands.go - executeCommand()]
  │ Type switch: "shell"
  │ Validates against blockedCommandPattern:
  │   `(?i)\b(rm|remove-item|removeitem|del|erase|rmdir|rd|format)\b`
  │ Prepares command context with commandTimeout = 30 * time.Second
  │ Platform Execution:
  │   Windows: powershell.exe -NoProfile -ExecutionPolicy Bypass -Command ...
  │   Linux/Darwin: sh -c "systemctl status nginx"
  │ Executes: cmd.CombinedOutput()
  │ Sanitizes output: strips null characters ("\x00")
  ▼
[Go Agent: service/commands.go - chunkCommandResult()]
  │ Splits output into chunks of maxCommandChunkBytes = 12 * 1024 (12KB)
  │ If output > 12KB (e.g. 25KB -> 3 chunks):
  │   Chunk 0: status="stream", chunkIndex=0, chunkCount=3, finishedAt=""
  │   Chunk 1: status="stream", chunkIndex=1, chunkCount=3, finishedAt=""
  │   Chunk 2: status="ok",     chunkIndex=2, chunkCount=3, finishedAt="2026-09-27T12:00:00Z"
  ▼
[Go Agent: service/commands.go - sendStompFrame()]
  │ Loops through each chunk:
  │ Sends STOMP SEND frame:
  │   destination: /app/command-result
  │   content-type: application/json
  │   Body: JSON(CommandResult chunk)
  ▼
[Backend: controller/CommandWebSocketController.java - sendCommandResult()]
  │ @MessageMapping("/command-result")
  │ Resolves companyId from Principal
  │ Validates: device.getCompany().getId().equals(companyId)
  │ Relays to:
  │   messagingTemplate.convertAndSend("/topic/command-result/" + result.getDeviceId(), result)
  ▼
[Backend: SimpleBroker]
  │ Forwards message to browser subscription "/topic/command-result/{deviceId}"
  ▼
[Frontend: app/(dashboard)/devices/[deviceId]/page.tsx - upsertCommandResult()]
  │ Incoming chunk arrives:
  │ Stores part in commandChunkBuffers.current[commandId].outputParts[chunkIndex] = chunk.output
  │ Calls assembleChunks(): iterates from 0 to chunkCount-1 and joins strings
  │ If chunk has finishedAt:
  │   Deletes buffer from commandChunkBuffers
  │   Sets status = chunk.status ("ok" or "error")
  │ Updates commandHistory state: unshifts completed result, slices to 20 items
  │ UI updates: terminal window displays full command stdout/stderr with status badge
```

---

## Trace 4: Agent Registration & Retry Under Backend Outage

This trace illustrates what happens when the agent boots before the backend is running, or when the backend temporarily crashes.

```text
[Agent Start: service.StartWorker()]
  │ cfg = config.Load()
  │ Launches Goroutine 1: StartCommandLoop(cfg, stop)
  │ Launches Goroutine 2: StartMetricsWebSocketLoop(cfg, stop)
  │ Launches Goroutine 3: StartDetailedMetricsLoop(cfg, stop, 30s)
  │ Launches Goroutine 4: Background Registration Retry Loop
  │ Launches Goroutine 5: 20s Local Heartbeat Ticker
  ▼
[Agent: service/metrics_ws.go - StartMetricsWebSocketLoop()]
  │ Loop check: cfg.DeviceID == ""
  │ Calls RegisterIfNeeded(cfg):
  │   postJSON("http://localhost:8080/agent/register", payload)
  │   ──> Connection Refused! err = dial tcp 127.0.0.1:8080: connect: connection refused
  │   Logs: "Registration failed: dial tcp ..."
  │   Sleeps 5 seconds
  │ (Goroutine 1 & 4 do the exact same thing concurrently)
  ▼
[5 Seconds Later...]
  │ Loop repeats: cfg.DeviceID == "" -> calls RegisterIfNeeded(cfg) -> Fails again
  ▼
[Backend Comes Online]
  │ Backend Tomcat starts on port 8080
  │ PostgreSQL database is ready
  ▼
[Agent: Next Tick in RegisterIfNeeded()]
  │ postJSON("http://localhost:8080/agent/register", payload) -> Returns 200 OK
  │ Backend AgentController.register() -> AgentService.registerDevice()
  │ Device looked up or created: deviceRepository.save(device)
  │ Returns JSON: {"id": "f47ac10b-58cc-4372-a567-0e02b2c3d479", ...}
  │ Agent sets cfg.DeviceID = result.ID
  │ Agent saves to disk: config.Save(cfg) -> writes to ~/.monitor-agent.json
  ▼
[Agent: Next Iteration in Metrics & Command Loops]
  │ cfg.DeviceID is now non-empty!
  │ Metrics loop calls runMetricsSession():
  │   Dials ws://localhost:8080/ws
  │   Sends STOMP CONNECT frame
  │   Starts streaming batches!
  │ Command loop calls runCommandSession():
  │   Dials ws://localhost:8080/ws
  │   Sends STOMP CONNECT frame
  │   Sends STOMP SUBSCRIBE /topic/agent/{deviceId}
  │   Ready to receive commands!
```

---

## Trace 5: Backend Outage & Network Failure During Metric Batching

This trace proves what happens to metrics when the network disconnects halfway through transmission.

```text
[Go Agent: service/metrics_ws.go - runMetricsSession()]
  │ Goroutine samples metrics every 1s-5s
  │ Appends to in-memory slice: batch = append(batch, payload)
  │ Batch reaches 10 items.
  │ Calls sendMetricsBatch(conn, batch)
  │   sendStompFrame(conn, frame) -> conn.WriteMessage()
  │ ──> NETWORK FAILURE / TCP RST / BROKEN PIPE ──>
  │ conn.WriteMessage() returns error: "write: broken pipe"
  ▼
[Failure Propagation]
  │ sendMetricsBatch() returns err
  │ runMetricsSession() immediately returns err!
  │ defer conn.Close() executes
  │ ⚠️ DATA LOSS VERIFIED:
  │   The `batch` variable is a local function slice in runMetricsSession().
  │   When runMetricsSession() returns, the in-memory batch slice of 10 items
  │   is DEALLOCATED AND GARBAGE COLLECTED. It is NOT saved to disk or queue.
  ▼
[Reconnection Backoff]
  │ StartMetricsWebSocketLoop() catches error:
  │   logrus.Error("Metrics websocket error:", err)
  │ Executes: time.Sleep(3 * time.Second)
  │ Next iteration dials wsURL again:
  │   websocket.DefaultDialer.Dial(wsURL, nil)
  │ If backend is still down, dial fails, sleeps 3s, and repeats.
  │ Once backend recovers: dials successfully, sends CONNECT, and resumes fresh sampling.
```

---

## Trace 6: Device Offline Detection & UI Notification Sweep

This trace demonstrates how the backend detects an unresponsive agent and pushes status changes to the frontend.

```text
[Spring Boot Background Scheduler: service/DeviceStatusScheduler.java]
  │ Runs every 30,000ms (@Scheduled(fixedRate = 30000))
  │ Executes: List<Device> devices = deviceRepository.findAll()
  │ NOTE: Performs full table scan of all devices in the database
  ▼
[Evaluation Loop]
  │ For each device in devices:
  │   if device.getLastSeenAt() == null -> continue
  │   shouldBeOffline = device.getLastSeenAt().isBefore(Instant.now().minusSeconds(30))
  │
  │ Scenario: Agent process was terminated or host crashed 35 seconds ago.
  │   device.getLastSeenAt() is 35s old.
  │   shouldBeOffline == true
  │   device.getStatus() == ONLINE
  ▼
[Status Transition]
  │ device.setStatus(DeviceStatus.OFFLINE)
  │ deviceRepository.save(device) -> UPDATE device SET status = 'OFFLINE' WHERE id = ...
  │ println("Device marked OFFLINE: " + device.getHostname())
  │ Broadcasts status event:
  │   messagingTemplate.convertAndSend(
  │     "/topic/device-status/" + device.getId(),
  │     DeviceStatus.OFFLINE
  │   )
  ▼
[Backend: SimpleBroker]
  │ Dispatches STOMP MESSAGE frame to "/topic/device-status/{deviceId}"
  ▼
[Frontend: app/(dashboard)/devices/[deviceId]/page.tsx]
  │ Active subscription callback triggers:
  │   client.subscribe(`/topic/device-status/${deviceId}`, (message) => {
  │     const nextStatus = message.body?.trim() as DeviceStatus;
  │     setDevice((current) => current ? { ...current, status: nextStatus } : current);
  │   })
  │ Device header badge instantly switches from green "ONLINE" to red "OFFLINE".
```\n