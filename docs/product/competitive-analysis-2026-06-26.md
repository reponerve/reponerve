# Competitive analysis issue drafts - 2026-06-26

Status: Issue draft pack

Scope: RepoNerve competitive landscape research against Greptile, Sourcegraph Cody, AWS Kiro, Pieces.app, Aider, Cursor, Windsurf, Continue, Cline, GitHub Copilot coding agent, CodeRabbit, Qodo, DeepWiki, Devin, Swimm, Glean, Mintlify, and adjacent code-search / agent-memory tools.

This document preserves structured GitHub issue drafts for product triage. The current automation environment can read GitHub issues and labels, but does not have a write-capable issue-creation tool. Copy each draft below into GitHub, or file them with a write-capable issue automation.

## RepoNerve product anchors

- Vision: RepoNerve reduces the cost of software understanding and preserves software knowledge for humans and AI agents.
- Category: local-first software understanding infrastructure, not autonomous coding, code search alone, or cloud-required SaaS.
- Shipped surface: deterministic scan, SQLite memory, repository intelligence, code intelligence for Go plus 19 Tree-sitter languages, repository-code linking, feature understanding, MCP, bounded context packs, `ask`, `explain`, `plan`, `impact`, `review`, `reuse-check`, `ship-check`, `pr-context`, `doctor`, `explore`, session memory, and handoff bundles.
- Current non-goals: semantic/hybrid embedding search as primary authority, user-defined workflow composition, autonomous code modification/deployment, cloud-required core product, Sourcegraph-scale cross-repo federation, and full enterprise graph explorer product.

## Competitive landscape summary

| Segment | Representative tools | Market signal | RepoNerve implication |
| --- | --- | --- | --- |
| Codebase-aware PR review | Greptile, CodeRabbit, Qodo | PR reviews now include repo-wide context, severity/risk framing, summaries, diagrams, issue context, and one-click handoff to coding agents. | Improve `pr-context` packaging without becoming an autonomous reviewer or code modifier. |
| Agentic IDEs and CLI agents | Cursor, Windsurf, Cline, Continue, Aider, GitHub Copilot coding agent, Devin | Agents expect repo-local instruction files, Plan/Act modes, MCP, issue-to-branch workflows, and agent-ready handoff context. | Export RepoNerve evidence into the formats agents already load. |
| Spec-driven workflows | AWS Kiro | Requirements, design, and task specs are becoming the control layer before AI edits code. | Add fixed, evidence-backed plan/spec exports while avoiding user-defined workflow composition. |
| Repository understanding docs | DeepWiki, Swimm, Mintlify | Users expect browsable repo wikis, diagrams, docs freshness checks, AI-ready docs, and line/evidence citations. | Generate static, local, evidence-backed guide exports and strengthen doc freshness checks. |
| Developer / enterprise memory | Pieces.app, Glean | Memory is shifting toward long-lived personal, repo, and enterprise graphs exposed through MCP. | Keep RepoNerve's moat as repo memory with evidence, and avoid OS-level or enterprise federation scope creep. |
| Semantic code retrieval | Sourcegraph Cody, Cursor, Continue, Windsurf, Bloop | Semantic retrieval is table stakes for coding tools, but can weaken traceability and local-first guarantees. | Do not chase embeddings as primary authority in this release line; compete on deterministic evidence and memory depth. |

## Suggested label taxonomy

Existing labels returned by GitHub: `bug`, `documentation`, `duplicate`, `enhancement`, `good first issue`, `help wanted`, `invalid`, `question`, `wontfix`, `repo-audit`.

The requested type and priority taxonomy needs additional labels before these drafts can be filed exactly as requested:

- Type labels: `feature`, `enhancement`, `bug`, `suggestion`
- Priority labels: `priority:P0`, `priority:P1`, `priority:P2`, `priority:P3`

## Issue drafts

### 1. Add missing product triage labels for type and priority

````markdown
Title: Add missing product triage labels for type and priority

Labels: bug, priority:P1, competitive-analysis

## Finding

The repository label set does not currently support the requested product triage taxonomy. GitHub returns `bug` and `enhancement`, but not `feature`, `suggestion`, or any `priority:*` labels.

## Why this matters

Competitive analysis and product triage should produce consistently queryable issues. Without explicit type and priority labels, roadmap findings cannot be filtered by urgency or category, and automation cannot file issues exactly as requested.

## Proposal

Add the missing labels:

- `feature`
- `suggestion`
- `priority:P0`
- `priority:P1`
- `priority:P2`
- `priority:P3`
- optional: `competitive-analysis`

Document the taxonomy in the contribution or product triage docs so future issue-filing automations use the same labels.

## Acceptance criteria

- [ ] The GitHub repository has labels for all requested types: `feature`, `enhancement`, `bug`, `suggestion`.
- [ ] The GitHub repository has priority labels `priority:P0` through `priority:P3`.
- [ ] The label meanings are documented in a repo-tracked file.
- [ ] Product triage / automation prompts can reference the taxonomy directly.

## RepoNerve alignment

This is a small process fix. It does not change architecture, storage, or product scope.
````

### 2. Ingest GitHub issue and PR metadata as local repository memory

````markdown
Title: Ingest GitHub issue and PR metadata as local repository memory

Labels: feature, priority:P1, competitive-analysis, rfc-needed

## Finding

Competitors increasingly ground agent output in issue, PR, and discussion context:

- GitHub Copilot coding agent can start from issues, create branches, use related issue/PR discussions, and update PR descriptions.
- CodeRabbit links GitHub/Jira/Linear issues to PRs, validates changes against acceptance criteria, and can create issues from review comments.
- Glean connects code with docs, PRs, tickets, and conversations in an enterprise graph.
- Sourcegraph Cody supports repository, file, symbol, web URL, and multi-repository `@` context.

RepoNerve's mission includes understanding "what changes are required" and preserving repository knowledge, but the current shipped implementation status emphasizes Git commits, ADRs, code, ownership, graph intelligence, and session memory. GitHub issue and PR bodies are not listed as a shipped ingestion source.

## Proposal

Add an optional local-first ingestion path for GitHub issues and PR metadata:

- Source types: issue, pull request, review comment, PR description, linked branch/commit references.
- Initial transport: explicit import from `gh` JSON export or local files, not mandatory SaaS.
- Store as repository memory with source IDs, timestamps, URLs, and evidence references.
- Link issues/PRs to features, decisions, code entities, commits, and ownership when deterministic evidence exists.

## Non-goals

- No cloud-required core service.
- No Sourcegraph-scale cross-repo federation.
- No permission bypass. Imported data should reflect what the user explicitly exported or granted.
- No autonomous issue fixing or deployment.

## Acceptance criteria

- [ ] RFC drafted because this introduces new persisted source/memory types.
- [ ] Import accepts a deterministic JSON format for issues and PRs.
- [ ] Imported artifacts become searchable through existing query/context engines.
- [ ] Answers cite issue/PR evidence with URL, title, number, and timestamp.
- [ ] `plan`, `review`, and `pr-context` can include linked issue requirements when available.
- [ ] Tests verify deterministic ordering and evidence preservation.

## Competitive sources

- GitHub Copilot coding agent: https://docs.github.com/copilot/concepts/agents/cloud-agent/about-cloud-agent
- CodeRabbit issue tracker integration: https://docs.coderabbit.ai/integrations/issue-trackers
- Glean code search: https://docs.glean.com/security/how-code-search-works
- Sourcegraph Cody enterprise context: https://sourcegraph.com/docs/cody/enterprise/features
````

### 3. Add evidence-backed PR risk summary and review UX sections to `pr-context`

````markdown
Title: Add evidence-backed PR risk summary and review UX sections to pr-context

Labels: enhancement, priority:P1, competitive-analysis

## Finding

Greptile, CodeRabbit, and Qodo set a high UX bar for pull request review:

- Greptile shows PR summaries, confidence / safety ratings, diagrams, severity badges, and "Fix with agent" flows.
- CodeRabbit provides summaries, walkthroughs, architectural diagrams, category badges, linked issue validation, and one-click fixes.
- Qodo applies multiple review agents with repository context, PR history, standards, and low-noise findings.

RepoNerve already has `review`, `ship-check`, and `pr-context`, and `PRContextResult` includes review, ship-check, markdown, evidence, and source services. The gap is packaging: users expect a concise PR-level artifact that clearly separates blockers, advisory risks, impacted areas, ownership, and evidence.

## Proposal

Enhance `pr-context` output with deterministic, evidence-backed sections:

- "Merge readiness" derived from `ship-check` blockers and advisories.
- "Primary risks" grouped by impacted feature/component, missing tests, ownership gaps, and changed critical paths.
- "Reviewer focus" with concrete files/symbols and related decisions.
- "Evidence trail" with links to decisions, facts, code entities, and source services.
- "Diagram outline" as Mermaid-ready text when relationships are available, without requiring image generation.

Avoid opaque model-generated scores. If a risk tier is added, derive it from named deterministic inputs.

## Acceptance criteria

- [ ] `pr-context` markdown includes blocker/advisory/risk sections with stable headings.
- [ ] Risk indicators are derived from existing evidence, not subjective LLM judgment.
- [ ] Output remains bounded by existing token/format controls.
- [ ] Tests cover deterministic ordering and evidence preservation.
- [ ] Documentation shows an example PR context comment.

## Reuse evidence

RepoNerve reuse check surfaced `internal/agent/development.PRContextResult`, `PRContextRequest`, `FormatPRCommentMarkdown`, and related Team Delivery Intelligence decisions as the right extension points.

## Competitive sources

- Greptile review anatomy: https://www.greptile.com/docs/code-review/first-pr-review
- Greptile key features: https://www.greptile.com/docs/code-review/key-features
- CodeRabbit review overview: https://docs.coderabbit.ai/guides/code-review-overview
- Qodo code review: https://docs.qodo.ai/code-review
````

### 4. Generate a static, evidence-backed repository guide export

````markdown
Title: Generate a static evidence-backed repository guide export

Labels: feature, priority:P1, competitive-analysis

## Finding

DeepWiki, Swimm, Mintlify, and Devin show strong demand for browsable repository understanding:

- DeepWiki automatically creates architecture diagrams, summaries, and Q&A grounded in repository source.
- Swimm focuses on code-coupled documentation and onboarding walkthroughs.
- Mintlify turns documentation into agent-readable outputs such as `llms.txt`, `llms-full.txt`, `skill.md`, and MCP access.
- Devin uses generated wikis and Ask Devin to improve codebase understanding.

RepoNerve has `explore`, `ask`, `explain`, graph intelligence, and context export, but a user evaluating the project still needs to run commands interactively. A static export would make RepoNerve's evidence graph easier to demo, share, review, and index by AI assistants.

## Proposal

Add a static local export command that generates a repository guide from existing evidence:

- Overview: what the repository does, architecture pillars, languages, entry points.
- Decisions: ADR/RFC summaries with links to related code.
- Features: feature -> code -> ownership -> decisions -> impact.
- Ownership and review guidance.
- Evidence index with source IDs and paths.
- Optional Mermaid diagrams generated from existing relationships.

## Non-goals

- No hosted SaaS wiki.
- No full enterprise graph explorer beyond the local `explore` scope.
- No LLM-generated claims without evidence.

## Acceptance criteria

- [ ] Static export writes Markdown (and optionally HTML) to a local directory.
- [ ] Every conclusion cites existing RepoNerve evidence.
- [ ] Export works offline from `.reponerve` memory.
- [ ] Generated output is deterministic for the same repository state.
- [ ] Docs show how to publish the export manually.

## Competitive sources

- DeepWiki / Devin docs: https://docs.devin.ai/work-with-devin/deepwiki
- Cognition DeepWiki launch: https://cognition.ai/blog/deepwiki
- Swimm enterprise documentation platform: https://swimm.io/enterprise-documentation-platform
- Mintlify AI documentation tools: https://www.mintlify.com/library/best-ai-documentation-tools
````

### 5. Extend `doctor` to flag stale documentation and ADR-to-code links

````markdown
Title: Extend doctor to flag stale documentation and ADR-to-code links

Labels: enhancement, priority:P1, competitive-analysis

## Finding

Documentation freshness is a prominent competitor theme:

- Swimm's core positioning is keeping docs synchronized with code and blocking or alerting on stale documentation in CI.
- Mintlify markets self-updating docs and automatic suggestions when code changes.
- DeepWiki and Devin emphasize generated docs that stay useful as repositories change.

RepoNerve ships `doctor`, configurable document paths, repository-code linking, and evidence-backed outputs. A natural gap is detecting when indexed docs, ADRs, or feature explanations reference code that changed significantly after the evidence was created.

## Proposal

Extend `reponerve doctor` with documentation freshness checks:

- ADR/document references code entity that no longer exists.
- ADR/document linked code changed after the source document was last updated.
- Feature explanation has thin or missing code evidence.
- Configured document paths are empty, missing, or no longer scanned.

## Acceptance criteria

- [ ] `doctor` reports stale or broken doc-to-code links with evidence.
- [ ] Output distinguishes blockers from advisories.
- [ ] Checks are deterministic and do not require network access.
- [ ] CI-friendly JSON output includes path, source type, related entity, and recommendation.
- [ ] Tests cover deleted symbols, modified linked files, and missing configured document paths.

## Competitive sources

- Swimm documentation platform: https://swimm.io/enterprise-documentation-platform
- Swimm `/ask` context: https://swimm.io/blog/meetask-swimm-your-teams-contextual-ai-coding-assistant
- Mintlify developer documentation: https://www.mintlify.com/use-cases/developer-documentation
````

### 6. Add agent export profiles for Cursor, Cline, Kiro, Windsurf, Copilot, and Aider

````markdown
Title: Add agent export profiles for Cursor, Cline, Kiro, Windsurf, Copilot, and Aider

Labels: feature, priority:P1, competitive-analysis

## Finding

AI coding tools increasingly load repo-local instructions and context files:

- Cursor uses project rules, modes, subagents, MCP, and cloud agents.
- Cline uses `.clinerules`, Plan/Act, MCP, plugins, and CLI.
- Kiro uses `.kiro/steering`, specs, hooks, powers, and MCP.
- Windsurf uses rules, memories, modes, and Fast Context.
- GitHub Copilot uses repository instructions and issue/PR context for cloud agents.
- Aider uses `.aider.conf.yml`, repo maps, and Git-native terminal flows.

RepoNerve already generates Cursor integration on init and exposes MCP/context export. The gap is broader ecosystem packaging: users should be able to export the same evidence-backed guidance into the files that their chosen agent already reads.

## Proposal

Add fixed export profiles, for example:

```bash
reponerve context export --profile cursor
reponerve context export --profile cline
reponerve context export --profile kiro
reponerve context export --profile windsurf
reponerve context export --profile copilot
reponerve context export --profile aider
```

Each profile should emit a bounded, evidence-backed file or snippet with:

- Repo purpose and architecture constraints.
- Native Development Discipline workflow.
- Must-use RepoNerve commands before editing.
- Out-of-scope warnings.
- MCP setup reminder where applicable.

## Non-goals

- No user-defined workflow composition.
- No automatic code edits.
- No secrets or credentials in generated files.

## Acceptance criteria

- [ ] At least three high-demand profiles are implemented first.
- [ ] Each generated file is deterministic and safe to commit.
- [ ] Docs explain where each target tool loads the file.
- [ ] Existing Cursor init behavior remains backwards compatible.
- [ ] Tests verify no secrets, no co-author footers, and stable output.

## Competitive sources

- Cursor Agent docs: https://cursor.com/help/ai-features/agent
- Cline MCP docs: https://docs.cline.bot/mcp/mcp-overview
- Kiro repository docs: https://github.com/kirodotdev/Kiro
- Windsurf Cascade modes: https://docs.windsurf.com/windsurf/cascade/modes
- Aider Git integration: https://aider.chat/docs/git.html
````

### 7. Add plan-to-spec export for committed requirements, design, and tasks

````markdown
Title: Add plan-to-spec export for committed requirements, design, and tasks

Labels: enhancement, priority:P2, competitive-analysis

## Finding

AWS Kiro's strongest product signal is spec-driven development: prompts become requirements, design, and task files before implementation begins. This pattern gives teams reviewable control over AI work.

RepoNerve already has `plan`, `reuse-check`, `impact`, and review/ship readiness. The gap is durable task artifacts. A user can receive a plan, but there is no first-class fixed template that writes an evidence-backed plan to a repository-local spec file.

## Proposal

Add a fixed export mode:

```bash
reponerve plan "Add OAuth login" --write
```

Output a deterministic Markdown bundle, such as:

- `docs/plans/<slug>/requirements.md`
- `docs/plans/<slug>/design.md`
- `docs/plans/<slug>/tasks.md`
- `docs/plans/<slug>/evidence.md`

The files should include starting points, reuse candidates, relevant decisions, test expectations, non-goals, and recommended next tools.

## Non-goals

- No open-ended custom workflow engine.
- No automatic implementation.
- No Kiro-specific dependency.

## Acceptance criteria

- [ ] `plan --write` emits a fixed, documented template.
- [ ] Generated files cite RepoNerve evidence and source services.
- [ ] Output path is configurable only within repo-safe paths.
- [ ] Tests verify deterministic output and path safety.
- [ ] Documentation explains how agents should consume the spec.

## Competitive sources

- Kiro specs and steering: https://github.com/kirodotdev/Kiro
- InfoWorld Kiro analysis: https://www.infoworld.com/article/4023980/from-prompts-to-specs-awss-kiro-signals-the-next-phase-of-ai-coding-tools.html
````

### 8. Create agent-ready handoff bundles for PR review findings and external coding agents

````markdown
Title: Create agent-ready handoff bundles for PR review findings and external coding agents

Labels: feature, priority:P2, competitive-analysis

## Finding

Competitors are making "review finding -> coding agent" a one-click path:

- Greptile offers Fix with Agent / Fix All flows for Claude Code, Codex, Cursor, Devin, and others.
- CodeRabbit offers one-click and AI fixes.
- GitHub Copilot cloud agent can continue work from issues or chat into branches and PRs.
- Devin can pick up review feedback and CI results.

RepoNerve must not autonomously modify code, but it can package evidence so external agents make safer changes.

## Proposal

Extend session memory / context export with handoff bundle variants:

- `--target cursor`
- `--target aider`
- `--target copilot`
- `--target cline`
- `--target generic`

Bundles should include:

- The issue or PR finding.
- Changed files and impacted entities.
- Related decisions and ownership evidence.
- Recommended tests and `ship-check` blockers.
- Explicit non-goals and out-of-scope constraints.

## Non-goals

- No automatic commit, fix, deploy, or PR merge.
- No vendor-specific API calls in the first version.
- No hidden prompt injection from untrusted issue text.

## Acceptance criteria

- [ ] Handoff bundle schema is documented.
- [ ] Bundle output is available as Markdown and JSON.
- [ ] Each bundle includes evidence and source services.
- [ ] Unsafe external text is clearly delimited from RepoNerve-generated instructions.
- [ ] Tests cover deterministic ordering and injection-safe formatting.

## Competitive sources

- Greptile key features: https://www.greptile.com/docs/code-review/key-features
- CodeRabbit pull request review: https://docs.coderabbit.ai/overview/pull-request-review
- GitHub Copilot cloud agent: https://code.visualstudio.com/docs/copilot/copilot-cloud-agent
- Devin SDLC integration: https://docs.devin.ai/essential-guidelines/sdlc-integration
````

### 9. Publish local-first security and privacy posture for memory DB and MCP

````markdown
Title: Publish local-first security and privacy posture for memory DB and MCP

Labels: suggestion, priority:P1, competitive-analysis, documentation

## Finding

Security and privacy positioning is now part of developer-tool selection:

- Pieces emphasizes private-by-design, local-by-default, on-device workflow memory.
- Sourcegraph and Glean emphasize enterprise access controls, model/provider controls, and permission-aware context.
- Cursor/Cline/Kiro/Windsurf all expose MCP and tool execution surfaces that require guardrails.

RepoNerve has a strong local-first story, but evaluators need a concise security posture that explains what is stored locally, what MCP exposes, how secrets are handled, and what is intentionally not sent to cloud services.

## Proposal

Create a security/privacy document covering:

- What `.reponerve/memory.db` stores.
- What `scan` reads and ignores.
- What MCP tools can expose.
- How to keep the product air-gapped/local-first.
- Secret-handling guarantees and limitations.
- Recommended Git ignore / CI practices.
- Threat model for untrusted repository text and agent prompt injection.

## Acceptance criteria

- [ ] Documentation explains local storage and data flow.
- [ ] Documentation states that core product does not require mandatory SaaS.
- [ ] MCP exposure and operator responsibilities are documented.
- [ ] Secret handling guidance is explicit.
- [ ] Links are added from install, MCP, and market-positioning docs.

## Competitive sources

- Pieces Long-Term Memory: https://docs.pieces.app/products/desktop/long-term-memory
- Sourcegraph Cody docs: https://sourcegraph.com/docs/cody
- Glean MCP server: https://docs.glean.com/user-guide/mcp/usage
- Cline MCP docs: https://docs.cline.bot/mcp/mcp-overview
````

### 10. Add a guided first-run tutorial and demo path

````markdown
Title: Add a guided first-run tutorial and demo path

Labels: enhancement, priority:P2, competitive-analysis

## Finding

AI developer tools increasingly optimize for immediate "aha" moments:

- DeepWiki works by replacing `github.com` with `deepwiki.com`.
- Kiro has first-project guides for specs, steering, hooks, and MCP.
- Cursor and Windsurf expose clear Ask/Plan/Agent mode workflows.
- Aider's terminal workflow centers on Git, `/diff`, `/test`, `/undo`, and repo maps.

RepoNerve is powerful but command-rich. New users need a guided path that proves the core value quickly: scan once, ask/explain/plan with evidence, then hand context to an agent.

## Proposal

Add a guided first-run command or tutorial:

```bash
reponerve demo
reponerve tutorial
```

It should:

- Verify install and memory status.
- Run or explain `scan`.
- Ask 2-3 sample questions.
- Show evidence and source services.
- Demonstrate MCP setup / context export.
- Suggest next commands based on persona.

## Acceptance criteria

- [ ] Tutorial works in an existing repository without modifying code.
- [ ] Output is bounded and friendly for terminal users.
- [ ] It avoids network access by default.
- [ ] It points to `ask`, `explain`, `plan`, `reuse-check`, `ship-check`, and MCP setup.
- [ ] Docs include a "60-second demo" script.

## Competitive sources

- DeepWiki launch: https://cognition.ai/blog/deepwiki
- Kiro repository docs: https://github.com/kirodotdev/Kiro
- Cursor Agent docs: https://cursor.com/help/ai-features/agent
- Aider Git integration: https://aider.chat/docs/git.html
````

### 11. Add MCP and AI-readable documentation discovery metadata

````markdown
Title: Add MCP and AI-readable documentation discovery metadata

Labels: enhancement, priority:P2, competitive-analysis

## Finding

MCP is now table stakes across Cursor, Cline, Continue, Kiro, Glean, Pieces, Mintlify, Sourcegraph, and GitHub Copilot. Mintlify also highlights AI-readable docs outputs such as `llms.txt`, `llms-full.txt`, `skill.md`, and hosted MCP for documentation.

RepoNerve already ships MCP and Cursor integration, but growth depends on being easily discovered and consumed by agent ecosystems.

## Proposal

Add repo-level discovery artifacts:

- `llms.txt` / `llms-full.txt` for public docs.
- `skill.md` or equivalent agent-consumable quickstart.
- MCP capability manifest with tool categories and local-first posture.
- Registry-ready metadata for MCP directories where appropriate.

## Acceptance criteria

- [ ] AI-readable docs files are generated or maintained in repo.
- [ ] MCP tool categories and setup are summarized in a short manifest.
- [ ] The docs link to compatibility examples for Cursor, Copilot, Cline, Continue, JetBrains, and Windsurf.
- [ ] No secrets or environment-specific paths are included.
- [ ] CI or docs checks catch stale metadata where practical.

## Competitive sources

- Mintlify AI docs: https://www.mintlify.com/library/best-ai-documentation-tools
- Continue MCP docs: https://docs.continue.dev/customize/deep-dives/mcp
- Cline MCP docs: https://docs.cline.bot/mcp/mcp-overview
- Glean MCP server: https://docs.glean.com/user-guide/mcp/usage
````

### 12. Publish token-efficiency and onboarding benchmarks against representative tools

````markdown
Title: Publish token-efficiency and onboarding benchmarks against representative tools

Labels: suggestion, priority:P2, competitive-analysis, documentation

## Finding

The market is crowded with "codebase context" claims. RepoNerve's differentiator is not generic retrieval; it is deterministic, evidence-backed software understanding with lower exploration cost. That claim should be demonstrated with reproducible benchmarks.

Useful comparisons:

- RepoNerve `ask` / `plan` / `explain` versus raw agent grep/read exploration.
- RepoNerve context export versus prompt-pack tools.
- RepoNerve PR context versus generic AI review summaries.
- Day-one onboarding path versus DeepWiki-style generated docs and Sourcegraph/Glean/Cody-style search.

## Proposal

Publish a small benchmark suite and narrative report:

- Select 2-3 public repositories.
- Define tasks: "find owner", "explain feature", "plan a change", "review PR".
- Measure tool calls, output size, evidence count, and repeatability.
- Include limitations and non-goals.

## Acceptance criteria

- [ ] Benchmark methodology is documented and reproducible.
- [ ] Results include at least one public repository.
- [ ] Metrics include tokens/output size, evidence coverage, and deterministic repeatability.
- [ ] Report avoids unverifiable competitor performance claims.
- [ ] Market-positioning docs link to the benchmark.

## Competitive sources

- RepoNerve token economics: `docs/product/token-economics.md`
- Sourcegraph Cody context: https://sourcegraph.com/blog/how-cody-understands-your-codebase
- Windsurf Fast Context: https://docs.windsurf.com/context-awareness/fast-context
- Bloop archived code search: https://github.com/BloopAI/bloop
````

## Rejected or deferred competitive ideas

| Idea | Verdict | Reason |
| --- | --- | --- |
| Embedding / vector search as primary authority | Reject for current release line | Explicitly out of scope in `docs/roadmap/v1.x-backlog.md`; would require RFC if revisited. |
| Cross-repo enterprise federation | Reject for current release line | Explicitly out of scope; Sourcegraph/Glean own this segment. |
| Autonomous code modification and deployment | Reject | Conflicts with RepoNerve's product boundary as understanding infrastructure. |
| User-defined workflow composition / arbitrary hooks | Reject for current release line | Explicitly out of scope; fixed export templates are acceptable. |
| OS-level personal memory capture | Defer / reject | Pieces owns this adjacent category; RepoNerve should stay repository-scoped. |
| Full hosted graph explorer | Reject for current release line | Local `explore` is shipped; full enterprise explorer remains out of scope. |

## Source index

- RepoNerve vision: `docs/vision/vision.md`
- RepoNerve market positioning: `docs/product/market-positioning.md`
- RepoNerve implementation status: `docs/product/implementation-status.md`
- RepoNerve out-of-scope list: `docs/roadmap/v1.x-backlog.md`
- Greptile key features: https://www.greptile.com/docs/code-review/key-features
- Greptile graph context: https://www.greptile.com/docs/how-greptile-works/graph-based-codebase-context
- Sourcegraph Cody docs: https://sourcegraph.com/docs/cody
- Sourcegraph Cody codebase context: https://sourcegraph.com/blog/how-cody-understands-your-codebase
- Kiro repository: https://github.com/kirodotdev/Kiro
- Pieces Long-Term Memory: https://docs.pieces.app/products/desktop/long-term-memory
- Aider Git integration: https://aider.chat/docs/git.html
- Cursor Agent docs: https://cursor.com/help/ai-features/agent
- Continue MCP docs: https://docs.continue.dev/customize/deep-dives/mcp
- Windsurf Cascade modes: https://docs.windsurf.com/windsurf/cascade/modes
- Windsurf Fast Context: https://docs.windsurf.com/context-awareness/fast-context
- Cline MCP docs: https://docs.cline.bot/mcp/mcp-overview
- GitHub Copilot cloud agent: https://docs.github.com/copilot/concepts/agents/cloud-agent/about-cloud-agent
- CodeRabbit PR review: https://docs.coderabbit.ai/overview/pull-request-review
- Qodo code review: https://docs.qodo.ai/code-review
- DeepWiki docs: https://docs.devin.ai/work-with-devin/deepwiki
- Devin SDLC integration: https://docs.devin.ai/essential-guidelines/sdlc-integration
- Swimm enterprise documentation: https://swimm.io/enterprise-documentation-platform
- Glean code search: https://docs.glean.com/security/how-code-search-works
- Mintlify AI documentation tools: https://www.mintlify.com/library/best-ai-documentation-tools
- Bloop archived code search: https://github.com/BloopAI/bloop
