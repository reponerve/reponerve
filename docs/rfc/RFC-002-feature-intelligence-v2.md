# RFC-002: Feature Intelligence v2

Status: Accepted (baseline)  
Date: 2026-06-24

## Problem

Feature Understanding is orchestration-only. `explain "authentication"` resolves to unrelated symbols. No `list_features` or feature entity.

## Decision (Phase 1 baseline)

Derive features deterministically from:

- Expertise domains (`DomainKeywords`)
- `FEATURE_INTRODUCED` events (normalized titles)
- ADR/decision titles matching domain keywords

Expose:

- `list_features` / `reponerve list-features`
- `explain_feature` / `reponerve explain-feature`
- `explain [topic]` routes through feature match when `ShouldAutoExplain` passes (exact feature name or single-token alias); multi-word symbol topics stay on code/topic resolution

## Future (Phase 1b)

- Persisted `features` table (migration v10)
- Feature-code linking at scan time
- `feature_impact` MCP tool
