# ISSUE-053 — Repository Search

Status: Planned

Milestone: v1.0

---

# Objective

Implement Repository Search.

Repository Search provides deterministic retrieval of repository knowledge for AI agents and engineers.

It answers:

- What repository knowledge matches this query?
- Which decisions discuss a topic?
- Which facts mention a concept?
- Which contributors are associated with a domain?
- Which repository knowledge is relevant to a search request?

---

# Background

RepoNerve already provides:

- Memory
- Context
- Ownership
- Knowledge Graph
- Repository Intelligence
- Agent Context Packages

However, agents still need a retrieval mechanism to locate repository knowledge efficiently.

Repository Search provides that retrieval capability.

---

# Philosophy

Evidence First.

Repository Search retrieves repository knowledge.

Repository Search does not create repository knowledge.

Repository Search must remain deterministic.

---

# Search Authority Rule

Repository Search retrieves repository knowledge.

Repository Search does not create repository knowledge.

Repository Search must not:

- Generate discovery results
- Generate learning paths
- Generate reviewer recommendations
- Generate change plans
- Generate graph relationships
- Generate repository intelligence

Responsibilities:

Repository Knowledge
↓
Provides Searchable Data

Repository Search
↓
Retrieves Knowledge

Search is a retrieval layer.

It is not an intelligence layer.

---

## Retrieval Authority Rule

Repository Search retrieves repository knowledge.

Repository Search does not determine repository importance.

Repository Search does not determine repository impact.

Repository Search does not determine repository ownership.

Responsibilities:

Repository Search
↓
Retrieval Relevance

Discovery
↓
Importance

Impact Analysis
↓
Impact

Ownership Intelligence
↓
Ownership

Search must not replace any existing intelligence authority.

---

# Architecture Requirements

Reuse:

- Memory Readers
- Ownership Readers
- Knowledge Graph Readers
- Discovery Service

Do NOT:

- Access SQLite directly
- Execute Git commands
- Re-scan repositories
- Generate repository intelligence

---

# Scope

Create:

internal/agent/search/

Files:

- models.go
- service.go
- service_test.go

---

# Models

Implement:

```go
type SearchHit struct {
    EntityType string `json:"entity_type"`

    EntityID string `json:"entity_id"`

    Source string `json:"source"`

    MatchScore int `json:"match_score"`

    EvidenceJSON string `json:"evidence_json"`
}
```

Supported Source values:

- memory
- ownership
- graph
- discovery

---

Implement:

```go
type SearchResult struct {
    RepositoryID string `json:"repository_id"`

    Query string `json:"query"`

    Hits []*SearchHit `json:"hits"`
}
```

---

# Validation

Implement:

```go
func ValidateResult(
    result *SearchResult,
) error
```

Validate:

- RepositoryID exists
- Query exists
- Hits are valid
- EntityType exists
- EntityID exists
- Source exists
- MatchScore >= 0

Reject invalid results.

Reject unsupported Source values.

---

# Search Targets

Repository Search must support:

## Decisions

Search:

- Decision IDs
- Decision Titles
- Decision Content

---

## Facts

Search:

- Subject
- Predicate
- Object

---

## Events

Search:

- Event IDs
- Event Descriptions

---

## Contributors

Search:

- Contributor Name
- Contributor Email

---

## Expertise

Search:

- Domain
- Contributor Association

---

# Search Types

## Structured Query Rule

Structured queries must be parsed deterministically.

Supported prefixes:

- type:
- domain:

Unknown prefixes must return validation errors.

Examples:

Valid:

type:decision redis

type:fact authentication

domain:security

Invalid:

owner:alice

severity:high

Structured query support must remain explicit and deterministic.

---

## Exact Search

Examples:

redis

authentication

oidc

---

## Prefix Search

Examples:

auth*

dec*

cache*

---

## Structured Search

Examples:

type:decision redis

type:fact authentication

type:contributor redis

---

## Domain Search

Examples:

domain:security

domain:database

domain:caching

---

# Match Scoring

MatchScore represents retrieval relevance only.

MatchScore is not:

- Repository importance
- Discovery score
- Reviewer relevance
- Change priority
- Impact severity

MatchScore exists solely to rank search results for a given query.

The same repository entity may have:

Discovery Score: 12.5

MatchScore: 50

These values represent different concepts and must remain independent.

Recommended:

100

Exact match

---

75

Prefix match

---

50

Partial match

---

25

Weak match

Scoring must remain deterministic.

---

# Evidence Model

EvidenceJSON must explain why a hit matched.

Examples:

```json
{
  "match_type": "exact",
  "field": "title"
}
```

```json
{
  "match_type": "prefix",
  "field": "domain"
}
```

Evidence must be machine-readable.

---

# Ordering

Sort by:

1. MatchScore DESC
2. EntityType ASC
3. EntityID ASC

Output must be deterministic.

---

# Service

Implement:

```go
type Service struct {
}
```

---

Constructor:

```go
func NewService(...) *Service
```

Inject required readers and services.

---

# APIs

Implement:

```go
func (s *Service) Search(
    ctx context.Context,
    repositoryID string,
    query string,
) (*SearchResult, error)
```

---

# Search Behavior

Search must:

- Collect matches
- Generate evidence
- Generate match scores
- Sort deterministically

Search must not:

- Generate new repository intelligence
- Re-rank Discovery results
- Modify repository knowledge

---

# Unit Tests

Cover:

- Empty repositories
- Exact matches
- Prefix matches
- Structured search
- Domain search
- Validation
- Source validation
- Evidence generation
- Deterministic ordering

---

# Integration Tests

Create migration-backed SQLite tests.

Verify:

Repository Knowledge
↓
Repository Search

Verify:

- Search execution
- Evidence generation
- Source preservation
- Ordering
- Determinism

---

# Constraints

Do NOT:

- Add embeddings
- Add vector databases
- Add semantic ranking
- Add hybrid ranking
- Add AI reasoning

Only implement deterministic repository retrieval.

---

# Acceptance Criteria

Repository Search executes successfully.

Repository knowledge is retrieved correctly.

Evidence is generated.

Result provenance is preserved.

Match scores are deterministic.

Ordering is deterministic.

No repository intelligence is generated.

No repository importance is computed.

All tests pass.