# End-to-End Data Flow & Metric Lineage

## 1. Complete Metric Lineage: Kernel to Visual Pixel

The diagram below traces the precise transformations of a telemetry metric across every layer of the architecture:

```mermaid
flowchart TD
    subgraph Host["Target Machine Kernel"]
        RawCPU["/proc/stat or GetSystemTimes\n(Raw ticks / time elapsed)"]
        RawMem["/proc/meminfo or GlobalMemoryStatusEx\n(Bytes total, free, available)"]
        RawDisk["statvfs or GetDiskFreeSpaceEx\n(Blocks total, free)"]
        RawNet["/proc/net/dev or GetIfTable\n(Cumulative byte counters)"]
    end

    subgraph GoAgent["Go Agent Layer (monitor-agent)"]
        gopsutil["gopsutil v3 (cpu, mem, disk, net)"]
        RawCPU & RawMem & RawDisk & RawNet --> gopsutil
        
        CollectMetrics["CollectMetrics() (service/metrics.go)\n- cpuUsage: float64 (%)\n- memoryUsage: float64 (%)\n- diskUsage: float64 (%)\n- networkIn: uint64 (cumulative bytes)\n- networkOut: uint64 (cumulative bytes)"]
        gopsutil --> CollectMetrics
        
        BatchAppend["runMetricsSession() (service/metrics_ws.go)\n- Injects deviceId: cfg.DeviceID\n- Appends to batch: []map[string]interface{}\n- Checks: len(batch) >= 10 || elapsed >= 5s"]
        CollectMetrics --> BatchAppend
        
        StompSerialize["sendMetricsBatch() (service/metrics_ws.go)\n- json.Marshal(batch)\n- Wraps into STOMP SEND frame\n- Destination: /app/agent/metrics-batch"]
        BatchAppend --> StompSerialize
    end

    subgraph Transport1["Transport Layer (WebSocket)"]
        WSTransport["gorilla/websocket TCP Socket\n(ws:// or wss:// /ws)"]
        StompSerialize --> WSTransport
    end

    subgraph Backend["Spring Boot Backend"]
        InboundWS["WebSocket Inbound Channel & Interceptor"]
        WSTransport --> InboundWS
        
        WSController["AgentWebSocketController.receiveMetricsBatch()\n- Destination: /app/agent/metrics-batch\n- Maps to List<MetricRequest> DTOs\n- Verifies device ownership (tenant isolation)"]
        InboundWS --> WSController
        
        AgentService["AgentService.saveMetricsBatch()\n- @Transactional\n- Resolves Device entity\n- Generates Instant.now()\n- Builds List<Metric> entities\n- Sets device.lastSeenAt = now\n- Sets device.status = ONLINE"]
        WSController --> AgentService
        
        DBInsert["metricRepository.saveAll(metrics)\n- Batch INSERT into PostgreSQL / TimescaleDB hypertable\ndeviceRepository.save(device)\n- UPDATE device SET last_seen_at = ..., status = ..."]
        AgentService --> DBInsert
        
        WSBroadcast["messagingTemplate.convertAndSend()\n- Destination: /topic/device/{deviceId}\n- Payload: latestMetric (last item in batch)"]
        AgentService --> WSBroadcast
    end

    subgraph SimpleBrokerSub["Spring In-Memory Broker"]
        SimpleBroker["SimpleBroker (/topic)\n- Matches active subscriptions for /topic/device/{deviceId}"]
        WSBroadcast --> SimpleBroker
    end

    subgraph Transport2["Transport Layer (SockJS / WebSocket)"]
        ClientWSTransport["SockJS WebSocket Connection (Client Session)"]
        SimpleBroker --> ClientWSTransport
    end

    subgraph Frontend["Next.js Browser Client"]
        StompClient["@stomp/stompjs Client Subscription\n- Topic: /topic/device/{deviceId}"]
        ClientWSTransport --> StompClient
        
        StateUpdate["setMetrics((prev) => [parsed, ...prev].slice(0, 60))\nsetDevice((curr) => curr ? { ...curr, lastSeenAt } : curr)"]
        StompClient --> StateUpdate
        
        SVGChart["MetricChart.tsx Component\n- Pure SVG coordinate mapping\n- Real-time Polyline rendering\n- StatTile number updates"]
        StateUpdate --> SVGChart
    end
```

---

## 2. Deep Diagnostic Snapshot Data Lineage

The data flow for heavyweight system snapshots operates across REST and WebSocket:

```mermaid
sequenceDiagram
    autonumber
    participant Host as OS Subsystems
    participant Agent as Go Agent (Detail Loop)
    participant REST as Spring REST Ingest (/agent)
    participant DB as TimescaleDB (metric_detail)
    participant Broker as Spring SimpleBroker
    participant UI as Next.js Dashboard

    Note over Host,Agent: Every 30 seconds
    Agent->>Host: Processes: PID, name, cmdline, CPU%, RSS, threads
    Agent->>Host: Connections: Sockets (TCP/UDP, status, local/remote IP:port)
    Agent->>Host: Memory: Virtual memory, buffers, cached, slab, swap
    Agent->>Host: Services: sc.exe / systemctl list-units
    Agent->>Host: Logs: agent.log tail + wevtutil / journalctl
    Host-->>Agent: Raw OS data structures
    Agent->>Agent: Sort processes by CPU%/RSS, cap sockets to 200, cap logs to 16KB
    Agent->>REST: POST /agent/metrics-detail/batch (Header: x-agent-token)
    REST->>REST: Verify token & company ownership
    REST->>DB: INSERT INTO metric_detail (device_id, details_json, created_at)
    REST->>Broker: convertAndSend("/topic/device-detail/{deviceId}", response)
    Broker-->>UI: STOMP MESSAGE /topic/device-detail/{deviceId}
    UI->>UI: Parse detailsJson -> Render Process Table, Sockets, Services, Logs
```

---

## 3. Remote Command Execution Lineage

```mermaid
sequenceDiagram
    autonumber
    participant Operator as Browser UI (Operator)
    participant Broker as Spring SimpleBroker
    participant Agent as Go Agent (Command Loop)
    participant OS as Local Host Shell (sh/PowerShell)

    Operator->>Broker: SEND /app/command/{deviceId} {type: "shell", payload: "df -h"}
    Broker->>Broker: Verify companyId matches device owner
    Broker-->>Agent: MESSAGE /topic/agent/{deviceId}
    Agent->>Agent: Check blockedCommandPattern regex
    Agent->>OS: Execute with 30s timeout (PowerShell / sh)
    OS-->>Agent: Raw stdout & stderr bytes
    Agent->>Agent: Strip null bytes, split into <=12KB chunks
    loop For each chunk (0..N-1)
        Agent->>Broker: SEND /app/command-result {chunkIndex: i, chunkCount: N, status: "stream"}
        Broker-->>Operator: MESSAGE /topic/command-result/{deviceId}
        Operator->>Operator: Buffer chunk in memory & reassemble
    end
    Agent->>Broker: SEND /app/command-result {chunkIndex: N-1, status: "ok", finishedAt: timestamp}
    Broker-->>Operator: MESSAGE /topic/command-result/{deviceId}
    Operator->>Operator: Mark command finished, clear buffer, display complete output
```\n