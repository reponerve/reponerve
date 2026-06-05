# ISSUE-014 — Fact Extraction

## Objective

Extract repository facts from ADRs and documentation.

---

## Motivation

Facts describe repository truths.

---

## Inputs

- adr
- documentation (future)

---

## Outputs

- Fact

---

## Fact Schema

```go
type Fact struct {
    ID string

    RepositoryID string

    Subject string

    Predicate string

    Object string

    SourceID string
}
```

---

## Example

Input:

"Authentication service uses Redis"

Output:

Subject: Authentication Service
Predicate: Uses
Object: Redis

---

## Acceptance Criteria

- Facts extracted using deterministic rules
- Source references preserved
- Tests added