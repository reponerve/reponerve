# Product triage reference

## Vision anchor (must appear in every triage)

> RepoNerve reduces the cost of **software understanding** through **knowledge preservation** — local, evidence-backed memory that survives beyond individual contributors and works for humans and AI agents.

Users buy **understanding, speed, confidence, and reduced exploration** — not raw intelligence or code graphs alone.

Guiding principle: **Understanding first. Evidence second. AI third.**

---

## Align signals (support ALIGN verdict)

- Improves `ask`, `explain*`, `plan`, `review`, `reuse-check`, `ship-check`, `onboard`, `doctor`, `explore`, or MCP tools
- Strengthens deterministic scan, memory, repository–code linking, ownership, or feature understanding
- Reduces token waste / re-exploration (RFC-001 spirit)
- Ships **local-first**; no mandatory cloud
- Preserves evidence in outputs (no hallucinated rankings)
- Fixes real bugs or docs that block adoption
- npm / install / IDE integration friction (RFC-006 class)

---

## Reject signals (support REJECT verdict)

From `docs/roadmap/v1.x-backlog.md`:

| Out of scope | Why |
| --- | --- |
| Semantic / embedding search as primary authority | Conflicts with deterministic evidence model |
| Hybrid search where vectors override evidence | Same |
| User-defined workflow composition beyond fixed set | Scope creep |
| Autonomous code modification (write/commit/deploy) | Product boundary |
| Cloud-required core product | Local-first mission |
| Cross-repo enterprise federation | Not v1.x |
| Full-product graph explorer beyond local `explore` | Enterprise scope |

Also **reject** when:

- Feature is a thin wrapper around "let the LLM grep the repo"
- Duplicates Cursor/Copilot without RepoNerve's **why / who / what breaks** layer
- Adds SaaS dependency to the critical path without RFC

---

## RFC required (ALIGN — RFC)

From `docs/governance/rfc-process.md` — significant architecture, new memory types, search replacement, major MCP contract changes, new ingestion pipelines.

Before tagging a release: row in `docs/releases/versioning.md`.

---

## Shipped — do not re-propose (check DUPLICATE)

See `docs/product/implementation-status.md`. Latest line: **v1.5.1**.

Includes: 49 MCP tools, 20 languages, native discipline, reuse/ship-check, doctor, scoped scan, npm, explore UI, PR context, team delivery intelligence.

---

## Priority tie-breakers

1. Unblocks **universal understanding** (day-one dev, pasted ticket, weak models)
2. Closes evidence gaps in Development Experience
3. Maintainer burden / support cost reduction (doctor, install, docs)
4. Community **bug** with reproduction
5. Nice-to-have UI polish — **DEFER** unless tied to adoption metric

---

## Labels (GitHub)

| Label | Default triage bias |
| --- | --- |
| `bug` | ALIGN if reproducible; else `question` |
| `enhancement` | Requires vision rubric; often RFC or DEFER |
| `documentation` | ALIGN if accuracy/adoption; quick wins OK |
| `question` | Answer + close; no roadmap slot |
| `good first issue` | ALIGN if still vision-aligned after rubric |
