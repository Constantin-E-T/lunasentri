# LunaSentri Roadmap & Security Checklist

> Last updated: 2025-10-07
>
> Guiding principles: multi-tenant by default, outbound-only agents, least privilege, user-owned data.

## ✅ Completed (Current Sprint)
- [x] Authenticated dashboard with live metrics
- [x] Alert rules & events with UI management
- [x] Real-time alert badge/toasts
- [x] System info card & polished dashboard layout
- [x] dev-reset script for clean local start

## 🚧 In Progress / Next Up

### 1. Notification Channels
- [ ] Design alert notifier interface (fan-out pattern) *(backend)*
- [ ] Add webhook delivery (per-user config) *(backend)*
- [ ] Add webhook settings UI *(frontend)*
- [ ] Send test payload feature to validate webhook *(frontend)*

### 2. Multi-Server Support
- [x] Servers table with owner_user_id & API key *(backend)*
- [x] Metrics ingestion endpoint with agent token auth *(backend)*
- [x] UI: "Add Server" flow + API key management *(frontend)*
- [x] Dashboard toggle to select server scope *(frontend)*

### 3. Agent Packaging & Security
- [ ] Document threat model & security requirements *(docs)*
- [ ] Build minimal agent binary (outbound HTTPS only) *(agent repo)*
- [ ] Provide install scripts (bash, systemd, docker) with checksum verification *(agent repo)*
- [ ] Create agent token revocation flow *(backend/UI)*
- [ ] Add monitoring for agent heartbeat + alert when silent *(backend/UI)*

### 4. Testing & CI Guardrails
- [ ] Rebuild main package integration test (replace disabled file)
- [ ] Add smoke test hitting /metrics, /alerts, /system/info *(backend)*
- [ ] Add frontend e2e happy-path (login → view metrics → ack alert)* (optional after deploy)*

### 5. Security Hardening
- [ ] Store hashed webhook secrets & sign outgoing payloads *(backend)*
- [ ] Enforce rate limiting on alert evaluation & webhook delivery *(backend)*
- [ ] Provide customer guidance for running agent as non-root *(docs)*
- [ ] Publish agent source + build instructions *(agent repo)*
- [ ] Prepare response plan for token/binary compromise *(docs)*

### 6. Supply Chain Integrity
- [ ] Introduce reproducible builds with signed provenance (SLSA level 2 target) *(platform)*
- [ ] Add automated checksum + signature verification in release pipeline *(platform)*
- [ ] Gate installer publication on passing malware scan + dependency audit *(platform)*
- [ ] Stand up tamper-evident release ledger with human approver sign-off *(leadership)*

### 7. Guided Support Assistant (AI Chatbot)
- [ ] Stand up retrieval service with embedded docs/FAQs (RAG) *(platform)*
- [ ] Build in-app chat UI with guardrails + human handoff *(frontend/backend)*
- [ ] Instrument chat logging + feedback loop, curate training corpus *(product/docs)*
- [ ] Stage fine-tuning or prompt updates via review pipeline *(platform/leadership)*

## 📌 Future Considerations
- Telegram/email/SMS notification fan-out
- Web push (browser/mobile) after core channels work
- Historical metrics charts & retention policies
- Multi-user team sharing (invite other accounts to a workspace)
- Mobile-friendly control surface (PWA behaviors)
- Dedicated security status page for customer transparency
- Private agent beta with hardware-backed attestation
- AI assistant auto-learning pipeline (post-MVP, opt-in after safety review)

## 🔐 Security Notes
- Agents must never open inbound ports; only HTTPS POST to LunaSentri.
- All tokens are scoping to user/server; revocation available via UI.
- Binaries shipped with checksums; install scripts verify before running.
- Documented best practices for customers: firewall allow-list, non-root service, log auditing.
- Release pipeline requires multi-party approval and publishes signed SBOM per build.

## How to use this file
- Update checkboxes as tasks are completed.
- Link to PRs or docs when applicable.
- Security items can’t be moved to "Done" without confirming docs + guardrails.
