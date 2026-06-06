# ISSUE-018 — Memory List Commands

## Objective

Provide CLI commands for listing repository memories.

---

# Motivation

Users need visibility into repository memory without directly querying SQLite.

---

# Commands

## Events

```bash
reponerve memory list events
```

---

## Decisions

```bash
reponerve memory list decisions
```

---

## Intents

```bash
reponerve memory list intents
```

---

## Facts

```bash
reponerve memory list facts
```

---

## Relationships

```bash
reponerve memory list relationships
```

---

# Output Format

Default:

```text
ID
Title
Type
Created At
```

Human-readable table format.

---

# Filtering

Support:

```bash
--repository
```

Example:

```bash
reponerve memory list decisions \
  --repository repo_id
```

---

# Constraints

Read-only.

No memory mutation.

---

# Acceptance Criteria

* All memory types can be listed.
* Repository filtering supported.
* Unit tests added.
* Integration tests added.

```
```
