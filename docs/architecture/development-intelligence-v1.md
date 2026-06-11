# Development Experience V1

Status: Draft

Version: v1.0

Issue: ISSUE-057

Related: `docs/roadmap/v1.0-prd.md`, `docs/architecture/development-experience-contracts.md`

---

# Overview

Development Experience is the primary user-facing layer of RepoNerve.

It delivers **Software Understanding** and **Knowledge Transfer** by orchestrating Code Intelligence, Repository Intelligence, and Repository-Code links into development workflows.

RepoNerve serves as a **software memory system** — preserving understanding that would otherwise be lost through contributor turnover, architectural evolution, and documentation drift.

```text
Knowledge Preservation → Software Memory → Software Understanding
```

---

# Software Understanding Model

Development Experience delivers:

```text
Code Understanding
    +
Repository Understanding
    +
Ownership Understanding
    +
Architectural Understanding
    +
Change Understanding
    +
Historical Understanding
    ═══════════════════════
    Software Understanding
```

---

# Core Product Goal

A developer or AI coding agent should not need to repeatedly scan a repository to understand how code works, why it exists, who owns it, what depends on it, and how to change it safely.

Development Experience provides this through `ask`, `explain`, symbol-level explain commands, `plan`, `impact`, and `review`.

For most developers, `reponerve explain authentication` matters more than `reponerve explain-function CreateUser`.

---

# Product Position

```text
Knowledge Preservation          (Core Platform Capability)
    ↓
Software Memory
    ↓
Repository Intelligence
    +
Code Intelligence
    ↓
Repository-Code Linking
    ↓
Feature Understanding
    ↓
Development Experience
    ↓
Software Understanding
```

Authority boundaries between intelligence layers are unchanged.

---

# Knowledge Transfer

Development Experience transfers preserved knowledge to:

- New developers (onboarding)
- Existing developers (unfamiliar systems)
- Reviewers (review context)
- Architects (impact analysis)
- Engineering managers (ownership)
- AI coding agents (development context via CLI and MCP)

Success criteria align with `docs/roadmap/v1.0-prd.md` Goal 6 — Knowledge Transfer.

---

# AI Agent Context

The objective for AI agents is understanding — not retrieval alone.

Agents should obtain code, repository, ownership, architectural, and change context before implementation work, reducing token consumption from repeated exploration.

Delivery mechanisms:

* MCP tools (27 today) — bounded structured responses
* Development Experience CLI — orchestrated context packs
* Context engine — deterministic generate/export
* Compression service — truncation today; graph-aware token budgets (ISSUE-060, v1.0)

See `docs/product/token-economics.md` and `docs/roadmap/v1.0-iteration-plan.md`.

---

# Explain Output Contract

Development Experience combines three layers:

**Code Context** — modules, files, packages, structs, interfaces, type aliases, functions, methods, endpoints, call graph

**Repository Context** — decisions, facts, events, ownership, expertise, reviewers, impact, change plans

**Repository-Code Links** — deterministic cross-authority connections (e.g. Decision ADR-004 → oauth.go, AuthService)

No Purpose or History narrative fields. Rationale appears as structured Decisions, Facts, and Events.

---

# Core Engines

1. Natural Language Question Answering — `ask`
2. Repository and Feature Explanation — `explain`
3. Code Explanation — `explain-file`, `explain-function`, `explain-struct`, `explain-interface`, `explain-type`
4. Development Planning — `plan`
5. Development Impact — `impact`
6. Review Preparation — `review`

---

# Authority Boundaries

Unchanged. Development Experience orchestrates. It does not replace Code Intelligence or Repository Intelligence authorities.

Repository-Code Linking connects both authorities. Feature Understanding is orchestrated — not a separate authority.

See `docs/architecture/issue-057-architecture.md`.

---

# CLI Contract

```bash
reponerve ask "Who owns billing?"
reponerve explain "authentication"
reponerve explain-file "internal/auth/oauth.go"
reponerve explain-function "LoginHandler"
reponerve explain-struct "AuthService"
reponerve explain-interface "Authenticator"
reponerve explain-type "HandlerFunc"
reponerve plan "Add OAuth login"
reponerve impact user-service
reponerve review "metadata panel"
```

---

# Feature Understanding (v1.0)

Humans think in features, not files. First-class v1.0 goal:

```text
Feature → Code → Ownership → Decisions → Impact
```

Delivered through Development Experience topic resolution as part of ISSUE-057.

---

# Release Impact

Software Understanding is blocked until ISSUE-057 completes all v1.0 scope: Code Intelligence, Repository-Code Linking, Feature Understanding, and Development Experience.

There is no separate post-v1.0 product release.
