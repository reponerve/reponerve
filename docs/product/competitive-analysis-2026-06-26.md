# RepoNerve competitive analysis - 2026-06-26

Status: Issue-ready backlog draft

Research date: 2026-06-26

Related:

- `docs/vision/vision.md`
- `docs/product/market-positioning.md`
- `docs/product/implementation-status.md`
- `docs/roadmap/v1.x-backlog.md`
- `docs/governance/rfc-process.md`

---

## Vision anchor

RepoNerve reduces the cost of software understanding through knowledge preservation: local, evidence-backed memory that survives beyond individual contributors and works for humans and AI agents.

RepoNerve should not compete by becoming a generic coding agent, cloud SaaS core, semantic-search product, or Sourcegraph-scale federation layer. Competitive work should strengthen local-first repository memory, evidence-backed context, Development Experience, install trust, and growth loops.

---

## Constraints from current product scope

Already shipped in v1.5.1:

- 49 MCP tools
- Code intelligence for Go plus 19 Tree-sitter languages
- Repository-code linking
- Feature understanding
- `ask`, `explain`, `plan`, `impact`, `review`, `reuse-check`, `ship-check`, `pr-context`
- Session memory with `remember` / `forget`
- Native Development Discipline installed on `init`
- npm distribution, Homebrew formula, doctor, scoped monorepo scan
- Local Explore UI

Explicitly out of scope without a new RFC:

- Semantic or hybrid embedding search as primary authority
- User-defined workflow composition beyond fixed templates
- Autonomous code modification or deployment
- Cloud-required core product
- Cross-repo enterprise federation
- Full-product graph explorer beyond capped local `reponerve explore`

---

## Competitor map

| Category | Competitors | Competitive signal |
| --- | --- | --- |
| AI PR review | Greptile, CodeRabbit, Qodo | Full-codebase PR review, learned standards, fix handoff, pre-merge checks, multi-agent review |
| Enterprise code intelligence | Sourcegraph Cody, Tabnine Enterprise Context Engine, Augment-style context engines | Code graph, multi-repo and enterprise governance, analytics, compliance, deployment controls |
| Agentic IDEs and CLIs | Cursor, Windsurf/Devin Desktop, AWS Kiro, Claude Code, Cline, Aider, Continue | Rules, skills, hooks, memories, subagents, spec workflows, repo maps, autonomous edit loops |
| Local memory | Pieces.app, Cline Memory Bank, MemNexus-style MCP memory | Personal or workflow memory, OS-level capture, searchable timelines, cross-session continuity |
| Code graph + MCP | GitNexus, Cortex, Codebase-Memory, CodeGraphContext | Local graph indexes, MCP tools, setup automation, benchmark claims, broad language coverage |
| Context packing | Repomix, code2prompt | One-command prompt packs, token counts, simple install, weak but understandable UX |
| Repository docs and onboarding | DeepWiki / Devin Wiki | Auto-generated public repo docs, diagrams, chat, steerable wiki config |

---

## Ecosystem trends

1. **MCP is table stakes.** Cursor, Cody, Kiro, Pieces, Continue, Claude Code, Cline, DeepWiki, GitNexus, Cortex, and Codebase-Memory all position MCP or MCP-like context access as a primary integration layer.
2. **Context is becoming a product layer.** Competitors increasingly market "context engines", "knowledge graphs", "repo maps", and "long-term memory", not only chat or autocomplete.
3. **Agent harness features matter.** Hooks, rules, skills, subagents, checkpoints, and spec files are becoming standard UX primitives.
4. **PR review is a crowded wedge.** Greptile, CodeRabbit, and Qodo anchor adoption in pull requests with comments, severity, standards, and fix workflows.
5. **Proof beats positioning.** Codebase-Memory and similar tools publish token and quality benchmarks. RepoNerve's token-economics narrative needs reproducible evidence.
6. **Local-first is still differentiating.** Pieces, GitNexus, Cortex, Aider, and Codebase-Memory show demand for local or BYOK workflows, while Sourcegraph, Qodo, Greptile, and Tabnine skew enterprise.
7. **Visual onboarding is expected.** DeepWiki, GitNexus, Pieces, and local graph UIs make understanding tangible; CLI-only claims are harder to evaluate.
8. **The durable gap is "why/who/what breaks".** Most competitors are strong at code structure, review, or chat memory; fewer combine repository memory, ownership, decisions, impact, and evidence requirements.

---

## Source notes

- Greptile: graph-based repo context, PR comments, team learning, IDE/MCP fix handoff, self-hosted/air-gapped options, TREX execution layer.
  - https://www.greptile.com/docs/introduction
  - https://www.greptile.com/docs/code-review/key-features
  - https://www.greptile.com/docs/how-greptile-works/graph-based-codebase-context
  - https://www.greptile.com/blog/trex-code-execution
- Sourcegraph Cody: Search API/code graph context, IDE chat/completions, Enterprise context filters, multi-repo `@` context, analytics and governance.
  - https://sourcegraph.com/docs/cody
  - https://sourcegraph.com/docs/cody/enterprise/features
- AWS Kiro: spec-driven development, steering files, hooks, MCP, structured requirements/design/tasks.
  - https://aws.amazon.com/documentation-overview/kiro/
  - https://kiro.dev/docs/steering/
  - https://kiro.dev/docs/hooks.md
- Pieces.app: on-device long-term memory, workflow capture, timeline/search, MCP integration, local privacy posture.
  - https://docs.pieces.app/products/desktop/long-term-memory
  - https://docs.pieces.app/products/privacy-security-your-data
  - https://docs.pieces.app/products/mcp/cursor
- Aider: terminal AI pair programming, repo map, tree-sitter, PageRank, auto-commit workflow, lint/test loop.
  - https://aider.chat/
  - https://aider.chat/docs/repomap.html
- Cursor: project rules, skills, semantic codebase context, Explore subagent, team rules.
  - https://cursor.com/docs/rules
  - https://cursor.com/learn/understanding-your-codebase
- Windsurf/Devin Desktop: Cascade, real-time action awareness, rules, memories, checkpoints, agentic edits.
  - https://docs.windsurf.com/windsurf/cascade/cascade
- Claude Code: CLAUDE.md, hooks, skills, subagents, MCP, context compaction, agent teams.
  - https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
- CodeRabbit and Qodo: PR review summaries, repository context, pre-merge checks, review standards, multi-agent review, PR history/rules mining.
  - https://docs.coderabbit.ai/guides/code-review-overview
  - https://docs.coderabbit.ai/pr-reviews/pre-merge-checks
  - https://docs.qodo.ai/code-review
  - https://docs.qodo.ai/core-concepts/qodo-platform-core-capabilities
- DeepWiki: autogenerated repository wiki, diagrams, source links, chat, MCP, `.devin/wiki.json` steering.
  - https://cognition.ai/blog/deepwiki
  - https://docs.devin.ai/work-with-devin/deepwiki
- GitNexus: local graph, MCP tools/resources, agent skills, hooks, multi-repo registry, web UI.
  - https://abhigyanpatwari-gitnexus.mintlify.app/mcp/overview
  - https://github.com/abhigyanpatwari/GitNexus
- Cortex: local repo-scoped MCP context engine with code, rules, ADRs, graph relationships, optional semantic search.
  - https://registry.npmjs.org/@danielblomma/cortex-mcp
- Codebase-Memory: Tree-sitter knowledge graph, broad language coverage, benchmark claims, single binary, 14 MCP tools.
  - https://arxiv.org/html/2603.27277v1
  - https://github.com/DeusData/codebase-memory-mcp
- Repomix: one-command repo packing, token-count tree, budget guard, Tree-sitter compression, security checks.
  - https://github.com/yamadashy/repomix/
  - https://repomix.com/guide
- Tabnine Enterprise Context Engine: organizational knowledge graph, policies, ownership, dependencies, enterprise deployment.
  - https://www.tabnine.com/enterprise-context-engine/
  - https://www.tabnine.com/blog/introducing-the-tabnine-enterprise-context-engine/

---

## Issue-ready findings

The GitHub repository currently has these labels: `bug`, `documentation`, `duplicate`, `enhancement`, `good first issue`, `help wanted`, `invalid`, `question`, `wontfix`, `repo-audit`.

The issue drafts below include:

- `existing-labels`: labels that can be applied today.
- `proposed-labels`: labels requested by this research prompt but not all present in GitHub yet, especially `feature`, `suggestion`, and priority labels.

### 1. Publish reproducible software-understanding benchmarks

```yaml
title: "Publish reproducible software-understanding benchmarks"
type: enhancement
priority: P0
existing-labels: ["enhancement"]
proposed-labels: ["enhancement", "priority:P0", "area:benchmarks", "competitive-gap"]
verdict: ALIGN - Now
```

#### Background

Codebase-Memory publishes benchmark claims for answer quality, token reduction, and tool-call reduction. Greptile, Tabnine, and Sourcegraph frame their value around richer context quality. RepoNerve already has a strong token-economics story, but the public product surface needs reproducible evidence for "software understanding" outcomes.

#### Finding

RepoNerve's differentiation is evidence-backed why/who/what-breaks context, not only graph retrieval. Without a repeatable benchmark suite, buyers and contributors must trust positioning instead of seeing measurable reductions in exploration, tokens, and missed context.

#### User value

- Makes the token-economics claim concrete.
- Gives contributors a regression harness for context quality.
- Helps weaker-model use cases by proving RepoNerve can reduce blind exploration.

#### Acceptance criteria

- Define a benchmark corpus with at least three public repositories of different sizes/languages.
- Include task categories that test code structure, decisions/ADRs, ownership, feature understanding, and impact.
- Compare RepoNerve against a baseline file-exploration workflow and a context-packing workflow.
- Record token usage, tool calls, answer completeness, and evidence citation quality.
- Document how to rerun benchmarks locally without external LLM APIs required by RepoNerve itself.

#### Notes

Do not introduce embeddings as primary authority. If an LLM grader is used, keep reference answers and scoring rubric in-repo and deterministic enough for maintainer review.

---

### 2. Add an integration smoke test for MCP, skills, hooks, and PATH setup

```yaml
title: "Add integration smoke test for MCP, skills, hooks, and PATH setup"
type: bug
priority: P0
existing-labels: ["bug", "enhancement"]
proposed-labels: ["bug", "priority:P0", "area:install", "area:mcp", "competitive-gap"]
verdict: ALIGN - Now
```

#### Background

GitNexus and Codebase-Memory emphasize one-command agent setup. Kiro, Cursor, Claude Code, Cline, Pieces, and Continue all make MCP/rules/skills setup part of the product experience. RepoNerve already installs config across clients, but install trust depends on users proving that the binary, MCP server, skills, rules, doctor, and hooks all work together.

During this research run, `go install ./cmd/reponerve` succeeded but `reponerve` was not on PATH until `$(go env GOPATH)/bin` was added. That is common Go friction and should be detected early.

#### Finding

`reponerve doctor` checks repository freshness, but there is room for a first-run integration smoke test that validates the user's selected agent host can actually call RepoNerve.

#### User value

- Reduces failed first-run experiences.
- Helps users compare RepoNerve favorably against one-command competitors.
- Produces support-friendly diagnostics for PATH, MCP JSON, stdout pollution, missing memory, stale scan, and hook state.

#### Acceptance criteria

- Add or extend a command that verifies:
  - `reponerve` binary is discoverable from expected agent environment.
  - `.reponerve/config.yaml` and `.reponerve/memory.db` exist.
  - MCP `tools/list` returns valid JSON with the expected tool count.
  - Installed skill/rule files exist for supported local clients.
  - `reponerve hook status` is reported when hooks are installed or absent.
  - PATH remediation is suggested when `go env GOPATH` bin is missing.
- Return structured JSON for agent consumption.
- Add docs showing a single command users can run after `init` and `scan`.

#### Notes

This can extend `doctor` if it remains deterministic and does not require external services.

---

### 3. Create a public demo benchmark and "golden path" onboarding script

```yaml
title: "Create a public demo benchmark and golden-path onboarding script"
type: suggestion
priority: P0
existing-labels: ["documentation", "enhancement"]
proposed-labels: ["suggestion", "priority:P0", "area:growth", "area:onboarding"]
verdict: ALIGN - Now
```

#### Background

DeepWiki makes repo understanding instantly visible by swapping `github.com` for `deepwiki.com`. Repomix wins on a single command. GitNexus and Codebase-Memory emphasize one-command setup. RepoNerve has strong surfaces but the quickest proof path is spread across install, scan, MCP, ask, explain, plan, review, and explore docs.

#### Finding

RepoNerve needs a frictionless "see the magic" flow that demonstrates why/who/what-breaks understanding on a recognizable repository.

#### User value

- Gives new users a reliable first success.
- Creates shareable growth artifacts for README, docs, demos, and release posts.
- Shows the difference between raw code retrieval and preserved software understanding.

#### Acceptance criteria

- Add a documented demo script for at least one public repository.
- The script should install or use local RepoNerve, run `init`, `scan`, `doctor`, and execute representative `ask`, `explain`, `plan`, `impact`, and `review` queries.
- Capture expected output shape and explain what evidence a user should see.
- Include a short "if this fails" troubleshooting path.
- Keep the demo local-first and avoid mandatory SaaS.

#### Notes

Do not hardcode volatile exact output. Prefer robust examples and screenshots or asciinema-style artifacts if maintainers want richer demos later.

---

### 4. Add static repository-understanding export for shareable docs

```yaml
title: "Add static repository-understanding export for shareable docs"
type: feature
priority: P1
existing-labels: ["enhancement", "documentation"]
proposed-labels: ["feature", "priority:P1", "area:docs", "area:understanding"]
verdict: ALIGN - RFC
```

#### Background

DeepWiki and Devin Wiki turn repositories into navigable documentation with architecture diagrams, source links, and chat. RepoNerve's local Explore UI and context export are shipped, but a static, evidence-backed understanding export would help teams share onboarding material without requiring an active local UI.

#### Finding

RepoNerve has the underlying memory, graph, feature, ownership, and evidence layers to generate useful static docs. The gap is a packaged export that produces a browsable software-understanding artifact for humans and agents.

#### User value

- Helps day-one developers understand a repo without live tooling setup.
- Gives maintainers a public artifact for OSS growth.
- Complements local-first operation: generated output can be committed, hosted, or pasted.

#### Acceptance criteria

- RFC defines scope, output format, and non-goals.
- Export includes repository overview, feature map, decisions, ownership, impact hotspots, and evidence links.
- Output must be deterministic from local memory and code index.
- No cloud service required.
- Avoid full-product graph explorer scope; this is a static or bounded export.

#### Notes

Likely requires RFC because it introduces a new product surface and may affect Development Experience contracts.

---

### 5. Convert `pr-context` and `ship-check` into actionable PR annotations/check summaries

```yaml
title: "Convert pr-context and ship-check into actionable PR annotations/check summaries"
type: enhancement
priority: P1
existing-labels: ["enhancement"]
proposed-labels: ["enhancement", "priority:P1", "area:pr-context", "area:ci", "competitive-gap"]
verdict: ALIGN - RFC
```

#### Background

Greptile, CodeRabbit, and Qodo win adoption through PR comments, severity, pre-merge checks, and fix handoff. RepoNerve already ships `review`, `ship-check`, `pr-context`, and a workflow example, but the competitive surface should be easy to consume in GitHub PRs.

#### Finding

RepoNerve can differentiate in PRs by producing evidence-backed context and blockers, not generic AI comments. However, teams need native check summaries, severity, and traceability that are easy to wire into CI.

#### User value

- Makes RepoNerve useful in the existing code review workflow.
- Avoids competing as a full AI reviewer while still surfacing why/who/what-breaks evidence.
- Gives reviewers high-signal context before reading diffs.

#### Acceptance criteria

- RFC defines PR output contract, severity model, and non-goals.
- Workflow can publish a GitHub check summary or Markdown artifact from `pr-context` / `ship-check` output.
- Output distinguishes blockers, advisories, evidence, owners/reviewers, and changed files.
- No autonomous code modification or deployment.
- Include docs for enabling the workflow safely.

#### Notes

Keep MCP/CLI business logic thin; avoid embedding GitHub-specific assumptions deep in core services.

---

### 6. Surface deterministic review-standard candidates from repository evidence

```yaml
title: "Surface deterministic review-standard candidates from repository evidence"
type: feature
priority: P1
existing-labels: ["enhancement"]
proposed-labels: ["feature", "priority:P1", "area:discipline", "area:review"]
verdict: ALIGN - RFC
```

#### Background

Greptile learns team preferences from feedback. Qodo mines PR history into review standards. Cursor, Kiro, Windsurf, Claude Code, and Cline rely on rules or steering files. RepoNerve already installs Native Development Discipline and writes repo-adaptive policy from scan evidence.

#### Finding

RepoNerve can compete without opaque preference learning by proposing deterministic, evidence-backed rule candidates from ADRs, existing rules, recurring review comments, and repository patterns.

#### User value

- Helps teams preserve standards before they become tribal knowledge.
- Reduces repeated review comments.
- Keeps humans in control of accepted discipline rules.

#### Acceptance criteria

- RFC defines sources, evidence requirements, and approval flow.
- Candidate standards include evidence links and confidence rationale.
- Maintainers can accept, reject, or defer candidates.
- Accepted rules update a version-controlled policy or rule file; rejected rules are not silently reintroduced.
- No semantic/vector search as primary authority.

#### Notes

This is not autonomous rule enforcement. It is evidence-backed recommendation plus human acceptance.

---

### 7. Add host-specific lifecycle hook recipes for evidence checks

```yaml
title: "Add host-specific lifecycle hook recipes for evidence checks"
type: enhancement
priority: P1
existing-labels: ["documentation", "enhancement"]
proposed-labels: ["enhancement", "priority:P1", "area:integrations", "area:hooks"]
verdict: ALIGN - Now
```

#### Background

Kiro, Claude Code, Cline, GitNexus, and Windsurf increasingly use hooks to enforce workflow events. RepoNerve already has `reponerve hook install` for post-commit scan freshness, and Native Development Discipline tells agents when to call `plan`, `reuse-check`, `ship-check`, and `review`.

#### Finding

RepoNerve can make evidence checks more automatic by documenting host-specific hook recipes that call existing CLI commands at safe lifecycle points.

#### User value

- Helps agents run RepoNerve at the right moment without relying only on prompt instructions.
- Keeps deterministic checks outside the model context when possible.
- Improves adoption for teams already using Claude Code, Kiro, Cline, or Cursor rules/hooks.

#### Acceptance criteria

- Add recipes for at least Claude Code hooks, Kiro hooks, Cline rules/hooks, and Cursor rules/skills where supported.
- Each recipe maps one lifecycle event to an existing RepoNerve command.
- Include expected JSON output and failure behavior.
- Make clear that hooks do not modify code autonomously.
- Keep examples local-first and secret-safe.

#### Notes

Start as documentation and examples before adding generator behavior.

---

### 8. Publish language coverage and parser-quality diagnostics

```yaml
title: "Publish language coverage and parser-quality diagnostics"
type: enhancement
priority: P1
existing-labels: ["enhancement", "documentation"]
proposed-labels: ["enhancement", "priority:P1", "area:code-intelligence", "competitive-gap"]
verdict: ALIGN - Now
```

#### Background

Codebase-Memory markets 158 or more Tree-sitter languages, hybrid LSP type resolution, and speed benchmarks. Aider markets broad Tree-sitter repo maps. RepoNerve supports Go plus 19 Tree-sitter languages and deeper repository-memory linking.

#### Finding

RepoNerve should avoid a language-count race, but users need transparent diagnostics about which files were indexed, skipped, partially parsed, or linked to repository evidence.

#### User value

- Sets accurate expectations for polyglot repositories.
- Turns "only 20 languages" into an honest quality story.
- Helps contributors target parser and linker gaps.

#### Acceptance criteria

- Add docs listing supported languages, entity types, and known limitations.
- Ensure `doctor` or scan output reports unsupported file types and parse/link coverage.
- Include a compact summary suitable for agent consumption.
- Add at least one fixture or test that verifies diagnostics on mixed-language repositories.

#### Notes

This can be incremental and should preserve deterministic extraction.

---

### 9. Improve local Explore UI onboarding for non-CLI evaluation

```yaml
title: "Improve local Explore UI onboarding for non-CLI evaluation"
type: enhancement
priority: P1
existing-labels: ["enhancement"]
proposed-labels: ["enhancement", "priority:P1", "area:explore", "area:ux"]
verdict: ALIGN - Now
```

#### Background

DeepWiki, GitNexus, Pieces, and graph tools make value visible through diagrams, timelines, or web UIs. RepoNerve ships a capped local Explore UI, but competitive evaluation often happens before users understand the CLI.

#### Finding

The local UI should guide evaluators toward RepoNerve's differentiators: features, decisions, ownership, impact, evidence, and token savings.

#### User value

- Makes software understanding tangible.
- Helps users explain RepoNerve to teammates.
- Reduces "what do I do next?" friction after scan.

#### Acceptance criteria

- Add an onboarding panel or first-run guide to local Explore UI.
- Include example queries for why/who/what-breaks workflows.
- Surface scan freshness and evidence counts.
- Link to CLI/MCP next steps without requiring cloud services.
- Keep scope within capped local UI; do not expand to full enterprise graph explorer.

#### Notes

No new architecture should be required if this stays within current Explore UI scope.

---

### 10. Clarify repository memory versus personal/workflow memory

```yaml
title: "Clarify repository memory versus personal workflow memory"
type: suggestion
priority: P1
existing-labels: ["documentation"]
proposed-labels: ["suggestion", "priority:P1", "area:positioning", "area:privacy"]
verdict: ALIGN - Now
```

#### Background

Pieces.app captures OS-level workflow context and personal long-term memory. Windsurf and Cline have persistent memories. Claude Code and Cursor rely on project rules, skills, and file-backed memory. RepoNerve has repository memory and session `remember` / `forget`, but its moat is remembering the repository, not screen/audio/personal behavior.

#### Finding

Users may confuse RepoNerve with personal AI memory tools. Clear positioning and privacy docs should state what RepoNerve records, what it never records, and how repository memory differs from OS-level capture.

#### User value

- Builds trust with local-first/privacy-sensitive users.
- Prevents feature requests that would pull RepoNerve toward out-of-scope personal surveillance.
- Makes composition with Pieces/Mem0/MemNexus-style tools clearer.

#### Acceptance criteria

- Add a docs page or section comparing repository memory, session memory, and personal workflow memory.
- State that RepoNerve does not capture screen, clipboard, audio, browser history, or personal activity.
- Explain `remember` / `forget` scope and storage.
- Include a comparison table for Pieces, Cline Memory Bank, Cursor/Claude rules, and RepoNerve.

#### Notes

This is documentation/positioning, not a new memory subsystem.

---

### 11. Add competitor-specific migration and composition guides

```yaml
title: "Add competitor-specific migration and composition guides"
type: suggestion
priority: P2
existing-labels: ["documentation"]
proposed-labels: ["suggestion", "priority:P2", "area:growth", "area:docs"]
verdict: ALIGN - Now
```

#### Background

RepoNerve composes with AI IDEs and agents rather than replacing them. Users evaluating Cursor, Claude Code, Cline, Aider, Sourcegraph Cody, Pieces, GitNexus, and Repomix need help understanding when to use RepoNerve alongside each tool.

#### Finding

The product positioning doc has a strong competitive narrative, but the docs need practical "if you use X, add RepoNerve for Y" guides.

#### User value

- Turns competitive confusion into adoption.
- Helps users keep their preferred agent while adding evidence-backed repository memory.
- Reduces pressure to duplicate autonomous coding features.

#### Acceptance criteria

- Add short guides for at least Cursor, Claude Code, Cline, Aider, Pieces, Sourcegraph Cody, Greptile/CodeRabbit/Qodo, GitNexus/Codebase-Memory, and Repomix.
- Each guide states:
  - What the other tool is good at.
  - What RepoNerve adds.
  - Minimal setup.
  - Example prompt/workflow.
  - Non-goals and composition boundaries.
- Link guides from market positioning and AI chat integration docs.

#### Notes

Keep claims factual and sourced. Avoid disparaging competitors.

---

### 12. Add context-pack token budget visualizations

```yaml
title: "Add context-pack token budget visualizations"
type: enhancement
priority: P2
existing-labels: ["enhancement"]
proposed-labels: ["enhancement", "priority:P2", "area:token-intelligence", "area:ux"]
verdict: ALIGN - Now
```

#### Background

Repomix makes token cost visible with token-count trees and budgets. RepoNerve has bounded context, graph-aware compression, compact/prose/json output, and token budgets, but users do not have an obvious way to inspect why a context pack consumed tokens.

#### Finding

RepoNerve can improve UX by exposing token budget allocation across evidence categories: code, decisions, ownership, graph, feature, and guidance.

#### User value

- Makes compression behavior explainable.
- Helps users tune prompts and budgets.
- Reinforces the token-economics moat with concrete UX.

#### Acceptance criteria

- Add an optional compact token breakdown to context-producing commands.
- Show budget used and truncated categories when applicable.
- Preserve current default output compatibility unless an RFC determines a contract change is needed.
- Include tests for deterministic ordering and totals.

#### Notes

If this changes JSON envelope contracts, route through RFC/versioning. Otherwise keep it additive.

---

### 13. Document boundaries for cross-repo and enterprise context

```yaml
title: "Document boundaries for cross-repo and enterprise context"
type: suggestion
priority: P2
existing-labels: ["documentation"]
proposed-labels: ["suggestion", "priority:P2", "area:positioning", "area:enterprise"]
verdict: ALIGN - Now
```

#### Background

Sourcegraph Cody, Tabnine Enterprise Context Engine, Qodo, GitNexus, and Codebase-Memory all position around cross-repo or enterprise-scale context. RepoNerve explicitly keeps Sourcegraph-scale cross-repo enterprise federation out of scope for the current release line.

#### Finding

RepoNerve should state what scoped monorepo support, local repository memory, and MCP integrations do and do not cover for enterprise users.

#### User value

- Sets expectations for teams with many repositories.
- Prevents accidental roadmap drift into out-of-scope federation.
- Helps evaluate when RepoNerve complements Sourcegraph/Tabnine rather than replacing them.

#### Acceptance criteria

- Add or update docs explaining:
  - Single-repo local-first scope.
  - Scoped monorepo scan support.
  - What "cross-repo federation" would require and why it is out of scope without RFC.
  - How teams can use RepoNerve per repo today.
- Link from market positioning and install/integration docs.

#### Notes

This is a documentation issue unless maintainers choose to reconsider cross-repo scope through RFC.

---

## Recommended first batch

If maintainers want to file only a small initial batch, prioritize:

1. Publish reproducible software-understanding benchmarks (P0)
2. Add integration smoke test for MCP, skills, hooks, and PATH setup (P0)
3. Create a public demo benchmark and golden-path onboarding script (P0)
4. Convert `pr-context` and `ship-check` into actionable PR annotations/check summaries (P1, RFC)
5. Surface deterministic review-standard candidates from repository evidence (P1, RFC)

These are the highest leverage because they improve proof, first-run trust, growth, and PR workflow adoption without violating local-first or evidence-backed boundaries.
