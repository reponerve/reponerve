# ISSUE-016 — Memory Linking

## Objective

Create relationships between memories.

---

## Motivation

Knowledge becomes useful when memories are connected.

---

## Inputs

- Event
- Decision
- Intent
- Fact
- Ownership

---

## Outputs

- Relationship

---

## Relationship Schema

```go
type Relationship struct {
    ID string

    FromID string

    ToID string

    Type string
}
```

---

## Examples

Intent
→ Decision

Decision
→ Event

Fact
→ Ownership

---

## Acceptance Criteria

- Relationships persisted
- Link traversal supported
- Source traceability preserved
- Tests added