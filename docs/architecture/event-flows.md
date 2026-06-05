# RepoNerve Event Flows

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the operational flows of RepoNerve.

Event flows describe how requests move through the system.

These flows serve as the implementation blueprint for:

* CLI commands
* Service orchestration
* Memory generation
* Context generation
* Query execution

---

# Design Principles

## Deterministic First

Use deterministic analysis whenever possible.

---

## Memory Before AI

Always retrieve existing memory before generating new information.

---

## Evidence Required

All outputs must remain traceable to repository artifacts.

---

## Incremental Processing

Avoid reprocessing repository artifacts unnecessarily.

---

# System Actors

---

## Developer

Human user interacting through CLI.

---

## RepoNerve CLI

Primary interface.

---

## Ingestion Engine

Repository processing system.

---

## Memory Engine

Knowledge extraction system.

---

## Query Engine

Memory retrieval system.

---

## Context Engine

Task-specific context generation system.

---

## Memory Store

Persistent repository memory.

---

# Flow 1: Repository Initialization

Command:

```bash
reponerve init
```

---

## Objective

Initialize RepoNerve inside a repository.

---

## Flow

```text
Developer
    │
    ▼
reponerve init
    │
    ▼
Validate Repository
    │
    ▼
Create Workspace
    │
    ▼
Create Configuration
    │
    ▼
Initialize Database
    │
    ▼
Success
```

---

## Detailed Steps

### Step 1

Validate repository.

Checks:

* Git repository exists
* Permissions available

---

### Step 2

Create workspace.

```text
.reponerve/
```

---

### Step 3

Create configuration.

```text
.reponerve/config.yaml
```

---

### Step 4

Initialize SQLite database.

```text
.reponerve/memory.db
```

---

### Step 5

Run migrations.

---

### Result

RepoNerve workspace ready.

---

# Flow 2: Initial Repository Scan

Command:

```bash
reponerve scan
```

---

## Objective

Build repository memory.

---

## Flow

```text
Repository
     │
     ▼
Discovery
     │
     ▼
Parsing
     │
     ▼
Normalization
     │
     ▼
Memory Extraction
     │
     ▼
Memory Linking
     │
     ▼
Storage
```

---

## Detailed Steps

### Step 1

Discover artifacts.

Sources:

* Source code
* Git history
* Pull requests
* ADRs
* Documentation

---

### Step 2

Parse artifacts.

Convert raw artifacts into structured records.

---

### Step 3

Normalize records.

Create unified internal format.

---

### Step 4

Extract memory.

Generate:

* Facts
* Events
* Decisions
* Intent
* Ownership
* Relationships

---

### Step 5

Link memory.

Connect:

```text
ADR
  ▼
Decision
  ▼
PR
  ▼
Commit
```

---

### Step 6

Store memory.

Persist in SQLite.

---

### Result

Repository memory available.

---

# Flow 3: Incremental Scan

Command:

```bash
reponerve scan
```

after initial indexing.

---

## Objective

Process only changes.

---

## Flow

```text
Repository
     │
     ▼
Detect Changes
     │
     ▼
Process Changed Artifacts
     │
     ▼
Update Memory
     │
     ▼
Update Relationships
     │
     ▼
Complete
```

---

## Benefits

* Faster execution
* Lower CPU usage
* Lower AI usage

---

# Flow 4: Repository Question

Command:

```bash
reponerve ask
```

Example:

```bash
reponerve ask "Why was Redis introduced?"
```

---

## Objective

Retrieve repository memory.

---

## Flow

```text
Question
    │
    ▼
Intent Analysis
    │
    ▼
Memory Search
    │
    ▼
Relationship Expansion
    │
    ▼
Evidence Collection
    │
    ▼
Response Generation
```

---

## Detailed Steps

### Step 1

Classify query.

Example:

```text
Why
```

becomes:

```text
Decision Query
```

---

### Step 2

Search memory.

Retrieve:

* Decisions
* Events
* Intent

---

### Step 3

Expand relationships.

Retrieve related memories.

---

### Step 4

Collect evidence.

Retrieve:

* PRs
* ADRs
* Commits

---

### Step 5

Generate response.

---

### Result

Evidence-backed answer.

---

# Flow 5: Explain Component

Command:

```bash
reponerve explain services/auth
```

---

## Objective

Explain repository components.

---

## Flow

```text
Component
     │
     ▼
Component Lookup
     │
     ▼
Relationship Discovery
     │
     ▼
Memory Retrieval
     │
     ▼
Evidence Collection
     │
     ▼
Explanation
```

---

## Example Output

```text
Purpose

History

Dependencies

Ownership

Related Decisions

Evidence
```

---

# Flow 6: Context Generation

Command:

```bash
reponerve context
```

Example:

```bash
reponerve context "Add MFA support"
```

---

## Objective

Generate task-specific repository context.

---

## Flow

```text
Task
   │
   ▼
Intent Detection
   │
   ▼
Memory Retrieval
   │
   ▼
Relationship Expansion
   │
   ▼
Context Ranking
   │
   ▼
Context Assembly
```

---

## Detailed Steps

### Step 1

Identify task intent.

Example:

```text
Authentication
```

---

### Step 2

Retrieve relevant memory.

Examples:

* Decisions
* ADRs
* Ownership

---

### Step 3

Expand relationships.

Retrieve:

* Related services
* Related implementations

---

### Step 4

Rank relevance.

Remove noise.

---

### Step 5

Build context pack.

---

### Result

Task-specific context.

---

# Flow 7: Decision Extraction

Triggered during scanning.

---

## Objective

Extract architectural decisions.

---

## Flow

```text
PR
 │
 ▼
Parser
 │
 ▼
Decision Detector
 │
 ▼
AI Enhancement
 │
 ▼
Validation
 │
 ▼
Decision Memory
```

---

## Example

Input:

```text
Redis selected because database
latency exceeded SLA.
```

---

Output:

```text
Decision:
Use Redis

Intent:
Reduce latency
```

---

# Flow 8: Intent Extraction

Triggered during scanning.

---

## Objective

Identify goals.

---

## Flow

```text
Artifact
   │
   ▼
Intent Detector
   │
   ▼
Intent Memory
```

---

## Example

Input:

```text
Reduce authentication latency.
```

Output:

```text
Intent:
Reduce Latency
```

---

# Flow 9: Evidence Retrieval

Triggered during queries.

---

## Objective

Provide explainability.

---

## Flow

```text
Memory
   │
   ▼
Evidence Lookup
   │
   ▼
Source Retrieval
   │
   ▼
Evidence Bundle
```

---

## Sources

* Commits
* Pull Requests
* ADRs
* Documentation

---

# Flow 10: MCP Context Request

Future Phase.

---

## Actor

AI Coding Agent

---

## Example Tool

```text
get_context_pack
```

---

## Flow

```text
AI Agent
      │
      ▼
MCP Tool
      │
      ▼
Context Engine
      │
      ▼
Memory Store
      │
      ▼
Context Pack
```

---

## Result

Repository-specific context.

---

# Error Handling Flow

---

## Memory Not Found

```text
Question
     │
     ▼
Search
     │
     ▼
No Results
     │
     ▼
Return Explanation
```

---

Example:

```text
No repository memory found
for requested topic.
```

---

## Low Confidence

```text
Memory
     │
     ▼
Low Confidence
     │
     ▼
Flag Result
```

---

Example:

```text
Confidence: Weakly Inferred
```

---

# Performance Objectives

---

## Init

< 5 seconds

---

## Incremental Scan

< 30 seconds

Typical repository.

---

## Ask

< 2 seconds

Cached memory.

---

## Explain

< 3 seconds

---

## Context

< 5 seconds

---

# Event Flow Summary

| Flow   | Command  | Purpose                   |
| ------ | -------- | ------------------------- |
| EF-001 | init     | Initialize RepoNerve      |
| EF-002 | scan     | Initial memory generation |
| EF-003 | scan     | Incremental update        |
| EF-004 | ask      | Memory retrieval          |
| EF-005 | explain  | Component understanding   |
| EF-006 | context  | Context pack generation   |
| EF-007 | internal | Decision extraction       |
| EF-008 | internal | Intent extraction         |
| EF-009 | internal | Evidence retrieval        |
| EF-010 | MCP      | AI context generation     |

---

# Guiding Principle

Every flow should move the repository closer to a state where knowledge no longer needs to be rediscovered.
