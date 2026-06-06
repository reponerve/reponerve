# ISSUE-020 — Memory Trace Engine

## Objective

Enable relationship traversal across memories.

---

# Motivation

Knowledge becomes valuable when connected.

---

# Commands

## Trace Decision

```bash
reponerve memory trace decision <id>
```

---

## Trace Event

```bash
reponerve memory trace event <id>
```

---

## Trace Intent

```bash
reponerve memory trace intent <id>
```

---

# Traversal Rules

## Decision

Show:

```text
Decision
    ↑
Intent

Decision
    ↓
Event

Decision
    ↑
Fact
```

---

## Event

Show:

```text
Event
    ↑
Decision
    ↑
Intent
```

---

## Intent

Show:

```text
Intent
    ↓
Decision
    ↓
Event
```

---

# Output

Tree-like structure.

Example:

```text
Intent
└── Decision
    └── Event
```

---

# Constraints

Only relationship traversal.

No natural-language explanations.

---

# Acceptance Criteria

* Traversal implemented.
* Relationship queries supported.
* Tree output supported.
* Unit tests added.
* Integration tests added.

```
```
