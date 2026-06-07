# ISSUE-037 — Contributor Model

## Objective

Implement the foundational ownership data model.

This issue introduces the first ownership entities into RepoNerve.

---

# Background

Ownership Intelligence requires contributor and expertise entities before extraction, scoring, querying, or MCP exposure can be implemented.

This issue establishes the ownership memory layer.

---

# Scope

## Models

Create:

* Contributor
* Expertise

under:

pkg/models/

---

## Storage Interfaces

Create:

* ContributorStore
* ExpertiseStore

following existing storage conventions.

Interfaces must remain minimal.

---

## SQLite Stores

Implement:

* SQLiteContributorStore
* SQLiteExpertiseStore

Requirements:

* Idempotent upserts
* Existing SQLite patterns
* No business logic

---

## Database Migration

Create migration(s) for:

* contributors
* expertise

Include rollback support.

Update migration tests.

---

## Testing

Cover:

* Contributor storage
* Expertise storage
* Idempotent upserts
* Migration up
* Migration down

---

# Constraints

Do NOT:

* Parse Git history
* Calculate expertise
* Create ownership scores
* Create MCP tools

Only implement the ownership data model.

---

# Acceptance Criteria

Contributor entities can be stored.

Expertise entities can be stored.

Migrations succeed.

Rollback succeeds.

All tests pass.
