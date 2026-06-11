# ARCH-001 — Architecture Realignment For Software Understanding

Status: Approved

Milestone: v1.0

Blocks:

- Architecture approval for ISSUE-057 implementation
- Final v1.0 architecture sign-off

Related:

- ISSUE-057 — Code Intelligence & Development Experience
- `docs/architecture/issue-057-architecture.md`
- `docs/architecture/architecture-overview.md` (v1.1)

---

# Objective

Realign RepoNerve architecture documentation with the Software Understanding product model.

The implementation is sound. The architecture overview still reflects the earlier Repository Memory platform. This task updates documentation so architectural pillars match the product RepoNerve is building for v1.0.

---

# Problem

Current `docs/architecture/architecture-overview.md` still models:

```text
Repository → Scanner → Ingestion → Memory Extraction → Memory Store → Query Engine
```

This was correct when RepoNerve was primarily a repository memory platform.

RepoNerve is now a **software understanding platform** built around knowledge preservation. The architecture must explicitly model:

```text
Knowledge Preservation
Repository Intelligence
Code Intelligence
Repository-Code Linking
Feature Understanding
Development Experience
Software Understanding
```

---

# Scope

Update the following documents:

| Document | Changes |
| --- | --- |
| `docs/architecture/architecture-overview.md` | Primary realignment — system diagram, pillars, subsystems, success criteria |
| `docs/vision/vision.md` | Explicit architectural pillars |
| `docs/vision/mission.md` | Mission alignment with architecture |
| `docs/architecture/agent-native-repository-intelligence.md` | Full pillar model for agent consumption |

Out of scope:

* Code implementation
* Authority boundary changes
* ISSUE-057 sequencing changes
* Package renames in source code

---

# Required Architectural Changes

## 1. Mission Statement (top of architecture overview)

Add:

> RepoNerve is a software understanding platform built around knowledge preservation. Its purpose is to ensure that software understanding survives beyond individual contributors and remains accessible to both humans and AI systems.

## 2. System Overview Diagram

Replace memory-only pipeline with dual-intelligence model:

```text
Repository
        │
        ▼
┌─────────────────────┐
│ Repository Scanner  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Ingestion Pipeline  │
└───────┬─────┬───────┘
        │     │
        ▼     ▼
┌─────────────────┐   ┌─────────────────┐
│ Repository Int. │   │ Code Int.       │
└────────┬────────┘   └────────┬────────┘
         │                     │
         └─────────┬───────────┘
                   ▼
        ┌───────────────────┐
        │ Repository-Code   │
        │ Linking           │
        └─────────┬─────────┘
                  ▼
        ┌───────────────────┐
        │ Development Exp.  │
        └─────────┬─────────┘
                  ▼
        ┌───────────────────┐
        │ Software          │
        │ Understanding     │
        └───────────────────┘
```

## 3. Knowledge Preservation Layer

Elevate from philosophy to subsystem.

Responsible for durable storage of:

* Memory (Decisions, Facts, Events)
* Ownership
* Context
* Code entities and relationships
* Repository-Code links

## 4. Repository-Code Linking

New core subsystem section.

Cross-authority links between repository entities and code entities (e.g. Decision ADR-004 → oauth.go, AuthService, LoginHandler).

Required for Development Experience.

See `docs/architecture/issue-057-architecture.md` for link types and storage.

## 5. Feature Understanding

New architectural capability section.

```text
Feature → Code → Ownership → Decisions → Impact
```

v1.0 delivers partial support through Development Experience topic resolution. Not a separate authority — orchestrated across Repository Intelligence, Code Intelligence, and Repository-Code links.

## 6. Understanding Engine

Evolve Query Engine concept to Understanding Engine.

Responsibilities:

* Repository Intelligence retrieval
* Code Intelligence retrieval
* Repository-Code traversal
* Development context assembly
* Evidence collection

Existing Query Engine remains the repository-memory retrieval implementation within this layer.

## 7. Success Criteria

Replace infrastructure-only criteria with mission-aligned criteria:

* Developers can understand unfamiliar repositories
* Knowledge survives contributor turnover
* AI systems require less repository exploration
* Repository and code context remain connected
* Development guidance is evidence-backed

---

# Acceptance Criteria

ARCH-001 is complete when:

1. `docs/architecture/architecture-overview.md` models all v1.0 architectural pillars
2. System overview diagram reflects dual-intelligence architecture
3. Knowledge Preservation is documented as a subsystem
4. Repository-Code Linking has a dedicated architecture section
5. Feature Understanding is documented as an architectural capability
6. Understanding Engine replaces Query Engine as the conceptual retrieval layer
7. Success criteria align with Software Understanding mission
8. `docs/vision/vision.md`, `docs/vision/mission.md`, and `docs/architecture/agent-native-repository-intelligence.md` are aligned
9. No authority boundary changes introduced

---

# Constraints

* Documentation only — no code changes
* Do not alter ISSUE-057 implementation order
* Preserve Memory First as technical ingestion principle; Understanding First as product principle
* Repository Intelligence and Code Intelligence remain independent authorities

---

# Review

| Reviewer | Status | Date |
| --- | --- | --- |
| Architecture | Pending | |
| Product | Pending | |
