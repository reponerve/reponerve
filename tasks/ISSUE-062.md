# ISSUE-062 — Multi-Language Code Intelligence

Status: Complete — v0.15.0-alpha

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
| TypeScript | Tree-sitter (`gotreesitter`) | `.ts`, `.tsx` — files, functions, classes, interfaces, imports |
| JavaScript | Tree-sitter (`gotreesitter`) | `.js`, `.jsx`, `.mjs`, `.cjs` — files, functions, classes, imports |
| Python | Tree-sitter (`gotreesitter`) | `.py` — files, functions, classes, imports |
| Rust | Tree-sitter (`gotreesitter`) | `.rs` — files, functions, structs, traits, imports |
| Java | Tree-sitter (`gotreesitter`) | `.java` — files, classes, interfaces, methods, imports |
| C# | Tree-sitter (`gotreesitter`) | `.cs` — files, classes, interfaces, methods, usings |
| Ruby | Tree-sitter (`gotreesitter`) | `.rb` — files, classes, modules, methods, requires |
| Kotlin | Tree-sitter (`gotreesitter`) | `.kt`, `.kts` — files, classes, interfaces, functions, imports |
| Swift | Tree-sitter (`gotreesitter`) | `.swift` — files, classes, protocols, functions, imports |
| PHP | Tree-sitter (`gotreesitter`) | `.php` — files, classes, interfaces, methods, imports |
| C | Tree-sitter (`gotreesitter`) | `.c`, `.h` — files, structs, functions, includes |
| C++ | Tree-sitter (`gotreesitter`) | `.cpp`, `.cc`, `.hpp`, … — files, classes, methods, includes |
| Scala | Tree-sitter (`gotreesitter`) | `.scala` — files, classes, traits, objects, methods, imports |
| Lua | Tree-sitter (`gotreesitter`) | `.lua` — tables/modules, functions, methods, requires |
| Bash | Tree-sitter (`gotreesitter`) | `.sh`, `.bash` — functions, source imports |
| SQL | Tree-sitter (`gotreesitter`) | `.sql` — tables, views |
| Dart | Tree-sitter (`gotreesitter`) | `.dart` — classes, methods, functions, imports |
| Elixir | Tree-sitter (`gotreesitter`) | `.ex`, `.exs` — modules, functions, imports |
| Zig | Tree-sitter (`gotreesitter`) | `.zig` — structs, functions, `@import` |

* Incremental per-file indexing — Done
* Same storage and reader patterns as Go (`internal/code/`) — Done
* Repository-code linking works across languages — Done

---

# Acceptance Criteria

* `reponerve scan` indexes Go + at least TS, Python, Rust in a mixed repo — Done
* `explain-file` works for indexed non-Go files — Done
* Symbol resolution deterministic and tested — Done
* No LLM required for parsing — Done

---

# Implementation

```text
internal/code/lang/          # Tree-sitter extractors (TS, Python, Rust)
internal/code/indexer/       # Multi-language file discovery + indexing
```

Parser: pure-Go `github.com/odvcencio/gotreesitter` (no CGO).

---

# Non-Goals

* Full parity with Go endpoint detection on day one
* Semantic / embedding search

---

# Tag

Engineering checkpoint: `v0.15.0-alpha`

Final engineering checkpoint before v1.0.0 release review.
