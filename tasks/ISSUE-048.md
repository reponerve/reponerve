# ISSUE-048 — Repository Learning Paths

Status: Planned

Milestone: v0.9.0-alpha

---

# Objective

Implement Repository Learning Paths.

Learning Paths help humans and AI systems understand repository knowledge in a structured order.

They answer:

- Where should I start?
- What should I learn next?
- What repository knowledge should be understood before working in an area?

---

# Background

Knowledge Discovery identifies important repository knowledge.

Learning Paths organize repository knowledge into explainable learning sequences.

The goal is to reduce repository onboarding time and accelerate repository understanding.

---

# Philosophy

Evidence First.

Learning paths are recommendations.

Learning paths are not facts.

Every learning path must include:

- Evidence
- Explanations
- Ordered steps

Learning paths must remain deterministic.

---

# Scope

Create:

internal/intelligence/learning/

Files:

- models.go
- service.go
- service_test.go

---

# Architecture Requirements

Reuse:

- Knowledge Discovery Engine
- Context Engine
- Ownership Intelligence
- Knowledge Graph Intelligence

Do NOT:

- Access SQLite directly
- Execute Git commands
- Re-scan repositories

Learning Paths consume repository knowledge.

Learning Paths do not create repository knowledge.

---

## Ranking Rule

Learning Paths must consume Knowledge Discovery results when determining repository knowledge importance.

Learning Paths are responsible for sequencing repository knowledge.

Learning Paths must not independently compute repository importance rankings.

---

# Learning Path Models

Implement:

type LearningStep struct {
    EntityType string

    EntityID string

    Position int

    EvidenceJSON string

    Explanation string
}

---

Implement:

type LearningPath struct {
    Steps []*LearningStep
}

---

# Learning Path Categories

Support:

## Repository Overview Path

Answers:

What should I learn first about the repository?

---

## Domain Learning Path

Answers:

What should I learn before working in a repository domain?

---

## Contributor Learning Path

Answers:

What should I learn before contributing to a repository area?

---

# Path Construction Rules

Learning Paths should prioritize:

1. Repository Context
2. Important Decisions
3. Important Facts
4. Important Events
5. Ownership Knowledge

Earlier steps should provide context for later steps.

Paths must remain deterministic.

---

# Evidence Requirements

Every LearningStep must contain:

- EvidenceJSON
- Explanation

Steps without evidence are invalid.

---

# Explanation Requirements

Examples:

Repository Overview:

This decision is foundational because multiple repository concepts depend on it.

Domain Learning:

This fact should be understood before working in this repository area.

Contributor Learning:

This contributor expertise area is commonly involved in repository changes.

Explanations must be deterministic.

---

# Ordering

Sort Learning Steps by:

1. Position ascending

Position values must be deterministic.

The same repository state must generate the same learning path.

---

# Service

Implement:

type Service struct {
}

---

Constructor:

func NewService(...) *Service

---

APIs

Implement:

func (s *Service) GenerateRepositoryPath(
    ctx context.Context,
    repositoryID string,
) (*LearningPath, error)

---

Implement:

func (s *Service) GenerateDomainPath(
    ctx context.Context,
    repositoryID string,
    domain string,
) (*LearningPath, error)

---

Implement:

func (s *Service) GenerateContributorPath(
    ctx context.Context,
    repositoryID string,
    contributorID string,
) (*LearningPath, error)

---

# Validation

Validate:

- Position exists
- Entity exists
- Evidence exists
- Explanation exists

Reject invalid steps.

---

# Unit Tests

Cover:

- Empty repositories
- Repository overview paths
- Domain paths
- Contributor paths
- Deterministic ordering
- Evidence generation
- Explanation generation

---

# Integration Tests

Create SQLite-backed integration tests.

Verify:

Repository Knowledge
↓
Knowledge Discovery
↓
Learning Paths

Verify:

- Step generation
- Ordering
- Evidence preservation
- Determinism

---

# Constraints

Do NOT implement:

- Reviewer Recommendations
- Change Planning
- MCP tools

Only implement Learning Paths.

---

# Acceptance Criteria

Learning Paths are generated successfully.

Learning Paths preserve evidence.

Learning Paths preserve explanations.

Ordering remains deterministic.

All tests pass.