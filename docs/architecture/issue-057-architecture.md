# ISSUE-057 Architecture

Version: v1.0

Status: Draft — Pending Approval

Issue: ISSUE-057 — Code Intelligence & Development Experience

Acceptance Criteria: `docs/examples/development-experience.md`

Related:

* `docs/architecture/code-storage-model.md`
* `docs/architecture/development-experience-contracts.md`
* `docs/architecture/code-intelligence.md`
* `docs/product/implementation-status.md`
* `docs/product/token-economics.md`
* `docs/roadmap/v1.0-iteration-plan.md`

---

# Purpose

This document defines the architecture for ISSUE-057.

ISSUE-057 completes RepoNerve's path to **Software Understanding** by delivering Code Intelligence, Repository-Code Linking, Feature Understanding, and Development Experience.

Authority boundaries in this document are unchanged. See `docs/vision/vision.md` for product mission.

---

# Architectural Position

```text
Repository Source
    ↓
Ingestion
    ├── Git / ADR / Documentation Scan  (existing)
    ├── Code Indexing                   (new)
    └── Repository-Code Linking           (new)
    ↓
Storage
    ├── Repository Memory Store         (existing)
    ├── Code Store                      (new)
    └── Repository-Code Link Store      (new)
    ↓
Intelligence
    ├── Repository Intelligence         (existing — authoritative)
    └── Code Intelligence               (new — authoritative)
    ↓
Development Experience                (new — orchestrator)
    ↓
CLI / MCP
```

Authority rules:

| Layer | Authority |
| --- | --- |
| Code Intelligence | Code structure, symbols, call graphs |
| Repository Intelligence | Decisions, facts, events, ownership, graph |
| Repository-Code Links | Deterministic cross-authority references |
| Development Experience | Orchestration only — no new intelligence |

---

# Entity Hierarchy

```text
Module
    ↓
Package
    ↓
File
    ↓
Symbols
    ├── struct
    ├── interface
    ├── type_alias
    ├── function
    ├── method
    └── endpoint
```

Single-module repositories contain one module entity derived from `go.mod`.

Monorepos and `go.work` workspaces may contain multiple module entities.

---

# Code Entities

Code Intelligence indexes deterministic code entities.

## Entity Types

| EntityType | Description | Example |
| --- | --- | --- |
| `module` | Go module boundary | `github.com/reponerve/reponerve` |
| `package` | Go package | `internal/agent/search` |
| `file` | Source file | `internal/agent/search/service.go` |
| `struct` | Struct declaration | `Service` |
| `interface` | Interface declaration | `Searcher` |
| `type_alias` | Type alias declaration | `HandlerFunc` |
| `function` | Package-level function | `NewService` |
| `method` | Type-associated function | `(s *Service) Search` |
| `endpoint` | Exposed endpoint surface | `GET /api/v1/search` |

### Endpoint Types

Endpoint entities include an `EndpointType` discriminator:

| EndpointType | Description |
| --- | --- |
| `http` | HTTP handler or route |
| `grpc` | gRPC service method |
| `graphql` | GraphQL resolver or operation |
| `cli` | CLI command or subcommand |

v1 initial scope: `http` and `cli` detection where parser evidence exists.

---

## Entity Models

### Module

```go
type Module struct {
    ID           string
    RepositoryID string
    Name         string
    ModulePath   string // e.g. github.com/reponerve/reponerve
    Language     string
    EvidenceJSON string
    IndexedAt    time.Time
}
```

### CodeEntity

All non-module symbols use `CodeEntity`:

```go
type CodeEntity struct {
    ID            string
    RepositoryID  string
    EntityType    string // struct, interface, type_alias, function, method, endpoint, file, package
    Name          string
    QualifiedName string
    FilePath      string
    PackagePath   string
    ModulePath    string
    Language      string
    StartLine     int
    EndLine       int
    Signature     string  // optional
    EndpointType  string  // endpoint entities only: http, grpc, graphql, cli
    EvidenceJSON  string
    IndexedAt     time.Time
}
```

## Entity ID Rule

Entity IDs must be deterministic:

```text
ID = sha256(repositoryID + ":" + entityType + ":" + qualifiedName)
```

Re-indexing the same repository state must produce identical entity IDs.

## EvidenceJSON — Entity

```json
{
  "source": "go/ast",
  "file": "internal/agent/search/service.go",
  "start_line": 120,
  "end_line": 180,
  "parser": "go/parser"
}
```

## v1.0 Explain Commands

Explicit symbol entities enable v1.0 explain commands:

* `explain-struct`
* `explain-interface`
* `explain-type` (resolves `type_alias`)

v1.0 also ships `explain`, `explain-file`, and `explain-function`. All explain commands are required for v1.0 release.

---

# Code Relationships

Code Intelligence builds deterministic relationships between code entities.

## Relationship Types

| RelationshipType | From → To | Description |
| --- | --- | --- |
| `MODULE_CONTAINS_PACKAGE` | module → package | Module contains package |
| `BELONGS_TO_MODULE` | package → module | Package belongs to module |
| `BELONGS_TO_PACKAGE` | file, symbol → package | Entity belongs to package |
| `DEFINED_IN_FILE` | symbol → file | Symbol defined in file |
| `IMPORTS` | file → package | File imports package |
| `CALLS` | function/method → function/method | Invocation |
| `IMPLEMENTS` | struct → interface | Struct satisfies interface |
| `DEPENDS_ON` | package → package | Package dependency |
| `REFERENCES` | symbol → symbol | Non-call reference |
| `EXPOSES_ENDPOINT` | function/method → endpoint | Handler exposes endpoint |

`DEFINED_IN` is removed. Use `DEFINED_IN_FILE`, `BELONGS_TO_PACKAGE`, and `BELONGS_TO_MODULE` instead.

## Relationship Model

```go
type CodeRelationship struct {
    ID               string
    RepositoryID     string
    FromEntityID     string
    ToEntityID       string
    RelationshipType string
    EvidenceJSON     string
    IndexedAt        time.Time
}
```

## EvidenceJSON — Relationship

```json
{
  "source": "go/ast",
  "file": "internal/agent/search/service.go",
  "line": 145,
  "relationship": "CALLS",
  "from_symbol": "Search",
  "to_symbol": "collectMemoryHits"
}
```

## Graph Views

| Graph | Root | Traversal |
| --- | --- | --- |
| Module Graph | module entity | MODULE_CONTAINS_PACKAGE, BELONGS_TO_MODULE |
| File Graph | file entity | IMPORTS, REFERENCES, DEFINED_IN_FILE |
| Package Graph | package entity | DEPENDS_ON, IMPORTS, BELONGS_TO_MODULE |
| Call Graph | function/method entity | CALLS (bidirectional for callers) |

Graph traversal must be deterministic:

1. Depth ASC
2. RelationshipType ASC
3. ToEntityID ASC

---

# Repository ↔ Code Linkage

Repository Intelligence and Code Intelligence remain independent authorities.

Cross-authority references are stored in a dedicated relationship family.

## Relationship Types

| RelationshipType | From → To | Description |
| --- | --- | --- |
| `DECISION_REFERENCES_CODE` | decision → code entity | ADR or decision cites code |
| `FACT_REFERENCES_CODE` | fact → code entity | Fact references code artifact |
| `EVENT_REFERENCES_CODE` | event → code entity | Event references code change |
| `CONTEXT_REFERENCES_CODE` | context artifact → code entity | Context package references code |

## Relationship Model

```go
type RepositoryCodeRelationship struct {
    ID                   string
    RepositoryID         string
    RepositoryEntityID   string
    RepositoryEntityType string // DECISION, FACT, EVENT, CONTEXT
    CodeEntityID         string
    CodeEntityType       string // file, package, struct, interface, function, method, endpoint, ...
    RelationshipType     string
    EvidenceJSON         string
    IndexedAt            time.Time
}
```

## Linking Rules

Links must be deterministic.

### Allowed Evidence Sources

* ADR file path references
* ADR symbol or package mentions with exact match
* Source links in commit messages or PR metadata
* Explicit file path references in repository memory text
* Structured annotations in existing repository relationships
* Existing repository relationship targets that resolve to code entities

### Disallowed

* LLM inference
* Semantic guessing
* Embedding similarity
* Heuristic fuzzy matching without source evidence

## Traversal Behavior

Explain and related commands traverse links bidirectionally:

```text
Repository Entity
        ↓
Repository-Code Link
        ↓
Code Entity
```

Reverse traversal is also supported:

```text
Code Entity
        ↓
Repository-Code Link
        ↓
Repository Entity
```

Traversal ordering:

1. RelationshipType ASC
2. RepositoryEntityType ASC, RepositoryEntityID ASC
3. CodeEntityType ASC, CodeEntityID ASC

---

# Storage Model

Code entities, code relationships, and repository-code relationships persist in SQLite via store interfaces.

Consumers must not access SQLite directly.

## Tables

### code_entities

```sql
CREATE TABLE IF NOT EXISTS code_entities (
    id              TEXT PRIMARY KEY,
    repository_id   TEXT NOT NULL,
    entity_type     TEXT NOT NULL,
    name            TEXT NOT NULL,
    qualified_name  TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    package_path    TEXT NOT NULL,
    module_path     TEXT NOT NULL,
    language        TEXT NOT NULL,
    start_line      INTEGER NOT NULL,
    end_line        INTEGER NOT NULL,
    signature       TEXT,
    endpoint_type   TEXT,
    evidence_json   TEXT NOT NULL,
    indexed_at      TEXT NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
);

CREATE INDEX idx_code_entities_repo ON code_entities(repository_id);
CREATE INDEX idx_code_entities_type ON code_entities(repository_id, entity_type);
CREATE INDEX idx_code_entities_module ON code_entities(repository_id, module_path);
CREATE INDEX idx_code_entities_file ON code_entities(repository_id, file_path);
CREATE INDEX idx_code_entities_name ON code_entities(repository_id, name);
CREATE INDEX idx_code_entities_qualified ON code_entities(repository_id, qualified_name);
```

Supported `entity_type` values:

`module`, `package`, `file`, `struct`, `interface`, `type_alias`, `function`, `method`, `endpoint`

### code_relationships

```sql
CREATE TABLE IF NOT EXISTS code_relationships (
    id                TEXT PRIMARY KEY,
    repository_id     TEXT NOT NULL,
    from_entity_id    TEXT NOT NULL,
    to_entity_id      TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    evidence_json     TEXT NOT NULL,
    indexed_at        TEXT NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id),
    FOREIGN KEY (from_entity_id) REFERENCES code_entities(id),
    FOREIGN KEY (to_entity_id) REFERENCES code_entities(id)
);

CREATE INDEX idx_code_rels_repo ON code_relationships(repository_id);
CREATE INDEX idx_code_rels_from ON code_relationships(from_entity_id);
CREATE INDEX idx_code_rels_to ON code_relationships(to_entity_id);
CREATE INDEX idx_code_rels_type ON code_relationships(repository_id, relationship_type);
```

### repository_code_relationships

```sql
CREATE TABLE IF NOT EXISTS repository_code_relationships (
    id                     TEXT PRIMARY KEY,
    repository_id          TEXT NOT NULL,
    repository_entity_id   TEXT NOT NULL,
    repository_entity_type TEXT NOT NULL,
    code_entity_id         TEXT NOT NULL,
    code_entity_type       TEXT NOT NULL,
    relationship_type      TEXT NOT NULL,
    evidence_json          TEXT NOT NULL,
    indexed_at             TEXT NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id),
    FOREIGN KEY (code_entity_id) REFERENCES code_entities(id)
);

CREATE INDEX idx_repo_code_rels_repo ON repository_code_relationships(repository_id);
CREATE INDEX idx_repo_code_rels_repo_entity ON repository_code_relationships(repository_id, repository_entity_id);
CREATE INDEX idx_repo_code_rels_code_entity ON repository_code_relationships(code_entity_id);
CREATE INDEX idx_repo_code_rels_type ON repository_code_relationships(repository_id, relationship_type);
```

### code_index_state

```sql
CREATE TABLE IF NOT EXISTS code_index_state (
    repository_id      TEXT PRIMARY KEY,
    last_indexed_at    TEXT NOT NULL,
    module_count       INTEGER NOT NULL,
    file_count         INTEGER NOT NULL,
    entity_count       INTEGER NOT NULL,
    relationship_count INTEGER NOT NULL,
    link_count         INTEGER NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
);
```

## Store Interfaces

```text
internal/storage/store.go
    CodeEntityStore
    CodeRelationshipStore
    RepositoryCodeRelationshipStore
    CodeIndexStateStore

internal/storage/sqlite/
    code_entity_store.go
    code_relationship_store.go
    repository_code_relationship_store.go
    code_index_state_store.go

internal/query/storage/
    code_reader.go
        CodeEntityReader
        CodeRelationshipReader
        RepositoryCodeRelationshipReader
```

## Query Interfaces

```go
type CodeEntityReader interface {
    GetByID(ctx context.Context, id string) (*CodeEntity, error)
    ListByRepository(ctx context.Context, repositoryID string) ([]*CodeEntity, error)
    ListByFilePath(ctx context.Context, repositoryID, filePath string) ([]*CodeEntity, error)
    ListByModulePath(ctx context.Context, repositoryID, modulePath string) ([]*CodeEntity, error)
    ListByEntityType(ctx context.Context, repositoryID, entityType string) ([]*CodeEntity, error)
    FindByQualifiedName(ctx context.Context, repositoryID, qualifiedName string) ([]*CodeEntity, error)
}

type CodeRelationshipReader interface {
    ListByFromEntity(ctx context.Context, entityID string) ([]*CodeRelationship, error)
    ListByToEntity(ctx context.Context, entityID string) ([]*CodeRelationship, error)
    ListByRepository(ctx context.Context, repositoryID string) ([]*CodeRelationship, error)
}

type RepositoryCodeRelationshipReader interface {
    ListByRepositoryEntity(ctx context.Context, repositoryID, repositoryEntityID string) ([]*RepositoryCodeRelationship, error)
    ListByCodeEntity(ctx context.Context, repositoryID, codeEntityID string) ([]*RepositoryCodeRelationship, error)
    ListByRepository(ctx context.Context, repositoryID string) ([]*RepositoryCodeRelationship, error)
}
```

## Read Ordering

All store reads must return results sorted:

1. EntityType ASC
2. ModulePath ASC
3. FilePath ASC
4. StartLine ASC
5. QualifiedName ASC
6. ID ASC

---

# Ingestion Model

Code indexing and repository-code linking extend the existing scan pipeline.

## Extended Flow

```text
reponerve scan
    ↓
Coordinator.Run()
    ↓
Pipeline.Execute()  → Git, ADR scanners
    ↓
Extraction + Linking  → Memory entities
    ↓
CodeIndexer.Index()  → Code entities + code relationships
    ↓
RepositoryCodeLinker.Link()  → repository_code_relationships
    ↓
scan_state + code_index_state updated
```

## CodeIndexer Responsibilities

```text
internal/code/indexer.go
```

1. Parse `go.mod` and `go.work` to discover module entities
2. Discover changed Go files since last code index
3. Parse each file with `go/parser` and `go/ast`
4. Extract entities: module, package, file, struct, interface, type_alias, function, method, endpoint
5. Extract relationships using explicit relationship types
6. Upsert entities and relationships (idempotent)
7. Remove stale entities for deleted/changed files
8. Update `code_index_state`

## RepositoryCodeLinker Responsibilities

```text
internal/code/linker.go
```

1. Scan ADR and repository memory text for explicit file path references
2. Scan structured repository relationships for code-resolvable targets
3. Resolve references to indexed code entities
4. Create `repository_code_relationships` with evidence
5. Skip unresolved references — do not guess

## Incremental Indexing

* First scan: full index + full link pass
* Subsequent scans: re-index changed files; re-link affected repository entities
* Full re-index: deterministic — same source produces same entities and links

## Ignore Rules

Skip:

* `vendor/`
* Generated files (detected by `// Code generated` header)
* Non-Go files (v1 scope)
* Paths in `.gitignore`

---

# Development Experience Package

```text
internal/agent/development/
    models.go
    router.go
    resolver.go
    service.go
```

## Service Dependencies

```go
type Service struct {
    codeService                    *code.Service
    searchService                  *agentsearch.Service
    qaService                      *qa.Service
    guidanceService                *guidance.Service
    impactService                  *agentimpact.Service
    onboardingService              *onboarding.Service
    contextService                 *agentcontext.Service
    discoveryService               *discovery.Service
    learningService                *learning.Service
    reviewerService                *reviewers.Service
    changePlanService              *changeplan.Service
    workflowService                *workflow.Service
    graphImpactService             *graphimpact.Service
    repositoryCodeRelationshipReader storage.RepositoryCodeRelationshipReader
}
```

Development Experience must not generate free-form summaries.

Purpose and history are derived only from Decisions, Facts, and Events in `RepositoryContext`.

---

# Topic Resolution

```text
Input Text
    ↓
Normalize
    ↓
Parallel Resolution
    ├── Repository Search → repository entity hits
    └── Code Intelligence → code entity hits
    ↓
Repository-Code Link Traversal → connected cross-authority entities
    ↓
ResolvedTopic
```

## ResolvedTopic Model

```go
type ResolvedTopic struct {
    Input                  string
    RepositoryHits         []*agentsearch.SearchHit
    CodeEntities           []*code.CodeEntity
    RepositoryCodeLinks    []*RepositoryCodeRelationship
    PrimaryEntityType      string // "code" | "repository" | "mixed"
    MatchEvidence          string
}
```

---

# Explain Contract

## Commands

| Command | Input | Primary Resolution |
| --- | --- | --- |
| `explain` | natural language topic | Search + Code Intelligence + Repository-Code Links |
| `explain-file` | file path | Code Intelligence.ResolveFile |
| `explain-function` | symbol name | Code Intelligence.ResolveSymbol |
| `explain-struct` | struct name | Code Intelligence.ResolveSymbol |
| `explain-interface` | interface name | Code Intelligence.ResolveSymbol |
| `explain-type` | type alias name | Code Intelligence.ResolveSymbol |

All explain commands are required for v1.0 release.

## API

```go
func (s *Service) Explain(ctx context.Context, req DevelopmentRequest) (*DevelopmentExplanation, error)
func (s *Service) ExplainFile(ctx context.Context, repositoryID, filePath string) (*DevelopmentExplanation, error)
func (s *Service) ExplainFunction(ctx context.Context, repositoryID, symbol string) (*DevelopmentExplanation, error)
func (s *Service) ExplainStruct(ctx context.Context, repositoryID, symbol string) (*DevelopmentExplanation, error)
func (s *Service) ExplainInterface(ctx context.Context, repositoryID, symbol string) (*DevelopmentExplanation, error)
func (s *Service) ExplainType(ctx context.Context, repositoryID, symbol string) (*DevelopmentExplanation, error)
```

## Output Model

```go
type DevelopmentExplanation struct {
    Topic                  string
    CodeContext            *CodeContext
    RepositoryContext      *RepositoryContext
    RepositoryCodeLinks    []RepositoryCodeLinkRef
    Evidence               []EvidenceItem
    SourceServices         []string
}

type CodeContext struct {
    Modules      []EntityRef
    Files        []EntityRef
    Packages     []EntityRef
    Structs      []EntityRef
    Interfaces   []EntityRef
    TypeAliases  []EntityRef
    Functions    []EntityRef
    Methods      []EntityRef
    Endpoints    []EntityRef
    CallGraph    *code.CallGraph
    Dependencies []RelationshipRef
}

type RepositoryContext struct {
    Decisions   []EntityRef
    Facts       []EntityRef
    Events      []EntityRef
    Owners      []EntityRef
    Expertise   []EntityRef
    Reviewers   []EntityRef
    Impact      []EntityRef
    ChangePlans []EntityRef
}

type RepositoryCodeLinkRef struct {
    RelationshipType     string
    RepositoryEntityRef  EntityRef
    CodeEntityRef        EntityRef
    EvidenceJSON         string
}
```

`Purpose` and `History` fields are intentionally absent.

Development Experience must not generate free-form narrative summaries.

Rationale, origin, and evolution appear only as structured Decisions, Facts, and Events in `RepositoryContext`.

## Orchestration Flow

```text
Explain Request
    ↓
Resolve Topic
    ↓
Code Intelligence
    ├── ResolveFile / ResolveSymbol / search-matched entities
    ├── BuildCallGraph
    └── AnalyzeSymbolDependencies
    ↓
Repository Intelligence
    ├── Repository Search
    ├── Ownership Intelligence
    ├── Architectural Guidance
    └── Context Engine
    ↓
Repository-Code Link Traversal
    ├── repository entity → linked code entities
    └── code entity → linked repository entities
    ↓
Assemble DevelopmentExplanation
    ↓
Validate evidence on all sections
```

## Unified Explanation Example

```text
Decision: ADR-004 OAuth Authentication

Repository-Code Links:
  DECISION_REFERENCES_CODE → oauth.go
  DECISION_REFERENCES_CODE → AuthService (struct)
  DECISION_REFERENCES_CODE → LoginHandler (function)

CodeContext:
  Files: oauth.go
  Structs: AuthService
  Functions: LoginHandler

RepositoryContext:
  Decisions: ADR-004 OAuth Authentication
  Facts: Auth Service DEPENDS_ON SQLite
  Events: Introduce OAuth login flow
```

RepositoryContext and CodeContext are connected through `RepositoryCodeLinks`.

## Required Sections

Every explain output must include:

* `CodeContext` — at minimum Files when code entities resolved
* `RepositoryContext` — Decisions, Facts, or Events when repository matches exist
* `RepositoryCodeLinks` — when cross-authority links exist
* `Evidence` — non-empty when any section is non-empty
* `SourceServices` — lists all upstream authorities consulted

---

# Ask Contract

## API

```go
func (s *Service) Ask(ctx context.Context, req DevelopmentRequest) (*DevelopmentAnswer, error)
```

## Question Classification

| Pattern | AnswerType | Authority |
| --- | --- | --- |
| `who owns ...` | `ownership` | Ownership + Search |
| `who created ...` | `authorship` | Ownership + Search |
| `who worked on ...` | `authorship` | Ownership + Search |
| `why ...` / `why are we using ...` | `decision_rationale` | Guidance + Q&A + Search |
| `what depends on ...` | `dependency` | Graph + Code + Search + Repository-Code Links |
| `what is this repository` | `overview` | Onboarding |
| fallback | `search_summary` | Search + best-effort routing |

## Output Model

```go
type DevelopmentAnswer struct {
    Question       string
    AnswerType     string
    Summary        string // structured from upstream authorities only
    Related        []EntityRef
    Evidence       []EvidenceItem
    SourceServices []string
}
```

Summary must be assembled from upstream authority output — not free-form generation.

---

# Plan Contract

## API

```go
func (s *Service) Plan(ctx context.Context, req DevelopmentRequest) (*DevelopmentPlan, error)
```

## Output Model

```go
type DevelopmentPlan struct {
    Task                  string
    ImpactedAreas         []EntityRef
    RelevantDecisions     []EntityRef
    RelevantFacts         []EntityRef
    Owners                []EntityRef
    Reviewers             []EntityRef
    SuggestedWorkflow     string
    StartingPoints        []EntityRef
    RepositoryCodeLinks   []RepositoryCodeLinkRef
    Evidence              []EvidenceItem
    SourceServices        []string
}
```

Starting points may reference code entities (files, structs, functions) and repository entities (decisions, ADRs) connected via repository-code links.

---

# Impact Contract

## API

```go
func (s *Service) AnalyzeImpact(ctx context.Context, req DevelopmentRequest) (*DevelopmentImpactReport, error)
```

## Output Model

```go
type DevelopmentImpactReport struct {
    Subject               string
    ImpactedDecisions     []EntityRef
    ImpactedFacts         []EntityRef
    ImpactedEvents        []EntityRef
    CodeDependencies      []RelationshipRef
    DependentAreas        []EntityRef
    Owners                []EntityRef
    RepositoryCodeLinks   []RepositoryCodeLinkRef
    Evidence              []EvidenceItem
    SourceServices        []string
}
```

Impact traverses repository-code links to connect repository impact chains to code dependencies.

---

# Review Contract

## API

```go
func (s *Service) PrepareReview(ctx context.Context, req DevelopmentRequest) (*DevelopmentReviewGuide, error)
```

## Output Model

```go
type DevelopmentReviewGuide struct {
    Topic                 string
    RecommendedReviewers  []EntityRef
    RequiredExpertise     []EntityRef
    AffectedAreas         []EntityRef
    RelatedKnowledge      []EntityRef
    SuggestedWorkflow     string
    RepositoryCodeLinks   []RepositoryCodeLinkRef
    Evidence              []EvidenceItem
    SourceServices        []string
}
```

---

# Shared Types

```go
type EntityRef struct {
    EntityType string
    EntityID   string
    Label      string
}

type RelationshipRef struct {
    RelationshipType string
    FromEntityID     string
    ToEntityID       string
    Label            string
    EvidenceJSON     string
}

type EvidenceItem struct {
    Source  string
    Type    string
    Payload json.RawMessage
}
```

---

# CLI Contract

```text
internal/cli/ask/               → development.Service.Ask
internal/cli/explain/           → development.Service.Explain
internal/cli/explainfile/       → development.Service.ExplainFile
internal/cli/explainfunction/   → development.Service.ExplainFunction
internal/cli/explainstruct/     → development.Service.ExplainStruct
internal/cli/explaininterface/  → development.Service.ExplainInterface
internal/cli/explaintype/       → development.Service.ExplainType
internal/cli/plan/              → development.Service.Plan
internal/cli/impact/            → development.Service.AnalyzeImpact
internal/cli/review/            → development.Service.PrepareReview
```

All CLI commands above are required for v1.0 release.

---

# Determinism Requirements

Same repository state + same input = same output.

| Output | Sort Order |
| --- | --- |
| Entity lists | EntityType ASC, Label ASC, EntityID ASC |
| Repository-Code links | RelationshipType ASC, RepositoryEntityID ASC, CodeEntityID ASC |
| Evidence | Source ASC, Type ASC |
| SourceServices | alphabetical ASC |
| Call graph traversal | Depth ASC, RelationshipType ASC, ToEntityID ASC |

---

# Testing Strategy

## Unit Tests

* Module, package, file, struct, interface, type_alias, function, method, endpoint extraction
* Relationship extraction with explicit relationship types
* Repository-code link extraction from ADR references
* Entity ID determinism
* Topic resolution and routing
* Each contract with mocked upstream services
* Evidence preservation
* No Purpose/History field generation

## Integration Tests

* Migration-backed SQLite including `repository_code_relationships`
* Scan → code index → repository-code link → explain end-to-end
* All examples in `docs/examples/development-experience.md`

---

# Implementation Order

1. Code entities + storage migration (including `repository_code_relationships`)
2. Code indexer + module discovery (`go.mod`, `go.work`)
3. Repository-code linker
4. Code Intelligence service
5. Development Experience resolver + router
6. **Explain contract** — first developer-facing value
7. Ask, Plan, Impact, Review contracts
8. CLI commands
9. Integration tests against acceptance examples

---

# Architecture Approval Criteria

Architecture is approved only when:

* Repository Intelligence and Code Intelligence remain independent authorities
* Development Experience remains orchestration only
* Repository-Code relationships are defined with storage, indexing, query interfaces, and traversal behavior
* Explicit symbol entities exist (`struct`, `interface`, `type_alias`)
* Module entity and hierarchy exist
* Endpoint entity replaces generic `api` entity
* Purpose and History fields are absent from Development Experience output models
* Relationship naming uses `DEFINED_IN_FILE`, `BELONGS_TO_PACKAGE`, `BELONGS_TO_MODULE`

---

# Approval

| Reviewer | Status | Date |
| --- | --- | --- |
| Architecture | Pending | |
| Product | Pending | |

Implementation begins after approval.
