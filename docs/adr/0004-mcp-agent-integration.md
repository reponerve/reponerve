# 4. MCP as the agent integration surface

## Status

Accepted

RepoNerve exposes repository intelligence and Development Experience through an MCP server over STDIO. MCP tools remain thin wrappers over Query Engine, Context Engine, Graph Engine, and Development Experience services.

## Context

AI agents in IDEs (Cursor, Copilot) need structured access to repository memory and code understanding without re-implementing CLI logic.

## Decision

Register MCP tools for memory queries, graph traversal, intelligence workflows, and Development Experience commands (`ask`, `explain`, `plan`, `review`, `analyze_topic_impact`). Business logic stays in service layers; MCP handlers marshal inputs and format outputs only.

## Consequences

- Agents can answer questions and plan changes with the same evidence as CLI users
- Graph `analyze_impact` and topic-based `analyze_topic_impact` remain distinct tools
- MCP must not access SQLite directly
