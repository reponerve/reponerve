# RepoNerve Token Economics

Version: 1.0

Status: Draft

Updated: 2026-06-11

Related:

* `docs/vision/vision.md`
* `docs/product/market-positioning.md`
* `docs/architecture/agent-native-repository-intelligence.md`

---

# Thesis

As LLMs become more capable and more expensive, the bottleneck is not generation — it is **how much context is burned before useful work begins**.

RepoNerve reduces the **cost of software understanding** by moving understanding out of the token meter and into local, deterministic infrastructure.

Users buy:

* Understanding
* Development speed
* Confidence
* **Reduced exploration and token consumption**

RepoNerve does not make LLMs cheaper. It makes **using premium models efficiently the default**.

---

# The Exploration Tax

Typical agent workflow today:

```text
User: "How does auth work?"
Agent: read 12 files → grep → git log → read 8 more files → summarize
Cost: 40k–120k tokens before the first useful answer
```

Every session, teammate, and agent handoff repeats this. Context limits are hit not because the model failed, but because the workflow is wasteful.

---

# Cost Model Inversion

| Traditional workflow | RepoNerve workflow |
| --- | --- |
| Pay LLM tokens to *discover* the repo every session | Pay **zero LLM** to index once (`reponerve scan`) |
| Pay LLM to read files, grep, blame | Pay **near-zero** for structured MCP queries |
| Pay LLM to summarize what it just read | Pay LLM only for **reasoning and implementation** |
| Context grows unbounded | Context delivered in **token-budget packages** |

```text
EXPENSIVE:  LLM reads repo → LLM understands → LLM acts
CHEAP:      RepoNerve understands → LLM receives package → LLM acts
```

---

# Optimization Stack

## Layer 1: Understanding First, Evidence Second, AI Third

| Task | Executor | LLM cost |
| --- | --- | --- |
| AST parsing, symbol extraction | Deterministic (`go/ast`) | Zero |
| ADR/commit/decision extraction | Rule-based extractors | Zero |
| Ownership from git history | Deterministic rollup | Zero |
| Graph traversal, impact analysis | SQLite + graph engine | Zero |
| Intent/tradeoff interpretation | LLM (optional, targeted) | Minimal |

## Layer 2: Pre-Indexed Repository Memory

`reponerve scan` is a fixed-cost investment. Queries cost tokens proportional to answer size, not repository size.

## Layer 3: Token-Efficient Delivery

Development Experience and MCP expose **evidence-backed context packs**:

* Relevance-ranked subgraph for the task
* Truncation by token budget (not naive list limits)
* Structured output formats (`caveman` format — ISSUE-060, v1.0)

## Layer 4: MCP as Surgical Interface

27 MCP tools return bounded, structured responses instead of raw file dumps. Fewer tool calls, smaller responses.

## Layer 5: Durable Understanding Across Sessions

Understanding persists in `.reponerve/`. Session 50 recalls auth context in hundreds of tokens, not hundreds of thousands.

## Layer 6: Composability with RTK (adjacent tools)

RTK compresses shell output. RepoNerve compresses understanding. Together they address the two largest token sinks in agent sessions.

---

# Example: Same Task, Different Cost

**Task:** Add retry logic to payment service.

**Without RepoNerve (premium model):**

* 15 file reads, 3 greps, 2 git blames → ~60k exploration tokens
* Implementation → ~20k tokens
* Limit hit before tests complete

**With RepoNerve (premium model):**

* MCP `explain` + `analyze_impact` + change plan → ~2k tokens
* Implementation → ~15k tokens
* ~4× headroom remains in the same context window

---

# Target Metrics

| Metric | Target |
| --- | --- |
| Exploration tokens per task | 80–95% reduction |
| Tool calls per agent task | 50–70% reduction |
| Time-to-understanding (onboarding) | Days → hours |
| Productive sessions before context limit | 2–3× improvement |

---

# Implementation Status

| Cost lever | Status |
| --- | --- |
| Deterministic extraction (no LLM scan) | ✅ Shipped |
| Repository memory + graph + MCP | ✅ Shipped |
| Token-efficient context packs | ⚠️ Basic truncation; graph-aware budgeting planned |
| Code intelligence (fewer file reads) | ❌ ISSUE-057 |
| Development Experience commands | ❌ ISSUE-057 (mostly stubs) |
| Structured/caveman output format | ❌ ISSUE-060 (v1.0) |
| Incremental scan on commit | ❌ ISSUE-060 (v1.0) |

See `docs/product/implementation-status.md` for full gap analysis.

---

# Guiding Principle

Understanding first.

Evidence second.

AI third.

Premium models should spend tokens on building and deciding — not on re-learning a repository that already knew the answers.
