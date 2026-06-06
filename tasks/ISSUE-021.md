# ISSUE-021 — Memory Explain Engine

## Objective

Generate human-readable explanations from memory graphs.

---

# Motivation

Users should not need to understand graph structures.

---

# Commands

## Explain Decision

```bash
reponerve memory explain decision <id>
```

---

## Explain Event

```bash
reponerve memory explain event <id>
```

---

# Examples

Decision:

```text
Use Redis Cache
```

Produces:

```text
Decision:
Use Redis Cache

Reason:
Reduce Database Latency

Supporting Facts:
- Authentication Service USES Redis

Resulting Events:
- Introduce Redis Cache
```

---

# Explain Templates

Use deterministic templates.

No AI.

No LLMs.

No embeddings.

---

# V1 Templates

## Decision

Include:

* Decision
* Related Intents
* Supporting Facts
* Resulting Events

---

## Event

Include:

* Event
* Parent Decision
* Driving Intent

---

# Constraints

Deterministic only.

No generative AI.

---

# Acceptance Criteria

* Decision explanations generated.
* Event explanations generated.
* Traceability preserved.
* Unit tests added.
* Integration tests added.

```
```
