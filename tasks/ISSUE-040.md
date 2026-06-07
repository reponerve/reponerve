# ISSUE-040 — Ownership Query Engine

## Objective

Expose ownership intelligence through query APIs.

---

# Background

Ownership intelligence must be accessible through deterministic query interfaces.

---

# Scope

## Readers

Implement:

* ListContributors
* GetContributor
* ListExpertise
* TraceContributor

---

## Contributor Trace

Trace:

Contributor
↓
Domains
↓
Decisions
↓
Facts
↓
Events

---

## Query Layer

Reuse existing query architecture.

No direct database access from consumers.

---

## Testing

Cover:

* Contributor listing
* Contributor retrieval
* Expertise retrieval
* Contributor tracing

Include integration tests.

---

# Acceptance Criteria

Ownership intelligence can be queried through dedicated readers.

All tests pass.
