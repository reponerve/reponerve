# Development Experience Contracts

Status: Draft

Related Milestone:

* ISSUE-057 — Code Intelligence & Development Experience

Related:

* `docs/architecture/issue-057-architecture.md`
* `docs/examples/development-experience.md`

---

# Objective

Define the public behavior of Development Experience.

Development Experience is the primary user-facing layer of RepoNerve — how humans and AI systems consume Software Understanding.

Development Experience orchestrates Code Intelligence and Repository Intelligence into knowledge transfer workflows: ask, explain, plan, impact, review.

Feature Understanding is a v1.0 requirement: resolving feature topics to Code → Ownership → Decisions → Impact.

Development Experience does not generate intelligence.

Development Experience does not generate free-form summaries.

Rationale and history appear only as structured Decisions, Facts, and Events in RepositoryContext.

---

# Architectural Rule

Development Experience is an orchestration layer.

It may consume:

* Code Intelligence
* Repository Intelligence
* Repository-Code Relationships

It may not:

* Generate repository intelligence
* Generate code intelligence
* Generate Purpose or History narrative fields
* Re-score entities
* Re-rank discovery results

---

# Unified Explanation Model

Explain combines repository and code understanding through repository-code links:

```text
Repository Entity
        ↓
Repository-Code Link
        ↓
Code Entity
```

RepositoryContext and CodeContext must be connected through `RepositoryCodeLinks` when links exist.

---

# Command: ask

Example:

```bash
reponerve ask "Who created metadata panel?"
```

Behavior:

Answer repository and development questions using upstream authorities only.

Data Sources:

* Repository Search
* Ownership Intelligence
* Repository Q&A
* Code Intelligence (when dependency questions involve code)
* Repository-Code Links

Response:

* AnswerType
* Summary (structured from upstream — not free-form)
* Related entities
* Evidence
* Source services

---

# Command: explain

Example:

```bash
reponerve explain "metadata panel"
```

Behavior:

Provide combined repository and code understanding.

Response Sections:

CODE CONTEXT

* Modules
* Files
* Packages
* Structs
* Interfaces
* Type Aliases
* Functions
* Methods
* Endpoints
* Call graph
* Dependencies

REPOSITORY CONTEXT

* Decisions
* Facts
* Events
* Owners
* Expertise
* Reviewers
* Impact
* Change plans

REPOSITORY-CODE LINKS

* Connected repository entities and code entities with evidence

No Purpose or History fields.

Rationale and evolution appear as structured Decisions, Facts, and Events.

---

# Command: explain-file

Example:

```bash
reponerve explain-file "internal/agent/search/service.go"
```

Response:

* Module
* File
* Package
* Imports
* Structs
* Interfaces
* Functions
* Methods
* Endpoints
* Related repository entities via repository-code links

---

# Command: explain-function

Example:

```bash
reponerve explain-function "Search"
```

Response:

* Function or method entity
* Package
* Module
* File
* Callers
* Callees
* Related repository entities via repository-code links

---

# v1.0 Explain Commands

All required for v1.0 release:

| Command | Resolves |
| --- | --- |
| `explain-struct` | struct entity |
| `explain-interface` | interface entity |
| `explain-type` | type_alias entity |

---

# Command: plan

Example:

```bash
reponerve plan "Add OAuth Login"
```

Behavior:

Prepare implementation guidance from upstream authorities.

Response Sections:

* Impacted areas (code + repository)
* Relevant decisions
* Relevant facts
* Owners
* Reviewers
* Suggested workflow
* Starting points
* Repository-code links

---

# Command: impact

Example:

```bash
reponerve impact "user-service"
```

Behavior:

Assess modification impact across repository and code graphs.

Response:

* Impacted decisions, facts, events
* Code dependencies
* Dependent areas
* Owners
* Repository-code links
* Evidence

---

# Command: review

Example:

```bash
reponerve review "metadata panel"
```

Behavior:

Prepare review guidance from upstream authorities.

Response:

* Recommended reviewers
* Required expertise
* Affected code areas
* Affected repository knowledge
* Repository-code links
* Suggested workflow: `review_preparation`

---

# Evidence Preservation Rule

Development Experience must preserve evidence.

Repository evidence remains owned by Repository Intelligence.

Code evidence remains owned by Code Intelligence.

Link evidence remains owned by Repository-Code Link store.

Evidence may be displayed.

Evidence may not be modified.

---

# Output Composition Rule

Every command must identify:

CODE CONTEXT

and

REPOSITORY CONTEXT

when both are available.

Explain, Plan, Impact, and Review must include REPOSITORY-CODE LINKS when cross-authority links exist.

RepoNerve's primary differentiator is combining both views through deterministic links.

---

# Acceptance Criteria

Development Experience is complete only when:

* ask works
* explain works
* explain-file works
* explain-function works
* plan works
* impact works
* review works
* explain output contains no Purpose or History fields
* repository-code links connect RepositoryContext and CodeContext when evidence exists

All commands must consume existing intelligence services.

No duplicate intelligence systems are permitted.

See `docs/examples/development-experience.md` for end-to-end acceptance examples.
