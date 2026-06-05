# ISSUE-015 — Ownership Extraction

## Objective

Extract ownership information.

---

## Inputs

- adr
- documentation
- CODEOWNERS (future)

---

## Outputs

- Ownership

---

## Ownership Schema

```go
type Ownership struct {
    ID string

    RepositoryID string

    Owner string

    Resource string

    SourceID string
}
```

---

## Examples

Backend Team owns Auth Service

Platform Team owns Deployment System

---

## Acceptance Criteria

- Ownership records extracted
- Source traceability preserved
- Tests added