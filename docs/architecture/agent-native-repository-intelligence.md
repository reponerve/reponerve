# Agent-Native Repository Intelligence

Version: v1.0

Status: Draft

---

# Overview

Agent-Native Repository Intelligence is the architectural direction for RepoNerve v1.0.

The objective is to transform repository intelligence into agent-consumable workflows.

---

# Existing Architecture

Repository
↓
Ingestion
↓
Memory
↓
Context
↓
Ownership
↓
Knowledge Graph
↓
Repository Intelligence
↓
MCP

This stack already exists.

---

# v1.0 Extension

Repository
↓
Ingestion
↓
Memory
↓
Context
↓
Ownership
↓
Knowledge Graph
↓
Repository Intelligence
↓
Agent Intelligence Layer
↓
MCP
↓
AI Agents

---

# Agent Intelligence Layer

The Agent Intelligence Layer does not create repository knowledge.

The Agent Intelligence Layer consumes repository intelligence.

Responsibilities:

* Context Packaging
* Knowledge Retrieval
* Workflow Orchestration
* Agent Guidance

---

# Architectural Rule

Repository Intelligence remains authoritative.

Agent Intelligence consumes Repository Intelligence.

Responsibilities:

Repository Intelligence
↓
Generates Intelligence

Agent Intelligence
↓
Packages Intelligence

---

# Core Capabilities

## Agent Context Builder

Generate structured repository context for agents.

Examples:

* Repository Overview
* Domain Overview
* Contributor Overview

---

## Repository Search

Retrieve repository knowledge efficiently.

Search:

* Decisions
* Facts
* Events
* Contributors
* Expertise

---

## Workflow Intelligence

Support workflows such as:

* Code Review Preparation
* Repository Onboarding
* Change Planning
* Knowledge Discovery

---

## Session Intelligence

Allow agents to maintain repository-aware sessions.

---

# Dependency Direction

Storage
↓
Readers
↓
Memory
↓
Ownership
↓
Knowledge Graph
↓
Repository Intelligence
↓
Agent Intelligence
↓
MCP

Dependency direction must remain one-way.

---

# Evidence Requirements

All agent-facing outputs must preserve:

* Evidence
* Explanations
* Scores
* Priorities

No information loss is permitted.

---

# Determinism Requirements

Agent Intelligence must remain deterministic.

Agent-specific formatting may change.

Repository conclusions must not.

---

# Future Capabilities

Potential future additions:

* Repository Health Intelligence
* Knowledge Risk Analysis
* Architecture Drift Detection
* Repository Evolution Analysis

These remain outside the scope of v1.0.

---

# Summary

Agent-Native Repository Intelligence extends RepoNerve from repository intelligence infrastructure into an intelligence platform for AI agents while preserving determinism, explainability, and evidence-based reasoning.
