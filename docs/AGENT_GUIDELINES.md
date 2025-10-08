# Agent Operating Guidelines

> Security is the product. Every task must start by checking context, verifying assumptions, and using least-privilege patterns.

## Core Principles
- **Read first**: review `docs/PLAN.md`, relevant context files under `project/context/`, and the file history before editing.
- **Validate facts**: confirm behaviours in code/tests and prefer official framework or language documentation over guesses.
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

Keep these guardrails visible while working—repeat them if the task spans multiple turns.
