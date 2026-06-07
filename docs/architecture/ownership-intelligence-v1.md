# Ownership Intelligence V1

## Purpose

Ownership Intelligence extends RepoNerve's repository knowledge graph with the human dimension.

Current RepoNerve capabilities answer:

* What happened?
* Why did it happen?
* What is affected?

Ownership Intelligence introduces:

* Who contributed?
* Who has expertise?
* Who is most familiar with an area?
* Who should be involved?

Ownership Intelligence must remain deterministic, traceable, and evidence-based.

---

# Philosophy

Evidence First.

Ownership Second.

Recommendations Third.

Ownership must be derived from repository evidence.

Ownership must never rely on subjective judgments or AI-generated conclusions.

---

# Goals

Enable RepoNerve to:

* Identify repository contributors
* Detect expertise areas
* Associate contributors with repository knowledge
* Generate ownership recommendations
* Expose ownership intelligence through Query Engine, Context Engine, MCP, and Agent Intelligence

---

# Non-Goals

Ownership Intelligence does not:

* Assign organizational responsibility
* Replace CODEOWNERS
* Determine team structures
* Infer management relationships
* Make subjective assessments

---

# Architecture

Repository
↓
Ingestion Engine
↓
Memory Graph
↓
Ownership Intelligence
↓
Query Engine
↓
Context Engine
↓
MCP
↓
Agent Intelligence

---

# Ownership Model

Ownership Intelligence is built on four concepts:

## Contributor

A contributor is an identifiable repository participant derived from repository artifacts.

Examples:

* Git commit authors
* ADR authors
* Future repository contributors

Contributor is the foundational ownership entity.

---

## Expertise

Expertise represents evidence-backed familiarity with repository concepts.

Examples:

* Authentication
* Storage
* API Gateway
* Context Engine

Expertise is derived from repository activity.

---

## Knowledge Domain

A knowledge domain represents a logical repository area.

Examples:

* Authentication
* Infrastructure
* Persistence
* MCP
* Agent Intelligence

Knowledge domains organize expertise.

---

## Ownership Recommendation

Ownership recommendations are computed from:

* Contribution history
* Expertise evidence
* Recency
* Repository activity

Ownership is therefore a derived conclusion rather than a primary entity.

---

# Knowledge Graph Extensions

Current graph:

Intent
↓
Decision
↓
Event

Fact
↓
Decision

Extended graph:

Contributor
↓
Event

Contributor
↓
Decision

Contributor
↓
Fact

Contributor
↓
Knowledge Domain

Knowledge Domain
↓
Decision

Knowledge Domain
↓
Fact

---

# Ownership Types

Ownership Intelligence distinguishes several ownership concepts.

## Authorship

Who originally created something.

Examples:

* Commit author
* ADR author

---

## Expertise

Who demonstrates repeated familiarity with an area.

Examples:

* Frequent contributor
* Repeated domain involvement

---

## Maintainership

Who actively participates in ongoing repository evolution.

Derived from:

* Recent activity
* Sustained contribution patterns

---

## Reviewer

Who may be appropriate to review related changes.

Derived from:

* Expertise
* Historical activity

---

# Ownership Sources

## Git History

Primary source.

Provides:

* Author name
* Author email
* Commit timestamps
* Contribution frequency

---

## ADR Metadata

Provides:

* Decision authorship
* Architectural involvement

---

## Memory Graph

Provides:

* Decisions
* Facts
* Events
* Relationships

Used to connect contributors with repository knowledge.

---

# Contributor Model

Proposed model:

```go
type Contributor struct {
    ID string

    RepositoryID string

    Name string

    Email string

    FirstSeen time.Time

    LastSeen time.Time

    CommitCount int
}
```

Contributor records must be deterministic and deduplicated.

---

# Expertise Model

Proposed model:

```go
type Expertise struct {
    ID string

    RepositoryID string

    ContributorID string

    Domain string

    Score float64
}
```

Scores must be derived from objective repository evidence.

---

# Ownership Relationships

Proposed relationship types:

CONTRIBUTOR_CREATED_EVENT

CONTRIBUTOR_MADE_DECISION

CONTRIBUTOR_SUPPORTS_FACT

CONTRIBUTOR_EXPERT_IN_DOMAIN

DOMAIN_RELATES_TO_DECISION

DOMAIN_RELATES_TO_FACT

All relationships must be traceable to repository evidence.

---

# Query Capabilities

Examples:

Who contributed most to authentication?

Who authored this decision?

Who has expertise in storage?

Who should review changes related to MCP?

---

# Context Integration

Ownership data may enrich repository context.

Examples:

Key Contributors

Relevant Experts

Domain Specialists

Context generation remains deterministic.

---

# MCP Integration

Future MCP capabilities:

list_contributors

get_contributor

list_expertise

trace_contributor

recommend_reviewers

Ownership data must be exposed through existing MCP architecture.

---

# Agent Intelligence Integration

Ownership Intelligence complements:

* Repository Onboarding
* Repository Q&A
* Architectural Guidance
* Impact Analysis

Future questions:

Who knows this area?

Who should review this change?

Who created this decision?

Who has expertise in this domain?

---

# Constraints

Do not introduce:

* AI-generated ownership
* LLM-based scoring
* Embeddings
* Vector databases
* Subjective ranking

Ownership must remain evidence-based.

---

# Success Criteria

RepoNerve can answer:

What happened?

Why?

What is affected?

Who knows about it?

Who should be involved?

using deterministic repository evidence.

---

Version: 1.0

Status: Draft
