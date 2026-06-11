# ISSUE-062 — Multi-Language Code Intelligence

Status: Planned

Milestone: v0.15.0-alpha

Depends On: ISSUE-057 (Go code intelligence architecture)

Part of: v1.0.0 (single product release)

---

# Objective

Extend Code Intelligence beyond Go using Tree-sitter parsers while preserving deterministic, evidence-backed authority boundaries.

---

# Deliverables

| Language | Parser | Minimum entities |
| --- | --- | --- |
| TypeScript | Tree-sitter | files, functions, classes, imports |
| Python | Tree-sitter | files, functions, classes, imports |
| Rust | Tree-sitter | files, functions, structs, imports |

* Incremental per-file indexing
* Same storage and reader patterns as Go (`internal/code/`)
* Repository-code linking works across languages

---

# Acceptance Criteria

* `reponerve scan` indexes Go + at least TS, Python, Rust in a mixed repo
* `explain-file` works for indexed non-Go files
* Symbol resolution deterministic and tested
* No LLM required for parsing

---

# Non-Goals

* Full parity with Go endpoint detection on day one
* Semantic / embedding search

---

# Tag

Engineering checkpoint: `v0.15.0-alpha`

Final engineering checkpoint before v1.0.0 release review.
