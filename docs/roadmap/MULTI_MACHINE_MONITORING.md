# Multi-Machine Monitoring - Implementation Roadmap

**Status**: Planning Phase
**Priority**: High (Critical Feature Gap)
**Estimated Effort**: 3-4 weeks

---

## MVP Scope (Weeks 1–3)

- **Week 1**: Land database migrations (`machines`, `metrics_history`), storage layer, and machine registration API with hashed API keys + per-user isolation.
- **Week 2**: Build read-only Go agent (HTTPS POST + checksum), ship install script, add revocation path; no mTLS yet.
- **Week 3**: Frontend machine selector + onboarding UI; hide legacy host metrics behind `LOCAL_HOST_METRICS` flag and require machine selection.
- **Deferred**: WebSocket streaming, bulk performance tuning, supply-chain automation, mTLS PKI—log them in “Post-MVP Hardening” once MVP is live.

## Problem Statement

Currently, LunaSentri monitors only the server it's running on. All users see the same metrics (the host server's CPU, memory, disk). This is a critical limitation that prevents the system from being a true multi-user monitoring solution.

### Current Architecture (Single Server)
```
┌─────────────────────────────────────┐
│  LunaSentri Server                  │
│  ┌───────────────────────────────┐  │
│  │ Metrics Collector             │  │
│  │ (monitors host server only)   │  │
│  └───────────────────────────────┘  │
│           ↓                         │
│  ┌───────────────────────────────┐  │
│  │ All Users See Same Metrics    │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

### Target Architecture (Multi-Machine)
```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│ User A's Server  │     │ User B's Server  │     │ User C's Server  │
│ ┌──────────────┐ │     │ ┌──────────────┐ │     │ ┌──────────────┐ │
│ │ Agent        │─┼─────┼─│ Agent        │─┼─────┼─│ Agent        │ │
│ └──────────────┘ │     │ └──────────────┘ │     │ └──────────────┘ │
└──────────────────┘     └──────────────────┘     └──────────────────┘
         │                        │                        │
         └────────────────────────┼────────────────────────┘
                                  ↓
                    ┌─────────────────────────────┐
                    │  LunaSentri API             │
                    │  ┌────────────────────────┐ │
                    │  │ Machine Registry       │ │
                    │  │ Metrics Ingestion      │ │
                    │  │ User Isolation         │ │
                    │  └────────────────────────┘ │
                    └─────────────────────────────┘
                                  ↓
                    ┌─────────────────────────────┐
                    │  User Dashboard             │
                    │  - User A sees only A's     │
                    │  - User B sees only B's     │
                    │  - Admins see all           │
                    └─────────────────────────────┘
```

## Proposed Solution: Lightweight Agent Architecture

### Core Components

#### 1. **LunaSentri Agent** (New Component)
A lightweight binary that runs on user's servers and pushes metrics to the central API.

**Features**:
- Single binary (Go) - easy deployment
- Low resource footprint (<10MB RAM, <0.1% CPU)
- Secure authentication with API keys
- Automatic metric collection every 3-5 seconds
- Reconnection logic for network failures
- Health monitoring and self-healing

**Technology Stack**:
- Go (for cross-platform compilation)
- WebSocket or HTTP/2 for efficient metric streaming
- Same metrics collection code as current backend

#### 2. **Machine Registry** (Backend Enhancement)
Database tables and API endpoints for managing monitored machines.

**Database Schema**:
```sql
CREATE TABLE machines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    hostname TEXT,
    api_key TEXT UNIQUE NOT NULL,
    status TEXT DEFAULT 'offline',
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE metrics_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    machine_id INTEGER NOT NULL,
    cpu_pct REAL,
    mem_used_pct REAL,
    disk_used_pct REAL,
    net_rx_bytes INTEGER,
    net_tx_bytes INTEGER,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (machine_id) REFERENCES machines(id) ON DELETE CASCADE
);

CREATE INDEX idx_metrics_machine_time ON metrics_history(machine_id, timestamp);
```

#### 3. **Metrics Ingestion API** (Backend Enhancement)
New endpoints for agents to push metrics.

**Endpoints**:
- `POST /agent/register` - Register new machine and get API key
- `POST /agent/metrics` - Push metrics data (bulk insert)
- `GET /agent/health` - Health check for agent
- `WS /agent/stream` - WebSocket for real-time metric streaming

#### 4. **User Isolation Layer** (Backend Enhancement)
Ensure users only see their own machines.

**Access Control**:
- Users see only machines they own (`user_id` filtering)
- Admins can see all machines (with user context)
- API keys scoped to specific machine + user
- Alert rules scoped to user's machines only

#### 5. **Frontend Enhancements**
UI for managing multiple machines.

**New Pages/Components**:
- `/machines` - List user's registered machines
- `/machines/add` - Register new machine (get setup script)
- `/machines/:id` - Detailed metrics for specific machine
- Dashboard selector: Choose which machine to monitor
- Multi-machine overview: Grid view of all machines

---

## Implementation Phases

### **Phase 1: Database & Backend Foundation** (Week 1)
**Goal**: Setup data model and API endpoints

**Tasks**:
1. Create `machines` and `metrics_history` tables
2. Add machine CRUD operations in storage layer
3. Implement API key generation and validation
4. Create machine registration endpoint
5. Add user isolation logic
6. Write tests for machine management

**Files to Create/Modify**:
- `apps/api-go/internal/storage/machines.go` (new)
- `apps/api-go/internal/storage/sqlite.go` (add migrations)
- `apps/api-go/internal/machines/service.go` (new)
- `apps/api-go/main.go` (add machine endpoints)

**Deliverables**:
- ✅ Machine registration API
- ✅ API key authentication
- ✅ User isolation working
- ✅ Database migrations

---

### **Phase 2: Agent Development** (Week 2)
**Goal**: Build lightweight agent binary

**Tasks**:
1. Create new Go project: `apps/agent/`
2. Copy metrics collection code from backend
3. Implement API key authentication
4. Add metric push logic (HTTP POST every 5s)
5. Add configuration file support (YAML/JSON)
6. Build cross-platform binaries (Linux, macOS, Windows)
7. Create installation script (bash/PowerShell)

**Project Structure**:
```
apps/agent/
├── main.go                 # Agent entry point
├── config/
│   └── config.go          # Configuration loading
├── metrics/
│   └── collector.go       # Reuse from backend
├── client/
│   └── api_client.go      # HTTP client for API
└── Dockerfile             # Optional containerized agent
```

**Configuration Example** (`lunasentri-agent.yaml`):
```yaml
server:
  url: https://lunasentri-api.example.com
  api_key: your-machine-api-key-here

agent:
  interval: 5s
  machine_name: production-web-server

logging:
  level: info
```

**Deliverables**:
- ✅ Standalone agent binary
- ✅ Linux/macOS/Windows builds
- ✅ Installation script
- ✅ Configuration file support

---

### **Phase 3: Metrics Ingestion** (Week 2-3)
**Goal**: Accept and store metrics from agents

**Tasks**:
1. Create `/agent/metrics` POST endpoint
2. Validate API key and extract machine_id
3. Insert metrics into `metrics_history` table
4. Update machine `last_seen` timestamp
5. Set machine status (online/offline logic)
6. Add bulk insert optimization (batch metrics)
7. Add metrics retention policy (auto-delete old data)

**API Endpoint**:
```go
POST /agent/metrics
Authorization: Bearer <api_key>
Content-Type: application/json

{
  "timestamp": "2025-10-09T12:34:56Z",
  "cpu_pct": 45.2,
  "mem_used_pct": 67.8,
  "disk_used_pct": 23.5,
  "net_rx_bytes": 123456,
  "net_tx_bytes": 654321
}
```

**Deliverables**:
- ✅ Metrics ingestion endpoint
- ✅ Bulk insert support
- ✅ Machine status tracking (online/offline)
- ✅ Metrics retention policy

---

### **Phase 4: Frontend - Machine Management** (Week 3)
**Goal**: UI for adding and managing machines

**Tasks**:
1. Create Machines page (`/machines`)
2. Add "Register Machine" button → shows setup instructions
3. Generate one-liner install script with embedded API key
4. Display list of user's machines (name, status, last seen)
5. Add machine detail view (click machine → see its metrics)
6. Add machine delete functionality
7. Add machine rename/edit

**UI Components**:
```typescript
// apps/web-next/app/machines/page.tsx
- MachinesList component (grid/table)
- AddMachineModal (shows install script)
- MachineCard (status indicator, name, last seen)

// apps/web-next/app/machines/[id]/page.tsx
- MachineDashboard (same as current dashboard, scoped to machine)
```

**Install Script Example**:
```bash
# One-liner for user to copy-paste
curl -sSL https://lunasentri.example.com/install.sh | \
  bash -s -- --api-key=abc123xyz --name="production-server"
```

**Deliverables**:
- ✅ Machines management page
- ✅ Machine registration UI
- ✅ Install script generator
- ✅ Machine status indicators

---

### **Phase 5: Frontend - Multi-Machine Dashboard** (Week 4)
**Goal**: Visualize metrics from multiple machines

**Tasks**:
1. Add machine selector to dashboard (dropdown)
2. Fetch metrics for selected machine
3. Update WebSocket to support machine_id parameter
4. Create multi-machine overview (grid of mini-dashboards)
5. Update alert rules to be machine-specific
6. Add machine filter to alerts page

**Dashboard Enhancements**:
```typescript
// Machine selector in header
<MachineSelector
  machines={userMachines}
  selected={currentMachine}
  onChange={setCurrentMachine}
/>

// Multi-machine grid view
<MachineGrid>
  {machines.map(m => (
    <MiniDashboard machine={m} />
  ))}
</MachineGrid>
```

**Deliverables**:
- ✅ Machine-specific dashboard
- ✅ Multi-machine overview
- ✅ Machine selector UI
- ✅ Machine-scoped alerts

---

## Security Considerations

### 1. **API Key Security**
- Generate cryptographically secure API keys (256-bit)
- Store hashed in database (bcrypt or SHA-256)
- Include machine_id + user_id in key claims
- Implement key rotation mechanism
- Support key revocation

### 2. **Agent Authentication**
- Require API key on every request
- Validate key + extract machine_id
- Rate limit metric submissions (prevent abuse)
- Log all agent activity

### 3. **User Isolation**
- Strict user_id filtering on all queries
- Prevent users from seeing other users' machines
- Admins can view all with audit logging

### 4. **Data Protection**
- Encrypt API keys in transit (HTTPS only)
- Store metrics with user_id association
- Implement data retention policies
- GDPR compliance: allow data export/deletion

---

## Deployment Strategy

### 1. **Agent Distribution**
- Host binaries on GitHub Releases
- Provide install scripts for major platforms
- Create Docker image for containerized deployments
- Package managers (future): apt, yum, brew

### 2. **Database Migration**
- Create migration for new tables
- Backward compatible (don't break existing system)
- Add default machine for existing users (localhost)

### 3. **Rollout Plan**
1. Deploy backend with new endpoints (backward compatible)
2. Deploy frontend with machines page (optional feature)
3. Release agent binaries
4. Document setup process
5. Migrate existing single-machine setup to multi-machine

---

## Testing Strategy

### 1. **Agent Testing**
- Unit tests for metrics collection
- Integration tests for API communication
- Load testing (100+ machines reporting simultaneously)
- Network failure scenarios (reconnection logic)

### 2. **Backend Testing**
- Unit tests for machine CRUD
- Integration tests for metrics ingestion
- User isolation tests (security critical)
- Performance tests (bulk metric inserts)

### 3. **Frontend Testing**
- Component tests for machine UI
- E2E tests for registration flow
- Multi-machine dashboard tests

---

## Success Metrics

- ✅ Agent binary size < 15MB
- ✅ Agent memory usage < 10MB
- ✅ Agent CPU usage < 0.5%
- ✅ Metric latency < 1 second (agent → API → dashboard)
- ✅ Support 100+ machines per user
- ✅ System handles 1000+ machines globally
- ✅ Zero data leakage between users

---

## Alternative: Quick Win (Minimal Viable Solution)

If full agent architecture is too complex initially, consider this **Phase 0**:

### **SSH-Based Remote Monitoring** (1 week)
Store SSH credentials and run metrics collection remotely.

**Pros**:
- No agent installation required
- Reuse existing metrics code
- Faster to implement

**Cons**:
- Less secure (store SSH keys)
- Requires SSH access (firewall issues)
- Higher latency
- Not real-time

**Implementation**:
1. Add SSH credential storage (encrypted)
2. Add remote SSH executor
3. Run existing metrics collector via SSH
4. Store results per machine

**Not recommended long-term**, but could validate concept quickly.

---

## Recommendation

**Proceed with full agent-based architecture** following the 4-week phased approach. This provides:
- ✅ Scalable foundation
- ✅ Real-time monitoring
- ✅ Better security
- ✅ Professional product quality
- ✅ Market-ready solution

Start with **Phase 1** (backend foundation) this week.

---

## Next Actions

1. **Approve this roadmap** or request modifications
2. **Prioritize phases** (can we compress timeline?)
3. **Assign resources** (solo dev or team?)
4. **Create GitHub issues** for tracking
5. **Begin Phase 1 implementation**

---

**Document Status**: Draft for Review
**Last Updated**: 2025-10-09
**Author**: CTO / Technical Lead
