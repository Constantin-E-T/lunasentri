# LunaSentri CTO Execution Plan

> Last reviewed: 2025-10-07
> Guardrails: zero-trust mindset, no unaudited binaries, customer data isolation by default. Review `docs/AGENT_GUIDELINES.md` before assigning or starting any task.

## Mission Focus
- [ ] Deliver multi-tenant monitoring with uncompromised security posture
- [ ] Ship agents that are verifiably safe to install and easy to revoke
- [ ] Maintain developer velocity without eroding auditability or tests

## Milestones & Owners

### 1. Notification Channel Hardening (API + Web)
- [ ] Spec notifier interface with retry + rate limit requirements *(backend lead)*
- [ ] Implement webhook signing, secret hashing, and whitelisting *(backend lead)*
- [ ] Provide webhook config UI with secret rotation + test payload *(frontend lead)*
- [ ] Document incident response for misfires/false positives *(docs owner)*

### 2. Multi-Server & Tenant Isolation
- [ ] Finalize server schema with per-tenant API tokens + RBAC *(backend lead)*
- [ ] Enforce scoped queries in dashboard + server switcher UX *(frontend lead)*
- [ ] Ship agent token lifecycle (create/disable/rotate) *(backend lead)*
- [ ] Write customer runbook: non-root agent, firewalls, rotation cadence *(docs owner)*

### 3. Agent Packaging & Supply Chain Security
- [ ] Publish agent source with reproducible build steps *(agent team)*
- [ ] Provide install scripts that verify SHA256 + signature before execution *(agent team)*
- [ ] Automate release pipeline with SBOM + provenance attestation *(platform)*
- [ ] Monitor binary download metrics + anomaly detection *(platform)*

### 4. Platform Reliability & Observability
- [ ] Reinstate integration tests for metrics/alarm pipeline *(backend lead)*
- [ ] Add synthetic uptime checks per environment *(platform)*
- [ ] Track error budgets + SLO dashboards *(platform)*
- [ ] Run quarterly tabletop for incident response *(leadership)*

## Security Gates (Block Release if Unchecked)
- [ ] Threat model updated with latest feature surface
- [ ] Pen-test findings closed or accepted with compensating controls
- [ ] Code scanning + dependency auditing pass (Snyk, govulncheck, npm audit)
- [ ] Installer checksum posted + verified in CI before publish
- [ ] Access logs reviewed weekly; anomalies triaged

## Operational Cadence
- [ ] Monday standup: update roadmap status + unblockers
- [ ] Wednesday security sync: review alerts, new CVEs, agent telemetry
- [ ] Friday ship-review: confirm docs/tests for shipped work, tick checkboxes above

## Notes
- Roadmap items in `docs/ROADMAP.md` feed into this plan; sync weekly.
- Treat security items as product features—define owners, due dates, and success metrics.
- Detailed feature context lives under `project/context/`; review relevant note before kicking off work.
- Break work into small, verifiable tasks to prevent drift or speculative implementation.
