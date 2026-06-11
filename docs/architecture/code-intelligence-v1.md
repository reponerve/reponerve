# Code Intelligence V1

Status: Draft

Version: v1.0

Issue: ISSUE-057

See also: `docs/architecture/code-intelligence.md`

---

# Overview

Code Intelligence is the authoritative source for code understanding in RepoNerve.

Entity hierarchy:

```text
Module → Package → File → Symbols
```

Code Intelligence complements Repository Intelligence.

Repository Intelligence explains why.

Code Intelligence explains how.

---

# Entity Types

| EntityType | Description |
| --- | --- |
| `module` | Go module from go.mod / go.work |
| `package` | Go package |
| `file` | Source file |
| `struct` | Struct declaration |
| `interface` | Interface declaration |
| `type_alias` | Type alias declaration |
| `function` | Package-level function |
| `method` | Type-associated function |
| `endpoint` | Exposed endpoint (http, grpc, graphql, cli) |

Generic `type` and `api` entities are not used.

---

# Relationships

| RelationshipType | Description |
| --- | --- |
| `MODULE_CONTAINS_PACKAGE` | Module contains package |
| `BELONGS_TO_MODULE` | Package belongs to module |
| `BELONGS_TO_PACKAGE` | Entity belongs to package |
| `DEFINED_IN_FILE` | Symbol defined in file |
| `IMPORTS` | File imports package |
| `CALLS` | Invocation |
| `IMPLEMENTS` | Struct satisfies interface |
| `DEPENDS_ON` | Package dependency |
| `REFERENCES` | Symbol reference |
| `EXPOSES_ENDPOINT` | Handler exposes endpoint |

`DEFINED_IN` is removed.

---

# Repository ↔ Code Linkage

Cross-authority links stored in `repository_code_relationships`:

* `DECISION_REFERENCES_CODE`
* `FACT_REFERENCES_CODE`
* `EVENT_REFERENCES_CODE`
* `CONTEXT_REFERENCES_CODE`

Links must be deterministic. No LLM inference.

---

# Code Explanation Context

```go
type CodeExplanationContext struct {
    Subject      string
    Modules      []*CodeEntity
    Files        []*CodeEntity
    Packages     []*CodeEntity
    Structs      []*CodeEntity
    Interfaces   []*CodeEntity
    TypeAliases  []*CodeEntity
    Functions    []*CodeEntity
    Methods      []*CodeEntity
    Endpoints    []*CodeEntity
    CallGraph    *CallGraph
    Dependencies []*CodeRelationship
    Evidence     []EvidenceItem
}
```

---

# Explain Output Contract

Development Experience combines:

**Code Context** — modules, files, packages, structs, interfaces, type aliases, functions, methods, endpoints, call graph

**Repository Context** — decisions, facts, events, ownership, reviewers, impact

**Repository-Code Links** — deterministic cross-authority connections

No Purpose or History fields in Development Experience output.

---

# Package Structure

```text
internal/code/
    models.go
    parser.go
    indexer.go
    linker.go
    graph.go
    service.go
```

---

# Release Impact

Required for v1.0.0 via ISSUE-057.
