# ISSUE-022 — Context Read Layer

## Objective

Create a read layer for assembling repository context.

---

# Deliverables

Create:

* EventContextReader
* DecisionContextReader
* IntentContextReader
* FactContextReader

or equivalent aggregation interfaces.

---

# Responsibilities

Provide read access required by the Context Engine.

Aggregate memory entities through Query Engine readers.

---

# Constraints

Read-only.

No direct SQLite access from Context Builder.

Use existing Reader interfaces from ISSUE-017.

---

# Acceptance Criteria

* Context read interfaces implemented.
* Reader composition implemented.
* Unit tests added.
* Integration tests added.
