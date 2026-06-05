# RepoNerve RFC Process

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the Request For Comments (RFC) process used by RepoNerve.

The RFC process exists to ensure that significant changes are:

* Discussed openly
* Evaluated carefully
* Documented permanently
* Aligned with project goals

The goal is to improve decision quality while preserving project velocity.

---

# Why RFCs Exist

Large open-source projects often fail because important decisions happen:

* Informally
* Without documentation
* Without community input
* Without understanding long-term consequences

RFCs provide a structured decision-making process.

---

# What Requires an RFC

An RFC is required for significant changes.

Examples include:

---

## Architecture Changes

Examples:

* Storage redesign
* Memory model changes
* Context engine redesign
* Search engine replacement

---

## New Memory Types

Examples:

* Pattern Memory
* Incident Memory
* Technical Debt Memory
* Risk Memory

---

## Public API Changes

Examples:

* New API contracts
* Breaking API changes
* MCP protocol changes

---

## Major Product Changes

Examples:

* New product direction
* New deployment model
* Multi-repository architecture

---

## Governance Changes

Examples:

* Maintainer model changes
* Voting process changes
* Charter modifications

---

# What Does Not Require an RFC

Small improvements do not require RFCs.

Examples:

* Bug fixes
* Documentation updates
* Refactoring
* Test improvements
* Internal implementation details
* Minor UX improvements

---

# RFC Lifecycle

```text id="3g9w7s"
Idea
  │
  ▼
Draft RFC
  │
  ▼
Discussion
  │
  ▼
Revision
  │
  ▼
Decision
  │
  ▼
Accepted
Rejected
Deferred
```

---

# RFC States

---

## Draft

The RFC is being written.

---

## Proposed

The RFC is open for discussion.

---

## Accepted

The proposal has been approved.

---

## Rejected

The proposal will not be implemented.

---

## Deferred

The proposal may be revisited later.

---

## Implemented

The proposal has been completed.

---

# RFC Directory Structure

```text id="6k5ecm"
docs/
└── rfc/
    ├── README.md
    ├── RFC-0001-memory-model.md
    ├── RFC-0002-context-engine.md
    └── RFC-0003-mcp-integration.md
```

---

# RFC Naming Convention

Format:

```text id="5l1mgh"
RFC-XXXX-title.md
```

Examples:

```text id="vjx1j6"
RFC-0001-memory-model.md

RFC-0002-context-engine.md

RFC-0003-pattern-memory.md
```

---

# RFC Template

Every RFC should follow the same structure.

---

## Header

```markdown id="zqf1jz"
# RFC-XXXX Title

Status: Draft

Author:

Created:

Updated:
```

---

## Summary

Brief overview of the proposal.

---

## Motivation

Why is this change necessary?

What problem does it solve?

---

## Goals

What should happen?

---

## Non-Goals

What is intentionally excluded?

---

## Proposal

Detailed explanation.

---

## Alternatives Considered

What other approaches were evaluated?

Why were they rejected?

---

## Risks

Potential drawbacks.

---

## Migration Strategy

How will existing users transition?

---

## Open Questions

Remaining uncertainties.

---

# RFC Workflow

---

## Step 1

Create RFC draft.

Example:

```text id="ycx9dl"
RFC-0004-pattern-memory.md
```

---

## Step 2

Open pull request.

---

## Step 3

Community discussion begins.

---

## Step 4

Author updates RFC.

---

## Step 5

Maintainers review.

---

## Step 6

Decision made.

---

# Decision Outcomes

---

## Accepted

RFC becomes official project direction.

---

## Rejected

RFC archived.

Reason documented.

---

## Deferred

RFC remains available for future consideration.

---

# Evaluation Criteria

RFCs should be evaluated using:

---

## Alignment

Does it align with:

* Vision
* Mission
* Charter

---

## Simplicity

Does it reduce or increase complexity?

---

## Maintainability

Can the project sustain it long-term?

---

## Community Impact

Does it benefit users and contributors?

---

## Memory Impact

Does it improve repository memory?

---

## Context Impact

Does it improve context quality?

---

## Token Efficiency

Does it reduce unnecessary AI usage?

---

# Design Principles

RFCs should follow RepoNerve principles.

---

## Memory First

Protect repository memory.

---

## Context Second

Improve context generation.

---

## AI Third

Avoid unnecessary AI dependency.

---

## Evidence Over Assumptions

Support decisions with reasoning.

---

## Local First

Preserve local-first architecture.

---

# Maintainer Responsibilities

Maintainers should:

* Encourage discussion
* Provide feedback
* Preserve project direction
* Document decisions

---

# Author Responsibilities

Authors should:

* Clearly explain motivation
* Consider alternatives
* Address feedback
* Update RFCs as needed

---

# Community Participation

Community members are encouraged to:

* Ask questions
* Challenge assumptions
* Suggest alternatives
* Identify risks

Constructive disagreement is healthy.

---

# RFC Number Allocation

RFC numbers are assigned sequentially.

Example:

```text id="mthm4x"
RFC-0001

RFC-0002

RFC-0003
```

Numbers are never reused.

---

# Historical Record

RFCs are permanent project records.

Even rejected RFCs should remain accessible.

They provide valuable historical context.

---

# Example Future RFCs

```text id="nynr5k"
RFC-0001 Memory Model

RFC-0002 Context Engine

RFC-0003 MCP Integration

RFC-0004 Pattern Memory

RFC-0005 Multi-Repository Memory

RFC-0006 Incident Memory
```

---

# Success Criteria

The RFC process succeeds when:

* Major decisions are documented.
* Architectural consistency is preserved.
* Community participation increases.
* Historical reasoning remains discoverable.

---

# Guiding Principle

Important decisions should be remembered, not rediscovered.

The RFC process preserves the reasoning behind RepoNerve itself.
