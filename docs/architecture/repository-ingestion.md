# RepoNerve Repository Ingestion Architecture

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines how RepoNerve discovers, processes, and converts repository artifacts into repository memory.

The ingestion system is responsible for transforming:

* Source code
* Git history
* Pull requests
* Issues
* ADRs
* Documentation

into structured memory entities.

The ingestion system is one of the most critical components of RepoNerve because it determines:

* Memory quality
* Context quality
* Performance
* AI token consumption
* Scalability

---

# Design Goals

The ingestion system must be:

* Incremental
* Deterministic where possible
* AI-assisted only when necessary
* Explainable
* Source traceable
* Repository agnostic
* Efficient

---

# Core Principle

AI should not be the primary parser.

AI should be the final extractor.

The majority of repository understanding should be performed through deterministic techniques.

---

# High-Level Architecture

```text
Repository
     │
     ▼
Artifact Discovery
     │
     ▼
Artifact Parsing
     │
     ▼
Artifact Normalization
     │
     ▼
Memory Extraction
     │
     ▼
Memory Linking
     │
     ▼
Memory Store
```

---

# Ingestion Pipeline

RepoNerve processes repositories through five stages.

---

## Stage 1: Discovery

Purpose:

Identify available repository artifacts.

---

### Inputs

* Repository path
* Git metadata
* Configuration

---

### Outputs

Artifact inventory.

---

### Example

```text
Repository

├── Source Code
├── Git Commits
├── Pull Requests
├── ADRs
└── Documentation
```

---

# Stage 2: Parsing

Purpose:

Convert artifacts into structured records.

---

### Examples

Source code becomes:

```json
{
  "type": "function",
  "name": "CreateUser",
  "file": "user.go"
}
```

---

Git commit becomes:

```json
{
  "hash": "abc123",
  "author": "john",
  "message": "Introduce Redis cache"
}
```

---

ADR becomes:

```json
{
  "title": "Use Redis",
  "status": "accepted"
}
```

---

# Stage 3: Normalization

Purpose:

Convert multiple artifact formats into a common structure.

---

## Unified Artifact

```json
{
  "id": "...",
  "type": "...",
  "source": "...",
  "content": "...",
  "metadata": {}
}
```

---

Benefits:

* Simplifies extraction
* Simplifies indexing
* Reduces complexity

---

# Stage 4: Memory Extraction

Purpose:

Create repository memory.

---

### Input

Normalized artifacts.

---

### Output

Memory entities.

---

Example:

```text
Commit

"Introduce Redis cache"
```

produces:

```text
Event

Redis introduced
```

---

Potentially:

```text
Decision

Use Redis
```

---

# Stage 5: Memory Linking

Purpose:

Connect isolated memories.

---

Example:

```text
ADR
      │
      ▼
Decision
      │
      ▼
PR
      │
      ▼
Commit
      │
      ▼
Fact
```

---

This creates repository understanding.

---

# Artifact Types

---

# Source Code

Priority:

High

---

## Purpose

Generate:

* Facts
* Relationships

---

## Extraction Strategy

Deterministic

---

## Technologies

Tree-sitter

Language parsers

AST analysis

---

## Example Outputs

```text
AuthService USES Redis

BillingService CALLS PaymentGateway
```

---

## AI Usage

Not required.

---

# Git Commits

Priority:

High

---

## Purpose

Generate:

* Events
* Facts
* Relationships

---

## Example

```text
Commit:
Add Redis cache
```

Produces:

```text
Event:
Redis introduced
```

---

## AI Usage

Optional.

Only for decision extraction.

---

# Pull Requests

Priority:

Very High

---

## Purpose

Generate:

* Decisions
* Intent
* Tradeoffs

---

## Example

Discussion:

```text
Redis chosen because database latency
is too high.
```

Produces:

```text
Decision:
Use Redis

Intent:
Reduce latency
```

---

## AI Usage

Recommended.

This is one of the highest-value AI extraction points.

---

# ADRs

Priority:

Very High

---

## Purpose

Generate:

* Decisions
* Alternatives
* Tradeoffs

---

## Example

ADR:

```text
Use Kafka
```

Produces:

```text
Decision:
Use Kafka

Alternatives:
RabbitMQ

Tradeoff:
Operational complexity
```

---

## AI Usage

Optional.

Mostly structured parsing.

---

# Documentation

Priority:

Medium

---

## Purpose

Generate:

* Facts
* Intent
* Ownership

---

## Example

```text
Authentication owned by Platform Team
```

Produces:

```text
Ownership Memory
```

---

# Issues

Future Phase

---

## Purpose

Generate:

* Intent
* Historical context

---

## Example

```text
Database latency issue
```

Produces:

```text
Intent:
Reduce latency
```

---

# Incremental Indexing

Full repository scans should be avoided.

---

## First Run

```text
Repository
      ▼
Full Scan
```

---

## Subsequent Runs

```text
Repository
      ▼
Changed Artifacts Only
```

---

Benefits:

* Faster indexing
* Lower resource usage
* Lower AI usage

---

# AI Extraction Strategy

RepoNerve should minimize AI consumption.

---

## Deterministic First

Always attempt:

* Parsing
* AST analysis
* Metadata extraction

before AI.

---

## AI Last

Use AI only for:

* Intent extraction
* Decision extraction
* Tradeoff extraction
* Historical reasoning

---

# AI Extraction Workflow

```text
Artifact
      │
      ▼
Deterministic Analysis
      │
      ▼
Memory Candidate
      │
      ▼
AI Enhancement
      │
      ▼
Memory Validation
```

---

# Token Optimization Strategy

One of RepoNerve's primary goals is reducing AI token consumption.

---

## Rule 1

Never send entire repositories.

---

## Rule 2

Never send raw code when structure is sufficient.

---

## Rule 3

Extract memory once.

Reuse indefinitely.

---

## Rule 4

Store extracted knowledge.

Do not repeatedly regenerate it.

---

# Memory Caching

Extracted memory should be cached.

---

Example:

```text
Decision:
Use Redis
```

Should be stored permanently.

Future queries should retrieve memory.

Not regenerate it.

---

# Source Traceability

Every memory object must reference evidence.

---

Example:

```text
Decision:
Use Redis

Source:
ADR-14

PR-143

Commit abc123
```

---

# Failure Handling

When extraction confidence is low:

```text
Memory
     ▼
Needs Review
```

instead of:

```text
Memory
     ▼
Stored As Truth
```

---

# Future Sources

Future ingestion support:

* Jira
* Linear
* Incident systems
* Slack exports
* Design documents
* Meeting notes

These are intentionally out of scope for MVP.

---

# Success Criteria

The ingestion pipeline succeeds when:

* Repository artifacts become memory.
* Memory remains traceable.
* AI usage remains minimal.
* Incremental scans are efficient.
* Memory quality improves over time.

---

# Guiding Principle

Do not repeatedly analyze repositories.

Analyze once.

Preserve memory.

Reuse knowledge forever.
