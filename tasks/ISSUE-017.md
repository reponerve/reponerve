# ISSUE-017 — Memory Read Stores

## Objective

Create read-only access to repository memories.

---

## Deliverables

Create:

* EventReader
* DecisionReader
* IntentReader
* FactReader
* RelationshipReader

---

## Queries

Support:

GetByID()

ListByRepository()

ListAll()

---

## Constraints

Read-only.

No write methods.

No mutation operations.

---

## Acceptance Criteria

* Readers implemented
* SQLite implementations created
* Unit tests added
* Integration tests added
