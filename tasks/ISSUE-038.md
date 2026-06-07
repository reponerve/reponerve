# ISSUE-038 — Contributor Extraction

## Objective

Implement deterministic contributor extraction from repository evidence.

---

# Background

Ownership Intelligence begins with contributors.

Contributor records must be derived from repository artifacts.

Initial source:

* Git commit history

---

# Scope

## Extraction

Create:

internal/ownership/extraction/

Extract contributors from Git commit sources already stored by RepoNerve.

---

## Aggregation

Deduplicate contributors using:

* Author Name
* Author Email

---

## Statistics

Compute:

* FirstSeen
* LastSeen
* CommitCount

using commit history.

---

## Deterministic IDs

Contributor IDs must be generated deterministically.

Suggested inputs:

* RepositoryID
* Email

---

## Persistence

Integrate extraction into ingestion orchestration.

Flow:

Git Sources
↓
Contributor Extraction
↓
Contributor Store

---

## Testing

Cover:

* Single contributor
* Multiple contributors
* Deduplication
* Statistics calculation
* Deterministic IDs

Include integration tests.

---

# Explainability Requirement

Every contributor record must be traceable to repository evidence.

No inferred ownership.

No expertise calculations.

---

# Acceptance Criteria

Contributor records are extracted deterministically.

Contributor statistics are persisted.

All tests pass.

---

## Status

Completed. All tasks implemented and tested successfully.

