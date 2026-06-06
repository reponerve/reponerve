# ISSUE-019 — Memory Lookup Commands

## Objective

Retrieve individual memories by ID.

---

# Commands

## Event

```bash
reponerve memory get event <id>
```

---

## Decision

```bash
reponerve memory get decision <id>
```

---

## Intent

```bash
reponerve memory get intent <id>
```

---

## Fact

```bash
reponerve memory get fact <id>
```

---

# Output

Human-readable detail view.

Example:

```text
Decision

ID:
decision_xxx

Title:
Use Redis Cache

Status:
Accepted

Source:
adr_xxx
```

---

# Constraints

Read-only.

No graph traversal.

No explanations.

---

# Acceptance Criteria

* All memory types retrievable by ID.
* Not-found handling implemented.
* Unit tests added.
* Integration tests added.

```
```
