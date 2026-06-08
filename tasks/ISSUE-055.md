# ISSUE-055 — Workflow Intelligence

Status: Implemented

Milestone: v1.0

---

# Objective

Implement Workflow Intelligence.

Workflow Intelligence packages existing repository intelligence into reusable engineering workflows for AI agents and engineers.

It answers:

* How should an engineer onboard to this repository?
* How should an agent prepare for a code review?
* How should an agent prepare for a repository change?
* How should repository knowledge be explored?

---

# Background

RepoNerve already provides:

* Repository Intelligence
* Agent Context Packages
* Repository Search
* Agent Session Intelligence

However, agents still consume these capabilities independently.

Workflow Intelligence creates reusable workflow packages.

---

# Philosophy

Evidence First.

Workflow Intelligence orchestrates repository intelligence.

Workflow Intelligence does not generate repository intelligence.

Workflow Intelligence must remain deterministic.

---

# Workflow Authority Rule

Workflow Intelligence consumes:

* Discovery
* Learning Paths
* Reviewer Recommendations
* Change Plans
* Agent Context Packages
* Repository Search
* Agent Sessions

Workflow Intelligence must not:

* Generate discovery results
* Generate learning paths
* Generate reviewer recommendations
* Generate change plans
* Generate repository intelligence
* Generate graph relationships

Responsibilities:

Repository Intelligence
↓
Produces Intelligence

Workflow Intelligence
↓
Packages Workflows

Workflow Intelligence is an orchestration layer.

It is not an intelligence layer.

---

## Workflow Composition Rule

Workflow Intelligence composes existing intelligence.

Workflow Intelligence does not create intelligence.

Workflow Intelligence must not:

- Compute repository importance
- Compute reviewer relevance
- Compute change priority
- Compute impact
- Compute ownership

Workflow Intelligence packages existing outputs.

Responsibilities:

Discovery
↓
Importance

Reviewers
↓
Reviewer Relevance

Change Planning
↓
Change Priority

Workflow Intelligence
↓
Workflow Composition

---

# Workflow Reconstruction Rule

The same repository state must generate the same workflow.

Workflow outputs must be reproducible.

Workflow outputs must not contain:

* Hidden state
* Runtime-only intelligence
* Agent-generated conclusions

Workflow packages are deterministic orchestration artifacts.

---

# Architecture Requirements

Reuse:

* discovery.Service
* learning.Service
* reviewers.Service
* changeplan.Service
* agentcontext.Service
* agentsearch.Service
* agentsession.Service

Do NOT:

* Access SQLite directly
* Execute Git commands
* Re-scan repositories
* Recompute repository intelligence

---

# Scope

Create:

internal/agent/workflow/

Files:

* models.go
* service.go
* service_test.go

Package name:

workflow

---

# Models

Implement:

```go
type WorkflowArtifact struct {
    ArtifactType string `json:"artifact_type"`

    Source string `json:"source"`

    Data json.RawMessage `json:"data"`
}
```

Supported ArtifactType values:

* session
* context_package
* search_result
* discovery_report
* learning_path
* reviewer_report
* change_plan

---

Implement:

```go
type WorkflowPackage struct {
    WorkflowID string `json:"workflow_id"`

    WorkflowType string `json:"workflow_type"`

    RepositoryID string `json:"repository_id"`

    Version string `json:"version"`

    Artifacts []*WorkflowArtifact `json:"artifacts"`
}
```

Supported WorkflowType values:

* onboarding
* review_preparation
* change_preparation
* knowledge_exploration

---

# Validation

Implement:

```go
func ValidateWorkflow(
    workflow *WorkflowPackage,
) error
```

Validate:

* WorkflowID exists
* WorkflowType valid
* RepositoryID exists
* Artifacts valid
* ArtifactType valid
* Source exists
* Data exists

Reject invalid workflows.

WorkflowType validation must enforce:

- onboarding
- review_preparation
- change_preparation
- knowledge_exploration

Unknown workflow types must be rejected.

---

# Artifact Ownership Rule

Artifacts remain owned by their originating subsystem.

Workflow Intelligence must not modify artifact contents.

Workflow Intelligence only packages artifacts.

Examples:

Discovery Report
↓
Owned by Discovery Service

Learning Path
↓
Owned by Learning Service

Reviewer Report
↓
Owned by Reviewer Service

Change Plan
↓
Owned by Change Planning Service

Session
↓
Owned by Agent Session Intelligence

Workflow Intelligence must not modify owned artifacts.

---

# Workflow Types

## Repository Onboarding Workflow

Purpose:

Introduce an engineer or agent to a repository.

Artifacts:

1. Repository Session
2. Discovery Report
3. Learning Path

---

## Review Preparation Workflow

Purpose:

Prepare an agent for repository review activity.

Artifacts:

1. Repository Session
2. Reviewer Recommendation Report
3. Repository Search Result
4. Repository Context Package

---

## Change Preparation Workflow

Purpose:

Prepare an agent for repository modifications.

Artifacts:

1. Repository Session
2. Change Plan
3. Repository Search Result
4. Repository Context Package

---

## Knowledge Exploration Workflow

Purpose:

Explore repository knowledge around a topic.

Artifacts:

1. Repository Session
2. Search Result

---

# Workflow Identity

Workflow IDs must be deterministic.

Recommended:

sha256(
repositoryID +
":" +
workflowType +
":" +
identifier
)

Return IDs using prefix:

wrk_

Example:

wrk_abc123...

---

# Ordering

Artifacts remain in fixed workflow order.

Do not sort.

Do not reorder.

Ordering must be deterministic.

---

# Service

Implement:

```go
type Service struct {
}
```

Inject required services.

---

Constructor:

```go
func NewService(...) *Service
```

---

# APIs

Implement:

```go
func (s *Service) BuildOnboardingWorkflow(
    ctx context.Context,
    repositoryID string,
) (*WorkflowPackage, error)
```

---

Implement:

```go
func (s *Service) BuildReviewPreparationWorkflow(
    ctx context.Context,
    repositoryID string,
    query string,
) (*WorkflowPackage, error)
```

---

Implement:

```go
func (s *Service) BuildChangePreparationWorkflow(
    ctx context.Context,
    repositoryID string,
    entityID string,
) (*WorkflowPackage, error)
```

---

Implement:

```go
func (s *Service) BuildKnowledgeExplorationWorkflow(
    ctx context.Context,
    repositoryID string,
    query string,
) (*WorkflowPackage, error)
```

---

# Serialization Rules

Artifacts must preserve upstream structures unchanged.

Use:

```go
json.Marshal(...)
```

followed by:

```go
json.RawMessage(...)
```

No transformation.

No re-ranking.

No re-scoring.

---

# Unit Tests

Cover:

* Validation
* Invalid workflow types
* Invalid artifact types
* Onboarding workflow
* Review workflow
* Change workflow
* Knowledge workflow
* Artifact preservation
* Deterministic IDs
* Deterministic ordering

---

# Integration Tests

Create migration-backed SQLite tests.

Verify:

Repository Intelligence
↓
Agent Context
↓
Search
↓
Session
↓
Workflow

Verify:

* Workflow creation
* Artifact preservation
* Evidence preservation
* Ordering
* Determinism

---

## Persistence Rule

Workflow packages are derived artifacts.

Workflow packages are not persisted.

Persistence remains outside the scope of v1.0.

---

# Constraints

Do NOT:

* Persist workflows
* Add MCP tools
* Add AI reasoning
* Add autonomous planning
* Modify intelligence outputs

Only implement deterministic workflow orchestration.

---

# Acceptance Criteria

Workflow packages are created successfully.

Repository intelligence is reused.

Sessions are reused.

Search results are reused.

Context packages are reused.

No intelligence is recomputed.

No hidden workflow state exists.

Workflows are reconstructible.

Evidence is preserved.

Artifact ownership is preserved.

Workflow provenance is preserved.

Ordering is deterministic.

Workflow IDs are deterministic.

All tests pass.
