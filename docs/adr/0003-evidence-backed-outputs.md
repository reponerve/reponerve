# 3. Evidence-backed deterministic outputs

## Status

Accepted

Every RepoNerve conclusion must trace to repository evidence. Outputs are deterministic: the same repository state produces the same results. Subjective rankings and AI-generated ownership are not valid conclusions.

## Context

Agents and humans need to trust RepoNerve answers for planning, review, and impact analysis. Hallucinated or heuristic-only answers undermine the product mission of software understanding.

## Decision

Follow Understanding First, Evidence Second, AI Third. Use AI only for intent, decision, and tradeoff extraction from sources — never for repository scanning or AST parsing.

## Consequences

- All Development Experience responses include traceable evidence
- Graph edges and ownership recommendations expose supporting evidence
- Ordering of query results is stable and testable
