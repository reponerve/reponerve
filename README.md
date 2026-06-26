# RepoNerve

[![Release](https://img.shields.io/github/v/release/reponerve/reponerve?label=release)](https://github.com/reponerve/reponerve/releases)
[![npm](https://img.shields.io/npm/v/reponerve?label=npm)](https://www.npmjs.com/package/reponerve)
[![License](https://img.shields.io/github/license/reponerve/reponerve)](LICENSE)
[![CI](https://github.com/reponerve/reponerve/actions/workflows/test.yml/badge.svg)](https://github.com/reponerve/reponerve/actions/workflows/test.yml)
[![MCP tools](https://img.shields.io/badge/MCP-49_tools-blue)](docs/mcp/compatibility-matrix.md)
[![Local-first](https://img.shields.io/badge/cloud-none-local--first-22c55e)]()

> **Local-first software understanding** — tell agents and developers not just *where* code is, but *why* it exists, *who* owns it, and *what breaks* if you change it. Every answer is backed by repository evidence.

**Latest release:** [`v1.5.1`](docs/releases/v1.5.1.md) · [Website](https://reponerve.github.io/) · [Documentation](docs/README.md) · [Install guide](docs/install.md)

### Explain it to your team (30 seconds)

> **RepoNerve** scans our repository once — git, ADRs, and code structure — and builds local **software memory**. Developers and AI agents query that memory for *why* something exists, *who* owns it, and *what breaks* if we change it, instead of grepping the repo every session. Run `reponerve init` and `reponerve scan` in any project; it plugs into **Cursor**, **Copilot**, and other MCP hosts. No cloud required. No separate discipline skill packs — reuse, review, and ship-check habits ship with `init`.

---

## Demo

![RepoNerve setup and use — install, init, scan, understand, plan, reuse-check, review](docs/assets/reponerve-demo.gif)

The recording follows the **exact steps below** on a real repository. Regenerate: `vhs docs/assets/demo.tape` · Guide: [`docs/assets/README.md`](docs/assets/README.md)

---

## Setup and use (exact steps)

Do this **once per git repository**. Total time: about 2 minutes.

### Step 1 — Install RepoNerve

Pick one install path, then verify the binary:

```bash
# npm (Node 18+)
npm install -g reponerve

# macOS / Linux — no Node required
curl -fsSL https://raw.githubusercontent.com/reponerve/reponerve/main/scripts/install.sh | bash

# Go developers
go install github.com/reponerve/reponerve/cmd/reponerve@v1.5.1
```

```bash
reponerve --version
```

More install options (Homebrew, Windows, pinned versions): [`docs/install.md`](docs/install.md)

### Step 2 — Set up your repository

From the **root** of the project you want to understand:

```bash
cd /path/to/your-repo
reponerve init
reponerve scan
```

| Command | What it does |
| --- | --- |
| `init` | Creates `.reponerve/`, SQLite memory, MCP config (Cursor / VS Code / Continue), agent skill, and discipline rules |
| `scan` | Ingests git history, ADRs, and code into memory — **no LLM required** |

Optional — re-scan automatically after each commit:

```bash
reponerve hook install
```

### Step 3 — Verify setup

```bash
reponerve doctor
```

Fix anything marked `warn` or `fail` (usually run `reponerve scan` again).

### Step 4 — Understand the repository

Use these before editing code or pasting a ticket into your agent:

```bash
# Day-one orientation (decisions + repo map)
reponerve onboard

# Or ask a specific question
reponerve ask "What does this repository do?"

# Explain a feature, file, or symbol
reponerve explain "authentication"
reponerve explain-file "internal/auth/handler.go"
reponerve explain-function "HandleLogin" --package auth
```

Add `--format compact --token-budget 1500` for shorter output in scripts or agents.

### Step 5 — Plan and ship a change

When you have a task or are preparing a PR:

```bash
# 1. Scope the work (starting files + steps)
reponerve plan "Add OAuth login"

# 2. Reuse existing code before writing new code
reponerve reuse-check "add OAuth middleware"

# 3. Check impact of risky changes
reponerve impact "OAuth login"

# 4. Pre-merge review + discipline checks
reponerve review "OAuth login"

# 5. Ship readiness (blockers + advisories)
reponerve ship-check "OAuth login"

# 6. PR evidence pack (changed files)
reponerve pr-context --file internal/auth/oauth.go
```

Agents should use `--json` (same envelope as MCP): read `structured` → `agent` → `formatted`.

### Step 6 — Use with AI chat

After `init`, **restart MCP** in your IDE (Cursor → Settings → Tools & MCP).

| Mode | How |
| --- | --- |
| **MCP on** | Ask in natural language — the agent calls RepoNerve tools automatically |
| **MCP off** | Agent runs `reponerve ask "..." --json` in the terminal (same evidence) |

Example prompts:

- "Onboard me to this repo"
- Paste a ticket → "Where should I start?"
- "Why do we use SQLite?"
- "What can I reuse for rate limiting?"
- "Review my OAuth change"

| IDE | Guide |
| --- | --- |
| Any IDE / LLM | [`docs/ai-chat-integration.md`](docs/ai-chat-integration.md) |
| Cursor | [`docs/cursor-integration.md`](docs/cursor-integration.md) |
| VS Code + Copilot | [`docs/copilot-chat-integration.md`](docs/copilot-chat-integration.md) |

Refresh IDE files later: `reponerve integrate`

---

## Copy-paste demo script

Run in order on any git repository (matches the GIF above):

```bash
# Step 1 — Install (verify)
reponerve --version

# Step 2 — Set up repository
reponerve init
reponerve scan

# Step 3 — Verify
reponerve doctor

# Step 4 — Understand
reponerve onboard --format compact --token-budget 400

# Step 5 — Plan and ship a change
reponerve plan "Add webhook notifications" --format compact --token-budget 320
reponerve reuse-check "add webhook" --format compact --token-budget 320
reponerve review "webhook notifications" --format compact --token-budget 280
```

---

RepoNerve scans your repository once (git history, ADRs, code structure) and builds **software memory** — a local SQLite knowledge base your team and AI agents query instead of re-reading the whole codebase every session.

It is **not** an autonomous coding agent. It is the **understanding layer** that sits under Cursor, Copilot, Claude, or any terminal workflow.

```text
You or your AI agent asks a question
        ↓
RepoNerve returns evidence (decisions, symbols, owners, impact)
        ↓
Implement with confidence — less grep, fewer wrong edits, lower token cost
```

**Philosophy:** Understanding first. Evidence second. AI third.

---

## What questions does it answer?

| You ask… | RepoNerve helps with… |
| --- | --- |
| "What does this repo do?" | Orientation, decisions, architecture |
| "Explain this feature / file / function" | Code + why + linked ADRs |
| "Who owns authentication?" | Contributors and expertise |
| "What breaks if I change X?" | Impact and dependencies |
| "Where do I start for this ticket?" | Scoped `plan` with starting points |
| "Can we ship this?" | `ship-check` blockers and advisories |
| "What can I reuse?" | `reuse-check` before writing new code |

---

## What is RepoNerve?

## Command cheat sheet

| Goal | Command |
| --- | --- |
| Orientation | `reponerve onboard` |
| Question | `reponerve ask "..."` |
| Feature / topic | `reponerve explain "metadata panel"` |
| File | `reponerve explain-file path/to/file.go` |
| Symbol | `reponerve explain-function Name --package pkg` |
| Task planning | `reponerve plan "Add OAuth"` |
| Impact | `reponerve impact "user-service"` |
| Pre-merge review | `reponerve review "topic"` |
| Ship readiness | `reponerve ship-check "topic"` |
| Reuse existing code | `reponerve reuse-check "intent"` |
| Browse graph (local UI) | `reponerve explore --serve` |
| MCP for agents | `reponerve mcp` |

**49 MCP tools** mirror these commands. Full CLI reference: [`docs/architecture/cli-reference-v1.md`](docs/architecture/cli-reference-v1.md)

---

## How it works

```text
Repository (git, ADRs, source)
        ↓ scan (deterministic)
Software memory (.reponerve/memory.db)
        ↓ query
Development Experience (ask, explain, plan, review, …)
        ↓
CLI · MCP · Explore UI
        ↓
Developers and AI agents
```

**Layers:**

- **Repository Intelligence** — why (decisions, facts, events, ownership)
- **Code Intelligence** — how (symbols, call graphs; Go + 19 Tree-sitter languages)
- **Repository–code linking** — ADRs and events tied to real files and symbols
- **Development Experience** — the commands and MCP tools you use daily

Details: [`docs/architecture/architecture-overview.md`](docs/architecture/architecture-overview.md)

---

## Why RepoNerve?

| Problem | RepoNerve approach |
| --- | --- |
| Agents re-grep every session | Scan once, query cheaply |
| "Why does this exist?" lost when people leave | ADRs + git → durable memory |
| Generic agent rules ignore your repo | Repo-adaptive discipline policy after `scan` |
| Premium LLM tokens wasted on exploration | Bounded evidence packs (RFC-001) |

Compared to code-graph-only tools: RepoNerve owns the **why**, **who**, and **what breaks** layer with mandatory evidence.

See [`docs/product/market-positioning.md`](docs/product/market-positioning.md) and [`docs/product/token-economics.md`](docs/product/token-economics.md).

---

## Current status

**v1.0.0** shipped 2026-06-18. **Latest:** **v1.5.1**.

| Area | Status |
| --- | --- |
| Memory, graph, ownership | ✅ Shipped |
| Code intelligence (20 languages) | ✅ Shipped |
| Development Experience + 49 MCP tools | ✅ Shipped |
| Native discipline, reuse, ship-check, PR context | ✅ Shipped |
| Doctor, scoped scan, npm, explore UI | ✅ Shipped |

Honest snapshot: [`docs/product/implementation-status.md`](docs/product/implementation-status.md)  
Release line: [`docs/releases/versioning.md`](docs/releases/versioning.md)  
Post-1.0 scope: RFC-gated — [`docs/roadmap/v1.x-backlog.md`](docs/roadmap/v1.x-backlog.md)

---

## Documentation

| Start here | |
| --- | --- |
| [Docs index](docs/README.md) | Full documentation map |
| [AI chat integration](docs/ai-chat-integration.md) | Any IDE or LLM |
| [Install](docs/install.md) | All install paths |
| [Vision](docs/vision/vision.md) | Product direction |
| [Greenfield guide](docs/product/greenfield-guide.md) | New projects from day one |
| [Demo assets](docs/assets/README.md) | Record and add the README demo GIF |
| [Contributing](docs/governance/contribution-guide.md) | Developer setup |

---

## License

Apache License 2.0 — see [LICENSE](LICENSE).

Copyright © 2026 RepoNerve Contributors
