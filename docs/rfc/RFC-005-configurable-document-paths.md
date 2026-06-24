# RFC-005: Configurable Document Paths

Status: Accepted  
Date: 2026-06-24

Related:

* `docs/releases/v1.3.0.md`
* `docs/rfc/RFC-003-native-development-discipline.md`
* `internal/scanner/adr/scanner.go`

---

## Problem

ADR and architecture markdown ingestion only scanned five fixed directories (`docs/adr`, `docs/adrs`, `adr`, `adrs`, `docs/architecture`). Many repositories store decisions elsewhere (`docs/decisions/`, `docs/rfc/`, team-specific folders). Without configurable paths:

- `scan` misses decision memory
- `discipline-policy.json` omits ADR expectations
- `review` / `ship_check` lack linked decisions

Repos **without any ADR folders** still work via git events, code intelligence, and ownership — but teams with non-standard doc layout get no “why” layer.

## Decision

Ship **Configurable Document Paths** in **v1.3.0** — extend scan and discipline policy from `.reponerve/config.yaml` without cloud services.

### Config schema

```yaml
repository:
  path: .

ingestion:
  document_paths:
    - path: docs/decisions
      kind: adr
    - path: docs/design
      kind: architecture_doc
```

| Field | Values | Default when omitted |
| --- | --- | --- |
| `path` | Repository-relative directory | — |
| `kind` | `adr`, `architecture_doc` | `adr` |

Configured paths are **merged** with built-in defaults (not replaced). Order: user paths first, then defaults; duplicates deduped.

### Expanded defaults (v1.3)

Added without config:

- `docs/decisions`
- `docs/rfc`

So RepoNerve ingests this repository’s RFCs out of the box.

### Consumers

| Component | Behavior |
| --- | --- |
| ADR scanner | `adr.NewScanner(cfg.Ingestion.DocumentPaths...)` |
| `discipline.Derive` | `adr.PrimaryADRDirectory` over resolved paths |
| Decision extractor | Unchanged — processes ingested `adr` sources |

---

## Non-goals

- Non-markdown formats (Notion export HTML, PDF)
- External doc URLs without git files
- LLM classification of arbitrary `docs/**`
- Replacing git commit memory when no docs exist

---

## Success criteria

| Verify |
| --- |
| Custom path in `config.yaml` is scanned on `reponerve scan` |
| `docs/rfc/*.md` indexed on RepoNerve repo without extra config |
| `discipline-policy.json` sets `adr_directory` from first matching ADR path |
| Existing ADR integration tests still pass |

---

## v1.3.0 bundle (updated)

| RFC | Capability |
| --- | --- |
| RFC-003 Phase D | `discipline-policy.json` on scan |
| RFC-004 A–C | Evidence Review, PR Context, workflow template |
| RFC-005 | Configurable + expanded document paths |
