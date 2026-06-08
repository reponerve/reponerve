# ISSUE-050 — Change Planning Engine

Status: Planned

Milestone: v0.9.0-alpha

---

# Objective

Implement the Change Planning Engine.

The Change Planning Engine helps humans and AI systems understand what repository knowledge should be examined before making changes.

It answers:

- If I change this, what should I examine?
- Which repository knowledge is likely to be affected?
- What should I review before implementing changes?
- Which contributors may be relevant?

---

# Background

Knowledge Graph Intelligence can identify impact chains.

Repository Intelligence should transform impact information into actionable change guidance.

Change Planning builds on:

- Impact Analysis
- Knowledge Graph Intelligence
- Repository Context
- Ownership Intelligence

The goal is to help contributors make safer and more informed repository changes.

---

# Philosophy

Evidence First.

Change plans are recommendations.

Change plans are not facts.

Every recommendation must include:

- Evidence
- Explanation

Change Planning must remain deterministic.

---

# Scope

Create:

internal/intelligence/changeplan/

Files:

- models.go
- service.go
- service_test.go

---

# Architecture Requirements

Reuse:

- Impact Analysis
- Knowledge Graph Intelligence
- Ownership Intelligence
- Context Engine

Do NOT:

- Access SQLite directly
- Execute Git commands
- Re-scan repositories
- Generate new graph relationships

Change Planning consumes repository knowledge.

It does not create repository knowledge.

---

# Models

Implement:

type ChangePlanItem struct {
    EntityType string

    EntityID string

    Priority int

    EvidenceJSON string

    Explanation string
}

---

Implement:

type ChangePlan struct {
    Items []*ChangePlanItem
}

---

# Change Plan Categories

Support:

## Decision Change Plan

Answers:

What should be examined before changing a repository decision?

---

## Fact Change Plan

Answers:

What should be examined before changing a repository fact?

---

## Event Change Plan

Answers:

What should be examined before changing a repository event?

---

## Contributor Change Plan

Answers:

Which repository knowledge should be reviewed before modifying contributor-owned areas?

---

# Plan Construction Rules

Change Plans should consider:

- Impact paths
- Knowledge graph relationships
- Ownership information
- Repository context

Every plan item must be traceable back to repository evidence.

---

# Priority Rules

Priority values must be deterministic.

Suggested ordering:

1 = Highest priority

Higher priority should be assigned to:

- Directly impacted entities
- Frequently referenced entities
- Strong ownership dependencies

The same repository state must produce the same priorities.

---

# Evidence Requirements

Every ChangePlanItem must contain:

- EvidenceJSON
- Explanation

Items without evidence are invalid.

---

# Explanation Requirements

Examples:

Decision:

This decision should be reviewed because it directly depends on the changed decision.

Fact:

This fact participates in the impacted knowledge chain.

Contributor:

This contributor owns repository areas connected to the planned change.

Explanations must be deterministic.

---

## Impact Authority Rule

Impact Analysis is the canonical source of repository impact information.

Change Planning must consume Impact Analysis results.

Change Planning must not independently compute impact relationships.

Responsibilities:

Impact Analysis
↓
Determines Impact

Change Planning
↓
Determines Action Priority

---

Change Plans should consider:

- Impact Analysis results
- Ownership information
- Repository context

---

# Ordering

Sort by:

1. Priority ascending
2. EntityType ascending
3. EntityID ascending

Output must be reproducible.

---

# Service

Implement:

type Service struct {
}

---

Constructor:

func NewService(...) *Service

---

# APIs

Implement:

func (s *Service) GenerateDecisionPlan(
    ctx context.Context,
    repositoryID string,
    decisionID string,
) (*ChangePlan, error)

---

Implement:

func (s *Service) GenerateFactPlan(
    ctx context.Context,
    repositoryID string,
    factID string,
) (*ChangePlan, error)

---

Implement:

func (s *Service) GenerateEventPlan(
    ctx context.Context,
    repositoryID string,
    eventID string,
) (*ChangePlan, error)

---

Implement:

func (s *Service) GenerateContributorPlan(
    ctx context.Context,
    repositoryID string,
    contributorID string,
) (*ChangePlan, error)

---

# Validation

Validate:

- Entity exists
- Priority exists
- Evidence exists
- Explanation exists

Reject invalid plan items.

---

# Unit Tests

Cover:

- Empty repositories
- Decision plans
- Fact plans
- Event plans
- Contributor plans
- Deterministic priority generation
- Deterministic ordering
- Evidence generation
- Explanation generation

---

# Integration Tests

Create SQLite-backed integration tests.

Verify:

Impact Analysis
↓
Knowledge Graph
↓
Ownership
↓
Change Planning

Verify:

- Plan generation
- Evidence preservation
- Ordering
- Determinism

---

# Constraints

Do NOT implement:

- MCP tools

Only implement Change Planning.

---

# Acceptance Criteria

Change plans are generated successfully.

Plans contain evidence.

Plans contain explanations.

Priorities remain deterministic.

Ordering remains deterministic.

All tests pass.