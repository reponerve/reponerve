# Development Experience Examples

Version: v1.0

Status: Draft

Issue: ISSUE-057

Purpose: Acceptance criteria for Code Intelligence & Development Experience.

Architecture: `docs/architecture/issue-057-architecture.md`

---

# How To Use This Document

Each example defines:

1. **Command** — the developer-facing invocation
2. **Orchestration** — which authorities are consulted
3. **Expected Output** — the required output structure
4. **Acceptance Rules** — what must be true for ISSUE-057 to pass

All outputs must be:

* Deterministic
* Evidence-backed
* Source-attributed
* Reproducible for the same repository state

---

# Ask Examples

## Example 1 — Ownership

### Command

```bash
reponerve ask "Who owns authentication?"
```

### Orchestration

```text
Question
    ↓
Topic Resolution (Repository Search: "authentication")
    ↓
Ownership Intelligence
    ↓
DevelopmentAnswer
```

### Expected Output

```text
Question: Who owns authentication?

Answer Type: ownership

Summary:
  Primary owner for authentication domain: alice@example.com
  Supporting contributors: bob@example.com, carol@example.com

Related Entities:
  - CONTRIBUTOR / repo-1:alice@example.com / alice@example.com
  - EXPERTISE / repo-1:alice@example.com:authentication / authentication
  - DECISION / repo-1:decision-auth-jwt / Use JWT for session tokens

Evidence:
  - source: ownership
    type: expertise_match
    payload: {"domain":"authentication","score":42,"file_count":18}
  - source: ownership
    type: contributor_activity
    payload: {"email":"alice@example.com","commit_count":34}

Source Services:
  - repository_search
  - ownership_intelligence
```

### Acceptance Rules

* Topic resolved without requiring entity IDs in user input
* Every owner claim includes ownership evidence
* Ordering: expertise score DESC (from ownership authority), email ASC

---

## Example 2 — Decision Rationale

### Command

```bash
reponerve ask "Why are we using Redis?"
```

### Orchestration

```text
Question
    ↓
Topic Resolution (Repository Search: "Redis")
    ↓
Architectural Guidance + Repository Q&A
    ↓
DevelopmentAnswer
```

### Expected Output

```text
Question: Why are we using Redis?

Answer Type: decision_rationale

Summary:
  Redis was adopted for caching to reduce database load and improve read latency.

Related Entities:
  - DECISION / repo-1:decision-redis-cache / Use Redis Cache
  - INTENT / repo-1:intent-reduce-latency / Reduce Latency
  - FACT / repo-1:fact-api-uses-redis / API Service USES Redis

Evidence:
  - source: memory
    type: decision
    payload: {"id":"repo-1:decision-redis-cache","title":"Use Redis Cache","status":"Accepted"}
  - source: memory
    type: relationship
    payload: {"type":"INTENT_DRIVES_DECISION","from":"repo-1:intent-reduce-latency"}

Source Services:
  - repository_search
  - architectural_guidance
  - repository_qa
```

### Acceptance Rules

* Answer cites at least one decision with evidence
* No unsupported rationale beyond repository memory

---

## Example 3 — Authorship

### Command

```bash
reponerve ask "Who created metadata panel?"
```

### Expected Output

```text
Question: Who created metadata panel?

Answer Type: authorship

Summary:
  Most associated contributor for metadata panel: alice@example.com (12 matching commits)

Related Entities:
  - CONTRIBUTOR / repo-1:alice@example.com / alice@example.com
  - EVENT / repo-1:event-metadata-panel-intro / Introduce Metadata Panel

Evidence:
  - source: memory
    type: commit_authorship
    payload: {"email":"alice@example.com","count":12,"topic":"metadata panel"}

Source Services:
  - repository_search
  - ownership_intelligence
```

---

# Explain Examples

## Example 1 — Topic Explain

### Command

```bash
reponerve explain "metadata panel"
```

### Orchestration

```text
Topic
    ↓
Topic Resolution
    ├── Repository Search
    └── Code Intelligence (symbol/file match)
    ↓
Code Intelligence → Code Context
Repository Intelligence → Repository Context
Repository-Code Link Traversal → RepositoryCodeLinks
    ↓
DevelopmentExplanation
```

### Expected Output

```text
Topic: metadata panel

CODE CONTEXT

Modules:
  - github.com/reponerve/reponerve

Files:
  - internal/ui/metadata/panel.go
  - internal/ui/metadata/form.go

Packages:
  - github.com/reponerve/reponerve/internal/ui/metadata

Structs:
  - MetadataPanel
  - MetadataForm

Functions:
  - NewMetadataPanel()
  - validateMetadataInput()

Methods:
  - (p *MetadataPanel) Render()

Endpoints:
  - GET /api/v1/metadata (http)
  - PUT /api/v1/metadata (http)

Call Graph:
  Render()
    → loadMetadata()
    → validateMetadataInput()
    → renderForm()

Dependencies:
  - internal/service/metadata (IMPORTS)
  - internal/storage/sqlite (DEPENDS_ON)

REPOSITORY CONTEXT

Decisions:
  - Use component-based metadata UI (Accepted)

Facts:
  - Metadata Panel DEPENDS_ON user-service

Events:
  - Introduce Metadata Panel

Owners:
  - alice@example.com (domain: metadata, expertise score: 42)

Expertise:
  - alice@example.com — metadata (18 files)

Reviewers:
  - bob@example.com — metadata domain reviewer

Impact:
  - Changes to metadata panel affect user-service integration

REPOSITORY-CODE LINKS

  - DECISION_REFERENCES_CODE: decision-metadata-ui → internal/ui/metadata/panel.go
  - DECISION_REFERENCES_CODE: decision-metadata-ui → MetadataPanel (struct)
  - EVENT_REFERENCES_CODE: event-metadata-panel-intro → internal/ui/metadata/panel.go
  - FACT_REFERENCES_CODE: fact-metadata-depends-user → MetadataPanel (struct)

Evidence:
  - [code: file_definition, internal/ui/metadata/panel.go:14]
  - [code: calls, internal/ui/metadata/panel.go:48 → loadMetadata]
  - [memory: decision, repo-1:decision-metadata-ui]
  - [link: DECISION_REFERENCES_CODE, decision-metadata-ui → panel.go]
  - [ownership: expertise, alice@example.com:metadata]

Source Services:
  - code_intelligence
  - repository_search
  - repository_code_links
  - context_engine
  - ownership_intelligence
  - architectural_guidance
```

### Acceptance Rules

* Output contains CODE CONTEXT, REPOSITORY CONTEXT, and REPOSITORY-CODE LINKS sections
* No Purpose or History fields — rationale appears as Decisions, Facts, Events
* Every code entity includes file path and line evidence
* Every repository entity includes memory or ownership evidence
* Repository-code links include cross-authority evidence
* No speculative content beyond indexed code and repository memory

---

## Example 2 — Explain File

### Command

```bash
reponerve explain-file "internal/agent/search/service.go"
```

### Expected Output

```text
Topic: internal/agent/search/service.go

CODE CONTEXT

Modules:
  - github.com/reponerve/reponerve

Files:
  - internal/agent/search/service.go

Packages:
  - github.com/reponerve/reponerve/internal/agent/search

Structs:
  - Service

Functions:
  - NewService(...)

Methods:
  - (s *Service) Search(ctx, repositoryID, query)

Call Graph:
  Search()
    → collectMemoryHits()
    → collectOwnershipHits()
    → collectGraphHits()
    → sortHits()

Dependencies:
  - internal/query/storage (IMPORTS)
  - internal/agent/search/models (IMPORTS)

REPOSITORY CONTEXT

Owners:
  - carol@example.com (domain: search)

Decisions:
  - Repository Search must remain deterministic (Accepted)

REPOSITORY-CODE LINKS

  - DECISION_REFERENCES_CODE: decision-search-deterministic → internal/agent/search/service.go
  - DECISION_REFERENCES_CODE: decision-search-deterministic → Service (struct)

Evidence:
  - [code: file_definition, internal/agent/search/service.go:1]
  - [code: method_definition, internal/agent/search/service.go:120]
  - [link: DECISION_REFERENCES_CODE, decision-search-deterministic → service.go]
  - [ownership: expertise, carol@example.com:search]

Source Services:
  - code_intelligence
  - repository_search
  - repository_code_links
  - ownership_intelligence
```

### Acceptance Rules

* File resolved by path without requiring symbol name
* No Purpose field
* Code context is primary; repository context connected via repository-code links

---

## Example 3 — Explain Function

### Command

```bash
reponerve explain-function "Search"
```

### Expected Output

```text
Topic: Search

CODE CONTEXT

Modules:
  - github.com/reponerve/reponerve

Methods:
  - (s *Service) Search(ctx context.Context, repositoryID string, query string)

Defined In File:
  - internal/agent/search/service.go:120 (DEFINED_IN_FILE)

Call Graph:
  Search()
    → collectMemoryHits()
    → collectOwnershipHits()
    → sortHits()

Called By:
  - internal/agent/workflow/service.go: BuildKnowledgeExplorationWorkflow()

REPOSITORY CONTEXT

Decisions:
  - Search must not generate repository intelligence (Accepted)

REPOSITORY-CODE LINKS

  - DECISION_REFERENCES_CODE: decision-search-authority → Search (method)

Evidence:
  - [code: method_definition, internal/agent/search/service.go:120]
  - [code: calls, internal/agent/search/service.go:145 → collectMemoryHits]
  - [link: DECISION_REFERENCES_CODE, decision-search-authority → Search]

Source Services:
  - code_intelligence
  - repository_search
  - repository_code_links
```

### Acceptance Rules

* Symbol resolved deterministically with tie-breaking by qualified name
* No Purpose field
* Method entity type used for receiver functions
* Call graph includes both callees and callers where evidence exists

---

# Plan Examples

## Example 1 — Feature Addition

### Command

```bash
reponerve plan "Add OAuth login"
```

### Orchestration

```text
Task Description
    ↓
Topic Resolution (Repository Search + Code Intelligence)
    ↓
Context Engine
Knowledge Discovery
Learning Paths
Reviewer Recommendations
Change Planning
Workflow Intelligence
    ↓
DevelopmentPlan
```

### Expected Output

```text
Task: Add OAuth login

Impacted Areas:
  - internal/auth/ (package)
  - internal/session/ (package)
  - internal/api/handlers/auth.go (file)
  - user-service (domain)

Relevant Decisions:
  - Use JWT for session tokens (Accepted)
  - Local-first authentication only (Accepted)

Relevant Facts:
  - Auth Service DEPENDS_ON SQLite
  - API Gateway CALLS Auth Service

Owners:
  - alice@example.com (authentication)
  - bob@example.com (security)

Reviewers:
  - bob@example.com (security domain, score: 38)
  - carol@example.com (api domain, score: 22)

Suggested Workflow: change_preparation

Starting Points:
  - internal/auth/service.go
  - docs/architecture/authentication.md
  - ADR: Use JWT for session tokens
  - Learning path step 3: Authentication module

Evidence:
  - [search: match, domain:authentication, score:100]
  - [change_plan: priority:1, entity:DECISION/jwt-tokens]
  - [reviewers: recommendation, bob@example.com:security]

Source Services:
  - repository_search
  - code_intelligence
  - context_engine
  - knowledge_discovery
  - learning_paths
  - reviewer_recommendations
  - change_planning
  - workflow_intelligence
```

### Acceptance Rules

* Plan reuses Change Planning — does not invent new impact logic
* Every impacted area includes evidence
* Suggested workflow references an existing Workflow Intelligence type

---

# Impact Examples

## Example 1 — Service Impact

### Command

```bash
reponerve impact "user-service"
```

### Orchestration

```text
Topic
    ↓
Topic Resolution
    ↓
Knowledge Graph Impact
Agent Impact Analysis
Code Intelligence (symbol dependencies)
Ownership Intelligence
    ↓
DevelopmentImpactReport
```

### Expected Output

```text
Subject: user-service

Impacted Decisions:
  - Use JWT for session tokens
  - Adopt microservice boundaries

Impacted Facts:
  - Metadata Panel DEPENDS_ON user-service
  - API Gateway CALLS user-service

Impacted Events:
  - Introduce user-service

Code Dependencies:
  - internal/ui/metadata/panel.go (IMPORTS internal/service/user)
  - internal/api/handlers/user.go (BELONGS_TO_PACKAGE user-service boundary)

Dependent Areas:
  - metadata panel
  - authentication flow
  - api gateway

Owners:
  - alice@example.com (user-service domain)

Repository-Code Links:
  - FACT_REFERENCES_CODE: fact-metadata-depends-user → internal/service/user/handler.go
  - EVENT_REFERENCES_CODE: event-introduce-user-service → internal/service/user/

Evidence:
  - [graph: impact_chain, depth:2, entity:FACT/metadata-depends-user]
  - [code: depends_on, internal/ui/metadata/panel.go → internal/service/user]
  - [ownership: expertise, alice@example.com:user-service]

Source Services:
  - repository_search
  - knowledge_graph_impact
  - agent_impact_analysis
  - code_intelligence
  - ownership_intelligence
```

### Acceptance Rules

* Impact includes both repository and code dependency chains
* Graph impact reuses existing graph impact service — no new scoring

---

# Review Examples

## Example 1 — Feature Review Preparation

### Command

```bash
reponerve review "metadata panel"
```

### Orchestration

```text
Topic
    ↓
Topic Resolution
    ↓
Reviewer Recommendations
Ownership Intelligence
Workflow Intelligence (review_preparation)
Repository Search
    ↓
DevelopmentReviewGuide
```

### Expected Output

```text
Topic: metadata panel

Recommended Reviewers:
  - bob@example.com (metadata domain, score: 35)
  - alice@example.com (original author, score: 28)

Required Expertise:
  - metadata (UI components)
  - user-service integration

Affected Areas:
  - internal/ui/metadata/
  - internal/service/metadata/
  - user-service API boundary

Related Knowledge:
  - Decision: Use component-based metadata UI
  - Fact: Metadata Panel DEPENDS_ON user-service
  - Event: Introduce Metadata Panel

Suggested Workflow: review_preparation

Evidence:
  - [reviewers: domain_match, bob@example.com:metadata, score:35]
  - [reviewers: authorship, alice@example.com, commits:12]
  - [search: match, entity:DECISION/metadata-ui]

Source Services:
  - repository_search
  - reviewer_recommendations
  - ownership_intelligence
  - workflow_intelligence
```

### Acceptance Rules

* Reviewer recommendations reuse Reviewer Recommendations service
* Every recommended reviewer includes evidence and explanation
* Suggested workflow is `review_preparation`

---

# Cross-Cutting Acceptance Rules

All ISSUE-057 commands must satisfy:

1. **Topic resolution** — natural language or path/symbol input resolved before orchestration
2. **Dual authority** — code questions consult Code Intelligence; repository questions consult Repository Intelligence
3. **Repository-code links** — RepositoryContext and CodeContext connected through deterministic cross-authority links
4. **No narrative fields** — no Purpose or History; rationale appears as Decisions, Facts, Events
5. **Evidence** — every section includes traceable evidence
6. **Source attribution** — `Source Services` lists upstream authorities
7. **Determinism** — identical repository state + input = identical output
8. **Ordering** — all lists sorted deterministically per `docs/architecture/issue-057-architecture.md`
9. **No duplication** — Development Experience orchestrates; it does not reimplement upstream logic

---

# ISSUE-057 Completion Checklist

ISSUE-057 is complete when all examples in this document can be produced end-to-end against a seeded test repository.

| Command | Required |
| --- | --- |
| `reponerve ask "Who owns authentication?"` | ✅ |
| `reponerve ask "Why are we using Redis?"` | ✅ |
| `reponerve ask "Who created metadata panel?"` | ✅ |
| `reponerve explain "metadata panel"` (Feature Understanding) | ✅ |
| `reponerve explain-file "<path>"` | ✅ |
| `reponerve explain-function "<symbol>"` | ✅ |
| `reponerve explain-struct "<symbol>"` | ✅ |
| `reponerve explain-interface "<symbol>"` | ✅ |
| `reponerve explain-type "<symbol>"` | ✅ |
| `reponerve plan "Add OAuth login"` | ✅ |
| `reponerve impact "user-service"` | ✅ |
| `reponerve review "metadata panel"` | ✅ |

When this checklist passes, RepoNerve v1.0 release review may proceed. There is no v1.x product release.
