# 2. Code Intelligence via deterministic AST indexing

## Status

Accepted

Code Intelligence is extracted by parsing Go source with the standard AST. RepoNerve indexes packages, files, symbols, call relationships, and dependencies into SQLite. AI is not used for scanning or parsing.

## Context

Development Experience commands (`explain`, `plan`, `impact`) require authoritative code structure. Re-scanning the repository on every query would be slow and non-deterministic.

## Decision

Extract code entities once during `scan`, persist them in the code intelligence store, and consume them through the Code Service and Development Experience layer.

## Consequences

- Deterministic, reproducible code understanding
- Repository-code linking can connect ADRs and decisions to symbols
- Language support starts with Go; other languages are out of v1 scope
