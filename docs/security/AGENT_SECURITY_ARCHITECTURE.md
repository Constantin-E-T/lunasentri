# LunaSentri Agent - Security Architecture & Threat Model

**Status**: Phase 1 Implementation (API Key Auth) - mTLS Planned
**Priority**: CRITICAL - Security First Approach
**Last Updated**: 2025-10-09

---

## Current Implementation Status

### Phase 1: API Key Authentication (CURRENT)

The current implementation uses **scoped API keys with SHA-256 hashing** for machine authentication:

- **Machine Registration**: When a machine is registered, it receives a unique API key
- **Key Storage**: API keys are hashed (SHA-256) before storage in the database
- **Authentication**: Agents send the API key in request headers; server validates against stored hash
- **Revocation**: Keys can be revoked by deleting the machine record (immediate effect)
- **Transport Security**: All communication uses HTTPS (TLS 1.2+)

**Implementation Details**:

```go
// Machine service generates and hashes API keys
apiKey, _ := machines.GenerateAPIKey()          // 32-byte random key
hash := machines.HashAPIKey(apiKey)             // SHA-256 hash
machine, _ := service.RegisterMachine(ctx, userID, name, hostname, hash)

// Agent includes key in headers
req.Header.Set("X-API-Key", apiKey)

// Server validates by comparing hash
storedMachine, _ := store.GetMachineByAPIKey(ctx, hashFromRequest)
```

**Security Trade-offs**:

- ✅ **Simple**: Easy to implement and debug
- ✅ **Adequate for MVP**: Provides basic authentication and authorization
- ✅ **Revocable**: Can be invalidated immediately
- ⚠️ **Key Theft Risk**: If API server is compromised, attacker could steal hashed keys
- ⚠️ **No Automatic Rotation**: Keys must be manually rotated
- ⚠️ **Replay Risk**: Without additional measures, stolen keys can be replayed

### Phase 2: Mutual TLS (ROADMAP)

**Target Architecture**: Upgrade to client certificate-based mTLS authentication

See `docs/roadmap/MULTI_MACHINE_MONITORING.md` for the complete Phase 2 specification, which includes:

- Client certificate generation during machine registration
- mTLS handshake for cryptographic authentication
- Automatic certificate rotation (1-year validity, 30-day renewal window)
- CRL/OCSP-based revocation
- Elimination of API key theft vectors

**Migration Path**: The current API key implementation is designed to be replaced without breaking changes to the storage layer or machine registration flow.

---

## MVP Scope (Phase 1 Complete → Phase 2 Active)

- **Objective**: Ship a read-only agent that authenticates with scoped API keys, runs as a non-root user, and sends outbound HTTPS metrics only.
- **Completed**:
  - Machine registry with per-user isolation ✅
  - API key generation and hashing ✅
  - Database schema for machines and metrics_history ✅
  - Storage layer with full test coverage ✅
- **Next Phase (Phase 2)**:
  - Agent binary implementation (Go client)
  - Agent ingestion endpoints in API server
  - mTLS certificate infrastructure
  - Binary checksum verification
  - Onboarding documentation
- **Out of Scope (Post-Launch Hardening)**: External security audit, bug bounty program, full supply-chain attestation.

---

## Executive Summary

LunaSentri agents will run on user production servers with elevated privileges (to read system metrics). This creates a **critical trust boundary** where security is paramount. A compromised LunaSentri platform must NOT compromise user servers.

### Security Principles

1. **Zero Trust Architecture** - Assume LunaSentri API can be compromised
2. **Principle of Least Privilege** - Agent has minimal permissions
3. **Defense in Depth** - Multiple security layers
4. **Fail Secure** - Agent fails closed, not open
5. **Cryptographic Identity** - Mutual TLS authentication
6. **Immutable Audit Trail** - All actions logged

---

## Threat Model

### Attack Vectors

| Threat | Impact | Likelihood | Mitigation Priority |
|--------|--------|------------|-------------------|
| **API Server Compromise** | Attacker sends malicious commands to all agents | High | 🔴 CRITICAL |
| **Man-in-the-Middle** | Intercept/modify metrics in transit | Medium | 🟠 HIGH |
| **Stolen API Keys** | Impersonate legitimate agent | Medium | 🟠 HIGH |
| **Agent Binary Tampering** | Modified agent with backdoor | Medium | 🟠 HIGH |
| **Privilege Escalation** | Agent exploited to gain root access | High | 🔴 CRITICAL |
| **Injection Attacks** | SQL/Command injection via metrics | Low | 🟡 MEDIUM |
| **DDoS via Agents** | Use agents to attack third parties | Medium | 🟠 HIGH |
| **Data Exfiltration** | Steal sensitive data from user servers | High | 🔴 CRITICAL |

---

## Security Architecture Design

### 1. **Read-Only Agent (CRITICAL)**

**Decision**: Agent is **strictly read-only** with NO remote command execution.

```
┌─────────────────────────────────────────────────────┐
│  User's Production Server                           │
│  ┌───────────────────────────────────────────────┐  │
│  │  LunaSentri Agent (Read-Only Mode)            │  │
│  │                                                │  │
│  │  ✅ CAN:                                       │  │
│  │    - Read /proc/* for metrics                 │  │
│  │    - Read /sys/* for system info              │  │
│  │    - Read disk usage (df)                     │  │
│  │    - Send metrics to API (outbound only)      │  │
│  │                                                │  │
│  │  ❌ CANNOT:                                    │  │
│  │    - Execute remote commands                  │  │
│  │    - Write files                              │  │
│  │    - Modify system configuration              │  │
│  │    - Open listening sockets                   │  │
│  │    - Accept incoming connections              │  │
│  │    - Read sensitive files (/etc/shadow, SSH)  │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

**Implementation**:

- Agent has **no remote code execution** capability
- No websocket server (only client)
- No command interpreter
- Metrics collection uses safe system calls only
- Hardcoded metric collection logic (not configurable remotely)

**Why This Matters**:
🛡️ **Even if LunaSentri API is fully compromised**, attacker cannot:

- Execute commands on user servers
- Read sensitive files
- Install malware
- Pivot to other systems

---

### 2. **Mutual TLS (mTLS) Authentication**

**Decision**: Use client certificates for cryptographic authentication, not API keys.

> **MVP Note**: The initial launch uses scoped API keys with HTTPS + revocation. mTLS remains the target architecture and moves into active work immediately after the MVP agent ships.

```
┌──────────────────┐                    ┌──────────────────┐
│  Agent           │                    │  LunaSentri API  │
│  ┌────────────┐  │                    │  ┌────────────┐  │
│  │ Client     │  │  mTLS Handshake    │  │ Server     │  │
│  │ Cert       │──┼───────────────────>│  │ Cert       │  │
│  │            │<─┼────────────────────┤  │            │  │
│  └────────────┘  │  Verify Certs      │  └────────────┘  │
│                  │                    │                  │
│  Certificate:    │                    │  Verifies:       │
│  - Machine ID    │                    │  - Valid cert    │
│  - User ID       │                    │  - Not revoked   │
│  - Expiry: 1yr   │                    │  - Matches DB    │
│  - Signed by CA  │                    │  - Rate limits   │
└──────────────────┘                    └──────────────────┘
```

**Why mTLS > API Keys**:

- ✅ **Cannot be stolen** from server compromise (private key stays on agent)
- ✅ **Cryptographically verified** (not just a password)
- ✅ **Automatic rotation** (short-lived certs)
- ✅ **Revocation support** (CRL/OCSP)
- ✅ **Prevents replay attacks** (TLS nonce)

**Implementation**:

```go
// Agent side
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{clientCert},
    RootCAs:      caCertPool,
    MinVersion:   tls.VersionTLS13, // Force TLS 1.3
}

// API side
tlsConfig := &tls.Config{
    ClientAuth: tls.RequireAndVerifyClientCert,
    ClientCAs:  caCertPool,
    MinVersion: tls.VersionTLS13,
}
```

---

### 3. **Certificate Lifecycle Management**

**Certificate Generation** (on machine registration):

```bash
# User runs on their server
curl -sSL https://lunasentri.com/install.sh | bash

# Install script:
1. Generates CSR (Certificate Signing Request) locally
2. Sends CSR to API with user auth token
3. API signs CSR with internal CA
4. Returns signed certificate (valid 1 year)
5. Agent stores cert + private key (never leaves server)
```

**Certificate Storage**:

```
/etc/lunasentri/
├── agent.crt          # Public certificate (can be read)
├── agent.key          # Private key (600 permissions, encrypted)
├── ca.crt             # LunaSentri CA cert
└── config.yaml        # Agent configuration
```

**Certificate Revocation**:

- API maintains Certificate Revocation List (CRL)
- Agent checks CRL on startup
- Revoked certs rejected at TLS handshake
- User can revoke cert from dashboard (immediate effect)

**Auto-Rotation**:

- Certs expire after 1 year
- Agent auto-renews 30 days before expiry
- Zero-downtime rotation

---

### 4. **No Remote Code Execution (RCE) Protection**

**Architecture Decision**: Agent has **zero configurability** from API.

```go
// ❌ DANGEROUS - DO NOT IMPLEMENT
type MetricConfig struct {
    Command string `json:"command"` // Never allow this!
}

// ✅ SAFE - Hardcoded metric collection
func (a *Agent) CollectMetrics() Metrics {
    // Hardcoded, safe system calls only
    return Metrics{
        CPU:  getCPUUsage(),      // Safe: reads /proc/stat
        Mem:  getMemoryUsage(),   // Safe: reads /proc/meminfo
        Disk: getDiskUsage(),     // Safe: syscall.Statfs
        Net:  getNetworkStats(),  // Safe: reads /proc/net/dev
    }
}
```

**What Agent NEVER Does**:

- ❌ Execute shell commands from API
- ❌ Download and run scripts
- ❌ Modify its own code
- ❌ Accept configuration changes remotely
- ❌ Open reverse shells
- ❌ Run as root (drops privileges)

**Metrics Collection Whitelist**:

```go
// Only these files can be read
var allowedPaths = []string{
    "/proc/stat",
    "/proc/meminfo",
    "/proc/net/dev",
    "/proc/loadavg",
    "/sys/class/thermal/",
}

func readMetricFile(path string) ([]byte, error) {
    // Verify path is in whitelist
    if !isAllowedPath(path) {
        return nil, errors.New("forbidden path")
    }

    // Read with timeout
    return os.ReadFile(path)
}
```

---

### 5. **Privilege Isolation**

**Run as Non-Root User**:

```bash
# Agent runs as dedicated user with minimal permissions
useradd -r -s /bin/false lunasentri
chown -R lunasentri:lunasentri /opt/lunasentri

# Systemd service runs as lunasentri user
[Service]
User=lunasentri
Group=lunasentri
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=/
ReadWritePaths=/var/lib/lunasentri
```

**Linux Capabilities** (not full root):

```bash
# Only grant specific capabilities needed for metrics
setcap cap_sys_ptrace,cap_net_admin=eip /opt/lunasentri/agent

# Drop all other capabilities
```

**Sandboxing** (future enhancement):

- Run in container (Docker/systemd-nspawn)
- SELinux/AppArmor profiles
- Seccomp filters (restrict syscalls)

---

### 6. **Outbound-Only Communication**

**Network Architecture**:

```
┌──────────────────────────────────────┐
│  User's Server (Firewall)            │
│  ┌────────────────────────────────┐  │
│  │  Agent                         │  │
│  │  - No listening ports          │  │
│  │  - Outbound HTTPS only         │  │
│  │  - No incoming connections     │  │
│  └────────────────────────────────┘  │
│              ↓ (Outbound only)       │
│         ┌──────────┐                 │
│         │ Firewall │                 │
│         └──────────┘                 │
│              ↓                       │
└──────────────┼───────────────────────┘
               ↓ HTTPS (443)
    ┌──────────────────────┐
    │  LunaSentri API      │
    │  (Internet)          │
    └──────────────────────┘
```

**Implementation**:

- Agent initiates all connections (acts as HTTPS client)
- No listening sockets (not a server)
- Firewall-friendly (only needs outbound HTTPS)
- Works behind NAT/corporate firewalls

**Connection Pattern**:

```go
// Agent establishes long-lived connection
func (a *Agent) Connect() {
    for {
        // Open HTTPS connection
        conn, err := tls.Dial("tcp", "api.lunasentri.com:443", tlsConfig)
        if err != nil {
            time.Sleep(backoff)
            continue
        }

        // Send metrics periodically
        ticker := time.NewTicker(5 * time.Second)
        for range ticker.C {
            metrics := a.CollectMetrics()
            conn.Write(serializeMetrics(metrics))
        }
    }
}
```

---

### 7. **Code Signing & Binary Verification**

**Problem**: How do users trust the agent binary isn't malicious?

**Solution**: Cryptographically sign all agent binaries.

**Implementation**:

```bash
# 1. Build agent
GOOS=linux GOARCH=amd64 go build -o agent-linux-amd64

# 2. Sign binary with GPG key
gpg --detach-sign --armor agent-linux-amd64

# 3. Publish signature alongside binary
# GitHub Release:
#   - agent-linux-amd64
#   - agent-linux-amd64.asc (signature)
#   - checksums.txt (SHA256 hashes)
#   - checksums.txt.asc (signature)

# 4. User verifies before running
gpg --verify agent-linux-amd64.asc agent-linux-amd64
sha256sum -c checksums.txt
```

**Install Script Verification**:

```bash
#!/bin/bash
# install.sh - Verifies binary before running

# Download LunaSentri public key
curl -sSL https://lunasentri.com/gpg.pub | gpg --import

# Download agent binary
curl -sSL https://github.com/lunasentri/agent/releases/download/v1.0.0/agent-linux-amd64 -o /tmp/agent

# Download signature
curl -sSL https://github.com/lunasentri/agent/releases/download/v1.0.0/agent-linux-amd64.asc -o /tmp/agent.asc

# Verify signature
if ! gpg --verify /tmp/agent.asc /tmp/agent; then
    echo "ERROR: Binary signature verification failed!"
    exit 1
fi

# Install agent
sudo mv /tmp/agent /opt/lunasentri/agent
sudo chmod +x /opt/lunasentri/agent
```

---

### 8. **Data Minimization & Privacy**

**What Agent Sends**:

```json
{
  "timestamp": "2025-10-09T12:34:56Z",
  "cpu_pct": 45.2,
  "mem_used_pct": 67.8,
  "disk_used_pct": 23.5,
  "net_rx_bytes": 123456,
  "net_tx_bytes": 654321
}
```

**What Agent NEVER Sends**:

- ❌ Environment variables (may contain secrets)
- ❌ Running processes (may reveal business logic)
- ❌ File contents
- ❌ Network traffic content (only byte counts)
- ❌ User data
- ❌ Log files
- ❌ Configuration files

**Anonymization**:

- No personally identifiable information (PII)
- No business-sensitive data
- Only aggregated system metrics

---

### 9. **Incident Response & Kill Switch**

**Scenario**: LunaSentri API is compromised. What happens?

**Response Plan**:

1. **Immediate Kill Switch**:

```bash
# API operator revokes all certificates
POST /admin/emergency/revoke-all-certs

# All agents verify CRL on next heartbeat (30s)
# All agents disconnect and stop sending data
```

2. **Agent Auto-Shutdown**:

```go
// Agent checks certificate status every 30s
func (a *Agent) healthCheck() {
    cert := a.loadCertificate()
    if cert.IsRevoked() || cert.IsExpired() {
        log.Fatal("Certificate revoked or expired. Shutting down.")
        os.Exit(0) // Clean shutdown
    }
}
```

3. **User Notification**:

- Email all users about security incident
- Instructions to manually stop agents
- Timeline for resolution

4. **Post-Incident**:

- Issue new CA certificate
- Users re-register agents with new certs
- Audit logs reviewed
- Security report published

---

### 10. **Audit Logging**

**Agent-Side Logging**:

```
2025-10-09 12:34:56 INFO  Agent started (version 1.0.0)
2025-10-09 12:34:57 INFO  TLS connection established (api.lunasentri.com)
2025-10-09 12:34:57 INFO  Certificate verified (expires: 2026-10-09)
2025-10-09 12:35:02 INFO  Metrics sent (cpu: 45.2%, mem: 67.8%)
2025-10-09 12:35:07 INFO  Metrics sent (cpu: 46.1%, mem: 68.2%)
```

**API-Side Logging**:

```
2025-10-09 12:34:57 INFO  Agent connected (machine_id: 123, user_id: 45)
2025-10-09 12:34:57 INFO  Certificate validated (CN: machine-123, expires: 2026-10-09)
2025-10-09 12:35:02 INFO  Metrics received (machine_id: 123, valid: true)
```

**Immutable Audit Trail**:

- All agent connections logged
- Certificate issuance/revocation logged
- Metrics ingestion logged
- Logs stored in append-only storage
- Logs cannot be deleted (tamper-proof)

---

## Security Best Practices

### Development Phase

1. **Security Code Review**:
   - All agent code reviewed by security expert
   - Focus on RCE vulnerabilities
   - Static analysis (gosec, semgrep)
   - Dependency scanning (trivy, snyk)

2. **Penetration Testing**:
   - Hire external security firm
   - Test agent isolation
   - Attempt privilege escalation
   - Test TLS implementation

3. **Open Source**:
   - **Make agent code open source** (transparency builds trust)
   - Community can audit code
   - Bug bounty program

### Deployment Phase

1. **Gradual Rollout**:
   - Beta test with internal servers first
   - Limited rollout to trusted users
   - Monitor for anomalies
   - Full release after 30 days

2. **Security Monitoring**:
   - Anomaly detection (unusual metric patterns)
   - Failed TLS handshake alerts
   - Certificate revocation monitoring

3. **Incident Response Plan**:
   - 24/7 security contact
   - Emergency kill switch procedure
   - User notification templates
   - Post-mortem process

---

## Alternative Architecture: Agentless Monitoring

**If agent security is too risky**, consider:

### **Option: SSH Bastion Monitoring**

```
┌──────────────────┐
│  User's Server   │
│  ┌────────────┐  │
│  │ SSH Daemon │  │ (standard SSH, no custom agent)
│  └────────────┘  │
└────────┬─────────┘
         │ SSH
         ↓
┌────────────────────────────┐
│  LunaSentri Bastion        │
│  ┌──────────────────────┐  │
│  │ Isolated Collector   │  │
│  │ (runs in container)  │  │
│  └──────────────────────┘  │
│         ↓                  │
│  ┌──────────────────────┐  │
│  │ Metrics API          │  │
│  └──────────────────────┘  │
└────────────────────────────┘
```

**Pros**:

- ✅ No custom agent needed
- ✅ Uses standard SSH (well-tested)
- ✅ User controls SSH access (can revoke anytime)

**Cons**:

- ❌ Requires SSH access (firewall issues)
- ❌ SSH credentials stored in LunaSentri
- ❌ Higher latency
- ❌ If LunaSentri compromised, attacker has SSH access

**Not recommended** - SSH credentials are more dangerous than read-only agent.

---

## Recommended Architecture: mTLS + Read-Only Agent

**Final Recommendation**:

| Component | Security Measure |
|-----------|-----------------|
| **Authentication** | Mutual TLS with client certificates |
| **Authorization** | Certificate-based machine identity |
| **Communication** | TLS 1.3, outbound-only, no listening ports |
| **Privileges** | Non-root user, Linux capabilities only |
| **Code Execution** | Zero remote execution, hardcoded metrics only |
| **Binary Integrity** | GPG-signed binaries, verification on install |
| **Data Privacy** | Minimal metrics, no sensitive data |
| **Incident Response** | Certificate revocation, kill switch, audit logs |
| **Trust** | Open source agent code, penetration tested |

---

## Implementation Checklist

### Phase 1: Security Foundation (Week 1-2)

- [ ] Design mTLS certificate architecture
- [ ] Implement Certificate Authority (CA) service
- [ ] Build certificate issuance API
- [ ] Create certificate revocation system (CRL)
- [ ] Implement agent binary signing (GPG)
- [ ] Write security code review checklist

### Phase 2: Agent Development (Week 2-3)

- [ ] Build read-only metrics collector
- [ ] Implement mTLS client
- [ ] Add certificate verification
- [ ] Create non-root user installer
- [ ] Add privilege dropping logic
- [ ] Implement audit logging

### Phase 3: Testing & Hardening (Week 3-4)

- [ ] Static analysis (gosec, semgrep)
- [ ] Dependency scanning (trivy)
- [ ] Penetration testing
- [ ] Load testing (1000+ agents)
- [ ] Incident response drill
- [ ] Security documentation

### Phase 4: Open Source & Transparency (Week 4)

- [ ] Open source agent repository
- [ ] Security audit published
- [ ] Bug bounty program launched
- [ ] Community code review
- [ ] Security policy (SECURITY.md)

---

## Trust & Transparency

### User Trust Factors

1. **Open Source Agent**:
   - Full source code available on GitHub
   - Community can audit every line
   - No hidden backdoors possible

2. **Security Audit**:
   - Third-party security firm audit
   - Publish results publicly
   - Fix all findings before launch

3. **Clear Communication**:
   - Security policy published
   - Incident response plan public
   - Regular security updates

4. **User Control**:
   - Users can revoke agent access anytime
   - Certificate-based (user controls private key)
   - Agent can be uninstalled without API access

---

## Conclusion

**Security is the #1 priority** for LunaSentri agent architecture. By implementing:

- ✅ Read-only agent design
- ✅ Mutual TLS authentication
- ✅ No remote code execution
- ✅ Open source transparency
- ✅ Comprehensive incident response

We create a **trustworthy monitoring solution** that users can confidently deploy on production servers.

**Next Steps**:

1. Review this security architecture
2. Approve or request modifications
3. Begin Phase 1 security implementation
4. Consider external security consultation

---

**Document Status**: Security Architecture Proposal
**Requires**: CTO Approval, Security Expert Review
**Risk Level**: CRITICAL - User Server Security
