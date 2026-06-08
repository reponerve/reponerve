# ISSUE-049 — Reviewer Recommendation Engine

Status: Implemented

Milestone: v0.9.0-alpha

---

# Objective

Implement the Reviewer Recommendation Engine.

The Reviewer Recommendation Engine helps identify appropriate reviewers for repository changes.

It answers:

- Who should review this?
- Which contributors have relevant expertise?
- Which contributors are most connected to impacted repository knowledge?

---

# Background

Ownership Intelligence identifies:

- Contributors
- Expertise
- Repository ownership

Knowledge Graph Intelligence identifies:

- Repository relationships
- Impact chains

Reviewer Recommendation combines these capabilities into evidence-backed reviewer recommendations.

---

# Philosophy

Evidence First.

Reviewer recommendations are guidance.

Reviewer recommendations are not ownership facts.

Every recommendation must include:

- Evidence
- Explanation
- Score

Recommendations must remain deterministic.

---

# Scope

Create:

internal/intelligence/reviewers/

Files:

- models.go
- service.go
- service_test.go

---

# Architecture Requirements

Reuse:

- Ownership Intelligence
- Knowledge Graph Intelligence
- Impact Analysis

Do NOT:

- Access SQLite directly
- Execute Git commands
- Re-scan repositories

Reviewer Recommendation consumes repository knowledge.

Reviewer Recommendation does not create repository knowledge.

---

# Models

Implement:

type ReviewerRecommendation struct {
    ContributorID string

    Score float64

    EvidenceJSON string

    Explanation string
}

---

Implement:

type ReviewerRecommendationReport struct {
    Recommendations []*ReviewerRecommendation
}

---

# Recommendation Inputs

Recommendations should consider:

- Contributor expertise
- Ownership participation
- Impact analysis results
- Knowledge graph participation

The exact scoring must remain deterministic.

---

# Recommendation Categories

Support:

## Repository Reviewers

Answers:

Who are the strongest reviewers across the repository?

---

## Domain Reviewers

Answers:

Who should review changes in a repository domain?

---

## Impact Reviewers

Answers:

Who should review changes affecting impacted repository knowledge?

---

# Scoring Rules

Scoring must be deterministic.

Suggested inputs:

- Expertise score
- Ownership participation
- Graph participation
- Impact participation

Scores must be reproducible.

---

# Evidence Requirements

Every recommendation must contain:

- EvidenceJSON
- Explanation
- Score

Recommendations without evidence are invalid.

---

# Explanation Requirements

Examples:

Contributor A is recommended because they have expertise in the impacted domain and participate in related repository knowledge.

Contributor B is recommended because multiple impacted entities are connected to their expertise areas.

Explanations must be deterministic.

---

# Ordering

Sort recommendations by:

1. Score descending
2. ContributorID ascending

The same repository state must produce identical ordering.

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

func (s *Service) RecommendRepositoryReviewers(
    ctx context.Context,
    repositoryID string,
) (*ReviewerRecommendationReport, error)

---

Implement:

func (s *Service) RecommendDomainReviewers(
    ctx context.Context,
    repositoryID string,
    domain string,
) (*ReviewerRecommendationReport, error)

---

Implement:

func (s *Service) RecommendImpactReviewers(
    ctx context.Context,
    repositoryID string,
    entityID string,
) (*ReviewerRecommendationReport, error)

---

# Validation

Validate:

- Contributor exists
- Score exists
- Evidence exists
- Explanation exists

Reject invalid recommendations.

---

# Unit Tests

Cover:

- Empty repositories
- Repository reviewers
- Domain reviewers
- Impact reviewers
- Deterministic scoring
- Deterministic ordering
- Evidence generation
- Explanation generation

---

# Integration Tests

Create SQLite-backed integration tests.

Verify:

Ownership
↓
Knowledge Graph
↓
Impact Analysis
↓
Reviewer Recommendations

Verify:

- Recommendation generation
- Evidence preservation
- Ordering
- Determinism

---

# Constraints

Do NOT implement:

- Change Planning
- MCP tools

Only implement Reviewer Recommendation.

---

# Acceptance Criteria

Reviewer recommendations are generated successfully.

Recommendations contain evidence.

Recommendations contain explanations.

Scoring remains deterministic.

Ordering remains deterministic.

All tests pass.
