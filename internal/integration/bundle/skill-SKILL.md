---
name: reponerve
description: >-
  Use when the user asks any question about this codebase, architecture, symbols,
  planning, onboarding, or changes — especially if .reponerve/ exists. Works in AI
  chat WITHOUT MCP: run reponerve ask --json and answer from the JSON envelope.
trigger: /reponerve
---

# /reponerve

Evidence-backed repository context for AI chat. **MCP is optional.** When MCP is off, run `reponerve ask` in the terminal with `--json` and answer from the envelope — same contract as MCP tools.

## Usage

```text
/reponerve ask "Why do we use SQLite?"
/reponerve plan "Add OAuth login"          # pasted task
/reponerve onboard
/reponerve explain internal/mcp/server.go
```

Equivalent CLI:

```bash
reponerve ask "<question>" --json
reponerve plan "<task>" --json
reponerve onboard --json
reponerve explain-file "<path>" --json
```

---

## What You Must Do When Invoked

Follow these steps in order. **Do not skip RepoNerve and grep the repo.**

### Step 1 — Ensure RepoNerve memory exists

```bash
test -f .reponerve/memory.db || (reponerve init && reponerve scan)
```

If `reponerve` is not found, tell the user to run `go install ./cmd/reponerve` or install the release binary.

### Step 2 — Run the matching CLI command with `--json`

| User intent | Command |
| --- | --- |
| Any question | `reponerve ask "<question>" --json` |
| Pasted task / plan | `reponerve plan "<task>" --json` |
| Day one / new here | `reponerve onboard --json` |
| Explain topic | `reponerve explain "<topic>" --json` |
| Explain file | `reponerve explain-file "<path>" --json` |
| Explain symbol | `reponerve explain-function "<name>" --package <pkg> --json` |
| Impact / what breaks | `reponerve impact "<subject>" --json` |
| Review prep | `reponerve review "<topic>" --json` |

**Always use `--json`** in chat without MCP. It emits the same envelope as MCP: `structured`, `agent`, `formatted`.

### Step 3 — Read the envelope

1. `structured` — facts (`entity_briefings`, plan scope, evidence)
2. `agent` — `completeness`, `must_use_before_edit`, `recommended_next_tools`
3. `formatted` — human summary only

| `agent.completeness` | You must |
| --- | --- |
| `full` | Answer/edit from `structured`; no bulk file reads |
| `partial` | Run `recommended_next_tools` (more CLI commands) before editing |
| `retrieval_only` | Stop — do not answer confidently; run another command |

### Step 4 — Answer or edit from evidence only

- Cite only paths, symbols, ADRs in RepoNerve output
- Missing fact → say "RepoNerve has no evidence for X" and query more
- Homonyms → compare all `entity_briefings`; use `--package`

---

## MCP path (when connected)

If RepoNerve MCP tools are available, prefer them (`ask`, `plan`, `onboard`, …). Same envelope as `--json`.

---

## Anti-hallucination (mandatory)

```text
BAD:  grep → 20 files → guess
GOOD: reponerve ask/plan --json → briefings → 2 scoped files → review
```

---

## Reference

- CLI/MCP map: `.cursor/skills/reponerve/reference.md`
- Contract: `docs/architecture/agent-context-contract.md`
- All IDEs: `docs/ai-chat-integration.md`
