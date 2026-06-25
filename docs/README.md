# RepoNerve Documentation

Central index for RepoNerve documentation. **Latest release:** `v1.5.1` — see `releases/versioning.md`.

---

## Start here

| Document | Purpose |
| --- | --- |
| [../README.md](../README.md) | Product overview and quick start |
| [install.md](install.md) | Install paths (npm, shell, Go, Homebrew) |
| [ai-chat-integration.md](ai-chat-integration.md) | Use RepoNerve in any IDE or LLM chat |
| [product/implementation-status.md](product/implementation-status.md) | Honest shipped vs planned snapshot |

---

## Integration

| Document | Purpose |
| --- | --- |
| [cursor-integration.md](cursor-integration.md) | Cursor skill + MCP setup |
| [copilot-chat-integration.md](copilot-chat-integration.md) | VS Code + GitHub Copilot Chat |
| [mcp/compatibility-matrix.md](mcp/compatibility-matrix.md) | IDE/client compatibility (49 MCP tools) |
| [mcp/configuration-examples.md](mcp/configuration-examples.md) | MCP config templates per client |
| [mcp/troubleshooting.md](mcp/troubleshooting.md) | MCP server errors and recovery |

---

## Product

| Document | Purpose |
| --- | --- |
| [product/README.md](product/README.md) | Product docs index |
| [product/market-positioning.md](product/market-positioning.md) | Category and differentiation |
| [product/token-economics.md](product/token-economics.md) | AI cost optimization |
| [product/greenfield-guide.md](product/greenfield-guide.md) | New projects with RepoNerve |
| [product/use-cases.md](product/use-cases.md) | Primary workflows |
| [vision/vision.md](vision/vision.md) | Product vision |

---

## Architecture

| Document | Purpose |
| --- | --- |
| [architecture/architecture-overview.md](architecture/architecture-overview.md) | System overview |
| [architecture/package-structure.md](architecture/package-structure.md) | Go package layout |
| [architecture/agent-context-contract.md](architecture/agent-context-contract.md) | MCP/CLI JSON envelope |
| [architecture/repository-ingestion.md](architecture/repository-ingestion.md) | Scan and ingestion pipeline |
| [architecture/cli-reference-v1.md](architecture/cli-reference-v1.md) | CLI command reference |

---

## Releases and planning

| Document | Purpose |
| --- | --- |
| [releases/versioning.md](releases/versioning.md) | Semver policy and release line |
| [releases/v1.5.1.md](releases/v1.5.1.md) | Latest release notes |
| [roadmap/v1.x-backlog.md](roadmap/v1.x-backlog.md) | Post-1.0 RFC-gated backlog |
| [roadmap/v1.0-iteration-plan.md](roadmap/v1.0-iteration-plan.md) | Pre-1.0 engineering checkpoints (historical) |

---

## Governance

| Document | Purpose |
| --- | --- |
| [governance/contribution-guide.md](governance/contribution-guide.md) | Contributor setup |
| [governance/rfc-process.md](governance/rfc-process.md) | RFC workflow |
| [council/software-development-council.md](council/software-development-council.md) | Multi-perspective review framework |
| [adr/](adr/) | Architecture Decision Records |

---

## RFCs (post-1.0 shipped)

| RFC | Capability | Shipped |
| --- | --- | --- |
| [RFC-001](rfc/RFC-001-bounded-agent-responses.md) | Bounded agent responses | v1.1.0 |
| [RFC-002](rfc/RFC-002-feature-intelligence-v2.md) | Feature intelligence v2 | v1.1.0 |
| [RFC-003](rfc/RFC-003-native-development-discipline.md) | Native development discipline | v1.1.0–v1.3.0 |
| [RFC-004](rfc/RFC-004-team-delivery-intelligence.md) | Team delivery intelligence | v1.3.0 |
| [RFC-005](rfc/RFC-005-configurable-document-paths.md) | Configurable document paths | v1.3.0 |
| [RFC-006](rfc/RFC-006-npm-distribution.md) | npm distribution | v1.3.2 |
| [RFC-007](rfc/RFC-007-freshness-doctor.md) | Freshness doctor | v1.4.0 |
| [RFC-008](rfc/RFC-008-scoped-monorepo-scan.md) | Scoped monorepo scan | v1.4.0 |
| [RFC-009](rfc/RFC-009-local-explore-ui.md) | Local Explore UI | v1.5.0 |

---

## Audits (v1.0 release)

Historical release-readiness audits under `audits/` — see [audits/v1.0-release-review.md](audits/v1.0-release-review.md).
