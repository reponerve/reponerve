# RepoNerve GitHub Project Structure

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines how RepoNerve work is organized in GitHub.

The goal is to transform strategy, architecture, and roadmap documents into actionable engineering work.

---

# Project Management Philosophy

We organize work using:

```text
Vision
   ▼
Milestones
   ▼
Epics
   ▼
Issues
   ▼
Pull Requests
```

---

# GitHub Labels

## Type Labels

```text
type:feature
type:bug
type:docs
type:test
type:refactor
type:research
type:rfc
```

---

## Priority Labels

```text
priority:p0
priority:p1
priority:p2
priority:p3
```

---

## Area Labels

```text
area:scanner
area:memory
area:query
area:context
area:mcp
area:storage
area:search
area:docs
area:cli
area:architecture
```

---

## Status Labels

```text
status:blocked
status:ready
status:in-progress
status:review
status:done
```

---

# Project Board Structure

Columns:

```text
Backlog

Ready

In Progress

Review

Done
```

---

# Phase 0 Epic

## EPIC-001 Foundation

### Issues

```text
ISSUE-001 Initialize Repository

ISSUE-002 Configure Go Modules

ISSUE-003 Configure CI Pipeline

ISSUE-004 Configure Linting

ISSUE-005 Configure Testing

ISSUE-006 Add Documentation Structure
```

---

# Phase 1 Epic

## EPIC-002 Repository Scanner

### Issues

```text
ISSUE-010 Repository Discovery

ISSUE-011 Git Scanner

ISSUE-012 Commit Parser

ISSUE-013 ADR Parser

ISSUE-014 Markdown Parser

ISSUE-015 Incremental Scanner
```

---

# Phase 2 Epic

## EPIC-003 Memory Engine

### Issues

```text
ISSUE-020 Fact Extraction

ISSUE-021 Event Extraction

ISSUE-022 Decision Extraction

ISSUE-023 Intent Extraction

ISSUE-024 Ownership Extraction

ISSUE-025 Memory Linking
```

---

# Phase 3 Epic

## EPIC-004 Query Engine

### Issues

```text
ISSUE-030 Search Infrastructure

ISSUE-031 Query Parser

ISSUE-032 Evidence Retrieval

ISSUE-033 Ask Command

ISSUE-034 Explain Command
```

---

# Phase 4 Epic

## EPIC-005 Context Engine

### Issues

```text
ISSUE-040 Intent Detection

ISSUE-041 Context Retrieval

ISSUE-042 Ranking Engine

ISSUE-043 Context Packs

ISSUE-044 Context Command
```

---

# Phase 5 Epic

## EPIC-006 MCP Skills

### Issues

```text
ISSUE-050 MCP Server

ISSUE-051 Repository Memory Tool

ISSUE-052 Explain Component Tool

ISSUE-053 Context Pack Tool

ISSUE-054 MCP Documentation
```

---

# Definition of Done

An issue is complete when:

* Code implemented
* Tests added
* Documentation updated
* CI passing
* Review completed

---

# Release Strategy

## Alpha

```text
Scanner
Memory Engine
```

---

## Beta

```text
Query Engine
Context Engine
```

---

## RC

```text
MCP Skills
Performance Improvements
```

---

## v1.0

```text
Stable CLI
Stable Memory
Stable Context
Stable MCP
```

---

# Guiding Principle

Every issue should move RepoNerve closer to becoming the memory and context engine for software repositories.
