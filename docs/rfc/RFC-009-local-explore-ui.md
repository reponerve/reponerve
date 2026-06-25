# RFC-009: Local Explore UI

Status: Accepted  
Date: 2026-06-24

Related:

* `docs/rfc/RFC-006-npm-distribution.md`
* `internal/cli/explore/`
* `internal/ui/explore/`

---

## Problem

`reponerve explore` exports raw JSON in HTML. Humans and demos need a **small local inspector** for the knowledge graph — GitNexus-like browse without a separate SPA product or cloud host.

## Decision

Ship **Local Explore UI** in **v1.5.0**:

| Surface | Behavior |
| --- | --- |
| `reponerve explore --serve` | HTTP on `127.0.0.1` (default port `8765`) |
| `reponerve explore -o file.html` | Static export (same templ layout, no server) |
| Stack | **templ** + **htmx** + **Alpine.js**; Cytoscape.js for capped subgraph |
| Cap | Max **200** nodes in UI; full counts shown in header |

### Pages / routes

| Route | Purpose |
| --- | --- |
| `GET /` | Graph overview, filters, cytoscape panel, node table |
| `GET /nodes` | htmx partial — filter by `type`, search `q` |
| `GET /nodes/{id}` | htmx partial — evidence panel, edges, MCP hints |

### Evidence panel

Node detail shows: type, entity ID, stored vs derived edges, `evidence_json` snippets, suggested tools (`explain_decision`, `explain_file`, …).

### Non-goals

- Public bind / auth / multi-user
- Edit or write actions
- Full graph layout at unlimited scale
- Replacement for MCP (agents still use MCP)

---

## Architecture

```text
CLI explore --serve
    → internal/ui/explore.Loader (graph snapshot)
    → internal/ui/explore.Server (templ handlers)
    → query/storage readers (evidence only, no business logic in templates)
```

---

## Success criteria

| Verify |
| --- |
| `reponerve explore --serve` opens UI on localhost |
| Node click loads evidence partial via htmx |
| Static `-o` export renders without server |
| Binds `127.0.0.1` only |

---

## v1.5.0 bundle

| Item | Surface |
| --- | --- |
| Local Explore UI | `explore --serve`, improved `-o` export |
