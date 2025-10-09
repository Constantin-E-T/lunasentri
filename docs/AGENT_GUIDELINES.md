# Agent Operating Guidelines

> Security is the product. Every task must start by checking context, verifying assumptions, and using least-privilege patterns.

## Core Principles

- **Read first**: review `docs/PLAN.md`, relevant context files under `project/context/`, and the file history before editing.
- **Validate facts**: confirm behaviours in code/tests and cite official language or framework documentation; never rely on memory when uncertain.
- **No speculation**: if a detail is unclear, explicitly say so, ask for confirmation, or check authoritative references—never invent behaviour.
- **Bias for minimal scope**: break large work into small, reviewable steps and stop if requirements become unclear.
- **Security posture**: never weaken auth, secret handling, or sandboxing; surface concerns immediately.
- **Audit trail**: document commands run, decisions made, and follow-up items in the corresponding agent log.
- **State the plan**: restate the task scope in your own words before coding so reviewers know you understood it.
- **Right-sized tests**: prefer focused test files over monoliths; split suites when they exceed a few hundred lines or mix concerns.

## Execution Checklist

1. Confirm the task scope with the CTO note or context file and restate it in your own words.
2. Inspect existing files, tests, and recent commits before touching code.
3. Use established helpers/utilities; avoid reimplementing solved problems.
4. Run the smallest meaningful command set (targeted tests first, full suite before handing off).
5. Keep test files manageable—if a test grows beyond ~200 lines or multiple domains, split it.
6. Summarise outcomes, residual risks, and verification steps when you update your log.
7. Capture open questions or blockers in the relevant context file under an `Open Questions` section.
8. When logs grow beyond a sprint, move older entries into `project/logs/archive/` and link the archive from the active log.
9. Run terminal commands individually whenever possible—avoid chaining with `&&`/pipes unless verification requires it.

Keep these guardrails visible while working—repeat them if the task spans multiple turns.

## Prompting & Knowledge Hygiene

- **Lead with context**: include the scoped objective, relevant files, and acceptance criteria in every prompt so assistants stay focused.
- **Define verification**: specify the commands/tests that must run; require assistants to call those commands or document why they cannot.
- **Enforce fact-checking**: instruct assistants to consult official docs, release notes, or Go standard library references when behaviour is unclear, and to cite what they checked.
- **Constrain output**: ask for concise diffs or summaries to conserve tokens; request code blocks only for changed snippets.
- **Disallow dreaming**: remind assistants to respond with “unsure” when a fact cannot be verified; they must propose how to confirm it rather than guessing.
- **Update prompts as you learn**: when requirements or APIs change, refresh the prompt with that information before assigning follow-up tasks.

## Local Development Configuration

### Environment Flags

#### `LOCAL_HOST_METRICS`

**Default**: `false`  
**Purpose**: Controls whether the API server collects metrics from its own host machine.

- When `false` (production/multi-machine mode): The server does NOT collect local system metrics. This is the default behavior for multi-machine monitoring where metrics are ingested from remote agent clients via API.
- When `true` (local development mode): The server collects and exposes metrics from its own host. This is useful for local testing without setting up separate agents.

**Local Development Usage**:

```bash
export LOCAL_HOST_METRICS=true
cd apps/api-go
go run ./cmd/api
```

**Security Note**: Setting `LOCAL_HOST_METRICS=true` in production environments can leak host details about the API server itself. This flag should only be enabled for local development and testing purposes.

### Authentication Architecture

The current authentication mechanism uses **API key authentication** for machine-to-server communication:

- Each monitored machine receives a unique API key during registration
- API keys are hashed (SHA-256) before storage
- Agents include the API key in request headers for authentication

**Roadmap**: Mutual TLS (mTLS) authentication is planned for enhanced security. See `docs/roadmap/MULTI_MACHINE_MONITORING.md` for details on the upcoming agent implementation.

For agent ingestion endpoints and the complete multi-machine architecture, refer to:

- `docs/roadmap/MULTI_MACHINE_MONITORING.md` - Full Phase 1 & 2 specifications
- `docs/security/AGENT_SECURITY_ARCHITECTURE.md` - Security model and API-key handling
