# Extraction Rules V1

## Purpose

This document defines the deterministic extraction rules used by the RepoNerve Memory Engine.

The goal of V1 is to transform repository sources into structured memories without requiring AI models.

All extraction must be:

* Deterministic
* Repeatable
* Traceable
* Testable

---

# General Rules

## Rule 1 — Source Traceability

Every extracted memory must reference at least one source.

Example:

Commit
→ Event

ADR
→ Decision

ADR
→ Intent

---

## Rule 2 — Idempotency

Running extraction multiple times on the same source must produce identical memory records.

---

## Rule 3 — Deterministic First

Extraction logic must rely only on source content.

No LLMs.

No embeddings.

No external APIs.

---

# Event Extraction Rules

## Source Types

* commit

---

## Event Generation

Each qualifying commit produces one Event.

---

## Commit Classification

### Feature

Patterns:

```text
feat:
feature:
```

Produces:

```text
EventType = FEATURE_INTRODUCED
```

---

### Fix

Patterns:

```text
fix:
bugfix:
```

Produces:

```text
EventType = DEFECT_RESOLVED
```

---

### Refactor

Patterns:

```text
refactor:
```

Produces:

```text
EventType = CODE_REFACTORED
```

---

### Documentation

Patterns:

```text
docs:
```

Produces:

```text
EventType = DOCUMENTATION_UPDATED
```

---

### Chore

Patterns:

```text
chore:
```

Produces:

```text
EventType = MAINTENANCE_PERFORMED
```

---

## Event Title

Derived from commit title.

Example:

```text
feat(cache): introduce redis cache
```

Produces:

```text
Introduce Redis Cache
```

---

# Decision Extraction Rules

## Source Types

* adr

---

## Decision Generation

Each ADR produces one Decision.

---

## Title Extraction

ADR title becomes Decision title.

Example:

```text
# Use Redis Cache
```

Produces:

```text
Use Redis Cache
```

---

## Status Extraction

Supported statuses:

```text
Accepted
Rejected
Superseded
Proposed
Deprecated
```

Unknown values are preserved as-is.

---

# Intent Extraction Rules

## Source Types

* adr
* commit

---

## Intent Keywords

Intents are generated when any of the following terms appear:

```text
improve
reduce
increase
optimize
enhance
simplify
minimize
accelerate
stabilize
secure
```

Case-insensitive.

---

## Examples

Input:

```text
Reduce database latency.
```

Produces:

```text
Intent:
Reduce Database Latency
```

---

Input:

```text
Improve deployment reliability.
```

Produces:

```text
Intent:
Improve Deployment Reliability
```

---

## Intent Scope

V1 extracts one intent per matching statement.

---

# Fact Extraction Rules

## Source Types

* adr

---

## Supported Patterns

### Uses Relationship

Pattern:

```text
<subject> uses <object>
```

Produces:

```text
Subject
Predicate = USES
Object
```

---

Example:

```text
Authentication Service uses Redis.
```

Produces:

```text
Subject:
Authentication Service

Predicate:
USES

Object:
Redis
```

---

### Depends On Relationship

Pattern:

```text
<subject> depends on <object>
```

Produces:

```text
Predicate = DEPENDS_ON
```

---

### Calls Relationship

Pattern:

```text
<subject> calls <object>
```

Produces:

```text
Predicate = CALLS
```

---

### Stores In Relationship

Pattern:

```text
<subject> stores data in <object>
```

Produces:

```text
Predicate = STORES_IN
```

---

# Ownership Extraction Rules

## Source Types

* adr
* documentation

---

## Ownership Keywords

Patterns:

```text
owned by
owner:
maintainer:
responsible:
```

Case-insensitive.

---

## Examples

Input:

```text
Authentication Service owned by Platform Team
```

Produces:

```text
Owner:
Platform Team

Resource:
Authentication Service
```

---

Input:

```text
Owner: Backend Team
```

Produces ownership memory.

---

# Relationship Generation Rules

## Intent → Decision

When an intent and decision originate from the same ADR:

```text
INTENT_DRIVES_DECISION
```

---

## Decision → Event

When a commit implements an ADR decision:

```text
DECISION_RESULTS_IN_EVENT
```

Initially determined using title similarity.

---

## Fact → Ownership

When ownership references a known fact subject:

```text
FACT_OWNED_BY
```

---

# Confidence

V1 does not use confidence scores.

All extracted memories are considered deterministic.

Future versions may introduce:

* confidence scoring
* ranking
* AI-assisted extraction

These are out of scope for V1.

---

# Version

Version: 1.0

Status: Draft
