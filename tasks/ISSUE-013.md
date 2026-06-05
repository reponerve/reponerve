# ISSUE-013 — Intent Extraction

## Objective

Extract repository intent from ADRs and commit messages.

---

## Motivation

Intent explains why decisions were made.

---

## Inputs

- adr
- commit

---

## Outputs

- Intent

---

## Intent Schema

```go
type Intent struct {
    ID string

    RepositoryID string

    Description string

    SourceID string
}
```

---

## Extraction Rules

Detect phrases containing:

- improve
- reduce
- increase
- simplify
- optimize
- enhance

Examples:

"Reduce database latency"

"Improve deployment reliability"

---

## Acceptance Criteria

- Intents extracted deterministically
- No AI required
- Source references preserved
- Tests added