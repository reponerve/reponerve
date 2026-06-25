# Development Intelligence PRD

Version: v1.0

Status: Historical (superseded — v1.0.0 shipped 2026-06-18; latest **v1.5.1**)

Issue: ISSUE-057

> **Superseded for current status.** See `docs/product/implementation-status.md` and `docs/README.md`.

See also:

* `docs/roadmap/development-intelligence-prd.md`
* `docs/architecture/development-intelligence-v1.md`

---

# Overview

Development Experience is the primary user-facing layer of RepoNerve.

It orchestrates Code Intelligence, Repository Intelligence, and Repository-Code links to deliver **Software Understanding** and **Knowledge Transfer** through development workflows.

RepoNerve serves as a **software memory system**. See `docs/roadmap/v1.0-prd.md` for authoritative v1.0 goals.

---

# Knowledge Transfer

Development Experience is how preserved knowledge reaches users:

| Audience | Workflow |
| --- | --- |
| New developers | `explain`, `ask` |
| Existing developers | `explain`, `plan` |
| Reviewers | `review` |
| Architects | `impact`, `plan` |
| Engineering managers | `ask` (ownership) |
| AI coding agents | All commands + MCP |

The objective for AI agents is understanding — not retrieval alone. Agents obtain code, repository, ownership, architectural, and change context before implementation.

---

# Developer Workflows

Development Experience supports the workflows developers perform every day.

## Explain

Understand how code works and why it exists.

```bash
reponerve explain "metadata panel"
reponerve explain-file "internal/agent/search/service.go"
reponerve explain-function "Search"
```

Output combines:

**Code Context**

* Files, packages, types, functions, APIs
* Call graph, symbol dependencies

**Repository Context**

* Decisions, facts, events
* Ownership, expertise, reviewers
* Impact, change plans

---

## Ask

Answer natural language repository questions.

```bash
reponerve ask "Who created metadata panel?"
reponerve ask "Why are we using Redis?"
reponerve ask "Who owns authentication?"
reponerve ask "What depends on user-service?"
```

---

## Plan

Prepare for proposed development work.

```bash
reponerve plan "Add OAuth login"
reponerve plan "Add audit logging"
reponerve plan "Introduce rate limiting"
```

Output:

* Impacted areas
* Relevant decisions and facts
* Owners and reviewers
* Suggested workflow
* Suggested starting points

---

## Impact

Analyze change or removal impact.

```bash
reponerve impact "user-service"
reponerve impact "metadata panel"
reponerve impact "Redis"
```

Output:

* Impacted decisions, facts, events
* Code dependencies
* Dependent areas
* Owners

---

## Review

Prepare review guidance.

```bash
reponerve review "metadata panel"
reponerve review "authentication"
```

Output:

* Recommended reviewers
* Required expertise
* Affected areas
* Related knowledge

---

# AI Coding Agent Workflows

AI agents consume Development Intelligence through CLI and MCP.

## Agent Onboarding

Agent asks: "What is this repository? How does authentication work?"

RepoNerve provides combined code and repository context without the agent reading hundreds of files.

## Agent Change Planning

Agent asks: "I need to add OAuth login. What should I know?"

RepoNerve provides impacted areas, relevant decisions, owners, reviewers, and starting points.

## Agent Review Preparation

Agent asks: "Who should review changes to user-service?"

RepoNerve provides reviewer recommendations with evidence.

## Agent Impact Analysis

Agent asks: "What breaks if I change this function?"

RepoNerve provides call graph dependencies and repository impact chain.

---

# Token Reduction Strategy

RepoNerve reduces token consumption by:

1. **Pre-indexing code structure** — agents do not need to read every file to understand symbols and call graphs
2. **Pre-extracting repository memory** — decisions, facts, and ownership are already structured
3. **Orchestrating context assembly** — Development Experience produces focused guidance packages instead of raw repository dumps
4. **Evidence-backed filtering** — only relevant entities are included, ranked by upstream authority not LLM heuristics

Target outcome:

An agent performing "Add OAuth login" receives a structured plan with impacted areas, relevant decisions, owners, and starting points — not thousands of lines of source code.

---

# Context Assembly Strategy

Development Experience assembles context in layers:

```text
Input (natural language topic)
    ↓
Topic Resolution
    ├── Repository Search (repository entities)
    └── Code Intelligence (code entities)
    ↓
Authority Services
    ├── Code Intelligence → Code Context
    └── Repository Intelligence → Repository Context
    ↓
Development Guidance Output
```

Rules:

* Resolve topics before orchestration
* Preserve evidence from every upstream authority
* Attribute every section to its source service
* Sort all lists deterministically
* Do not introduce new scoring systems

---

# Expected Outputs

## DevelopmentAnswer

For ask queries.

Sections: Question, AnswerType, Summary, Evidence, RelatedEntities, SourceServices.

## DevelopmentExplanation

For explain, explain-file, explain-function.

Sections: Topic, CodeContext, RepositoryContext, RepositoryCodeLinks, Evidence.

No Purpose or History fields.

## DevelopmentPlan

For plan queries.

Sections: Task, ImpactedAreas, RelevantDecisions, RelevantFacts, Owners, Reviewers, SuggestedWorkflow, StartingPoints, Evidence.

## DevelopmentImpactReport

For impact queries.

Sections: Subject, ImpactedDecisions, ImpactedFacts, ImpactedEvents, CodeDependencies, DependentAreas, Owners, Evidence.

## DevelopmentReviewGuide

For review queries.

Sections: Topic, RecommendedReviewers, RequiredExpertise, AffectedAreas, RelatedKnowledge, SuggestedWorkflow, Evidence.

---

# Examples

## Explain metadata panel

```bash
reponerve explain "metadata panel"
```

Expected output structure:

```text
Topic: metadata panel

Code Context
  Files: metadata-panel.tsx, MetadataPanel.tsx
  Packages: components/metadata
  Functions: BuildMetadataPanel, renderMetadataPanel
  Call Graph: [deterministic graph]

Repository Context
  Decisions: "Use component-based metadata UI"
  Facts: "Metadata panel depends on user-service"
  Events: "Introduce Metadata Panel"
  Owners: alice@example.com
  Reviewers: bob@example.com (domain: metadata)

Repository-Code Links
  DECISION_REFERENCES_CODE → metadata-panel.tsx
  DECISION_REFERENCES_CODE → BuildMetadataPanel

Evidence: [...]
```

## Plan Add OAuth login

```bash
reponerve plan "Add OAuth login"
```

Expected output structure:

```text
Task: Add OAuth login

Impacted Areas
  - authentication module
  - session management
  - user-service API

Relevant Decisions
  - "Use JWT for session tokens"

Owners
  - alice@example.com (authentication)

Reviewers
  - bob@example.com (security domain)

Suggested Workflow: change_preparation

Starting Points
  - internal/auth/service.go
  - ADR-012: Authentication Strategy

Evidence: [...]
```

---

# Architectural Constraints

Development Experience must:

* Reuse all existing Repository Intelligence services
* Reuse Code Intelligence for code structure questions
* Remain deterministic and evidence-backed
* Keep CLI and MCP thin

Development Experience must not:

* Duplicate Discovery, Learning, Reviewers, or Change Planning
* Duplicate code parsing or call graph analysis
* Introduce LLM-required routing
* Introduce new intelligence scoring systems

---

# Release Criteria

| Capability | Required |
| --- | --- |
| Repository Intelligence | ✅ Complete |
| Code Intelligence | ✅ Required |
| Development Experience | ✅ Required |

RepoNerve v1.0 is not complete until all three are delivered.

Tracked in ISSUE-057.
