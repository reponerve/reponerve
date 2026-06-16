# 1. Local-first SQLite storage

## Status

Accepted

RepoNerve stores repository memory in a local SQLite database. No cloud services or external infrastructure are required for core operation. Each workspace owns its database file on disk.

## Context

Software understanding must remain available offline and under the contributor's control. Repository intelligence is derived from local sources (git history, ADRs, indexed code) and persisted locally.

## Decision

Use SQLite as the single persistence layer for repository memory, code intelligence indexes, and search (including FTS5).

## Consequences

- Fast local queries without network latency
- Simple deployment as a CLI tool
- Backup and portability via copying the database file
- Scaling limits are acceptable for v1 local-first scope
