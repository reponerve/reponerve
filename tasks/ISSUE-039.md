# ISSUE-039 — Expertise Detection

## Objective

Detect contributor expertise using repository evidence.

---

# Background

Expertise is derived from contributor activity.

Expertise must remain deterministic and explainable.

---

# Scope

## Knowledge Domains

Introduce domain detection.

Examples:

* Authentication
* Storage
* MCP
* Context Engine

Domains must be derived from repository evidence.

---

## Expertise Scores

Generate expertise scores using:

* Contribution frequency
* Repository activity
* Decision participation
* Fact associations

No AI.

No LLM scoring.

---

## Evidence Tracking

Every expertise score must include supporting evidence.

Example:

Authentication Expertise Score: 0.91

Evidence:

* 87 commits
* 4 ADR contributions
* Activity within 30 days

---

## Persistence

Persist expertise records using the ownership storage layer.

---

## Testing

Cover:

* Domain detection
* Expertise scoring
* Evidence generation
* Deterministic results

Include integration tests.

---

# Explainability Requirement

Every expertise score must be reproducible and traceable to repository evidence.

---

# Acceptance Criteria

Expertise records are generated deterministically.

Supporting evidence is available.

All tests pass.
