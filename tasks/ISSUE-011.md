# ISSUE-011 — Event Extraction

## Objective

Extract repository events from source artifacts.

Initial implementation will focus on Git commits.

---

## Motivation

Repository history contains a timeline of important events.

RepoNerve should preserve those events as structured memory.

Examples:

- Redis Introduced
- Authentication Service Created
- API Deprecated

---

## Inputs

### Source Types

- commit

---

## Outputs

### Memory Type

- Event

---

## Event Schema

```go
type Event struct {
    ID string

    RepositoryID string

    EventType string

    Title string

    Description string

    SourceID string

    Timestamp time.Time
}
```

---

## Extraction Rules

### Rule 1

Commit message beginning with:

- feat:
- fix:
- refactor:
- chore:
- docs:

creates an Event.

### Rule 2

Event title derived from commit title.

### Rule 3

Event references original commit source.

---

## Storage

Persist events in memory tables.

---

## Acceptance Criteria

- Events extracted from commits
- Events persisted
- Source traceability maintained
- Unit tests added
- Integration tests added