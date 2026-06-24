# RFC-001: Bounded Agent Responses

Status: Accepted  
Date: 2026-06-24

## Problem

On large repositories, Development Experience tools return unbounded `structured` JSON. Agents skip the payload and fall back to grep, defeating RepoNerve's purpose.

## Decision

1. Default `token_budget` of **1500** when unset (CLI + MCP).
2. Cap list fields before MCP/JSON emit: `related` (15), `evidence` (20), plan `starting_points` (8), `impacted_areas` (15).
3. Set `agent.truncated` and `agent.prefer_narrow_tools` when caps apply.
4. Downgrade guidance: use `explain_function` / `explain_file` instead of grep.

## Non-goals

- Semantic summarization of capped fields (future).
- Changing scan/index behavior.
