# ISSUE-012 — Decision Extraction

## Objective

Extract repository decisions from ADRs.

---

## Motivation

Architectural decisions are among the highest-value repository memories.

---

## Inputs

### Source Types

- adr

---

## Outputs

### Memory Type

- Decision

---

## Decision Schema

```go
type Decision struct {
    ID string

    RepositoryID string

    Title string

    Status string

    SourceID string
}
```

---

## Extraction Rules

### Rule 1

ADR title becomes Decision title.

### Rule 2

ADR status becomes Decision status.

### Rule 3

Decision references originating ADR.

---

## Acceptance Criteria

- Decisions extracted from ADRs
- Status preserved
- Source traceability maintained
- Tests added