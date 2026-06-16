# Universal Understanding

Version: 1.0

Status: Active

Updated: 2026-06-11

Related:

* `docs/vision/vision.md`
* `docs/architecture/agent-context-contract.md`
* `docs/product/token-economics.md`
* `.cursor/skills/reponerve/SKILL.md`

---

# North Star

Anyone — a developer on day one, a reviewer unfamiliar with the area, or an AI agent (including weaker models) — should be able to **understand and work on a repository without worrying about context**.

RepoNerve carries the context burden. AI narrates and implements from **evidence**, not from exploration, guesswork, or hallucination.

```text
User intent (question or pasted task)
        ↓
RepoNerve (deterministic scan + memory + code index)
        ↓
Structured context package (briefings, scope, constraints)
        ↓
Human or AI (explain, plan, edit — minimal extra discovery)
```

---

# Who This Serves

| Persona | Need | RepoNerve entry |
| --- | --- | --- |
| **Day-one developer** | What does this repo do? How is it organized? | `ask`, `explain`, `list_decisions` |
| **Assigned engineer** | Paste ticket → where to start, what breaks | `plan`, `analyze_topic_impact`, `explain_*` |
| **Reviewer** | What to check, who knows this area | `review`, `explain` |
| **Weak AI model** | Pre-digested facts, not open-ended repo search | MCP `structured` + `agent` metadata; any IDE chat |
| **Strong AI model** | Same contract — less token waste, fewer wrong edits | Same path |
| **Any IDE chat** | Natural language → RepoNerve tools | MCP in Copilot, Cursor, JetBrains, … — see `docs/ai-chat-integration.md` |
| **Web LLM** | No MCP in browser | `reponerve context export` → paste evidence |

Capability of the model must not determine whether understanding is available. **RepoNerve quality** determines that.

---

# Task Intake (Paste a Description)

When someone pastes a task — Jira ticket, Slack message, PRD snippet — RepoNerve is the first step, not file grep.

## Workflow

```text
1. ask "<pasted task>"  → task_plan with plan, briefings, suggested_steps
   OR plan("<pasted task>")
2. explain_file / explain_* on starting_points only
3. analyze_topic_impact on risky areas
4. implement within scope
5. review("<topic from task>")
```

Example:

```bash
reponerve ask "Add Google OAuth login to the API"
reponerve onboard "Add Google OAuth login to the API"
reponerve impact "OAuth login"
reponerve explain-file internal/auth/handler.go
```

`ask` detects task descriptions (`add`, `implement`, `fix`, long pasted tickets) and returns `answer_type: task_plan` — weak models get a bounded package without picking the right tool.

The agent equivalent: paste the ticket to MCP `ask` or `onboard`, then follow `agent.recommended_next_tools` until `completeness` is `full` before editing.

---

# First-Day Onboarding

Suggested sequence for someone new to the repository:

```bash
reponerve onboard                           # key decisions + orientation
reponerve onboard "My first assignment..."  # + assignment plan + briefings
```

Or stepwise:

1. **Orientation** — `ask "What does this repository do?"` / `explain "<product area>"`
2. **Architecture** — `list_decisions`, `trace_decision` on key ADRs
3. **Code map** — `explain` on domains they will touch (packages, services)
4. **Assignment** — `ask` or `plan` with pasted task text

No tribal knowledge required. No "ask Sarah" — RepoNerve surfaces decisions, owners, and code links from evidence.

---

# Anti-Hallucination Design

Hallucination and confusion waste tokens and produce wrong edits. RepoNerve prevents this by design.

## RepoNerve must not

* Invent Purpose, History, or narrative fields not in memory
* Guess ownership without git evidence
* Rank or score without traceable payloads
* Present `search_summary` as complete understanding
* Hide symbol ambiguity (homonyms must surface as multiple briefings)

## RepoNerve must

* Return deterministic, reproducible context for the same repo state
* Label completeness (`full`, `partial`, `retrieval_only`)
* Attach `evidence` and `source_services` to conclusions
* Tell agents what to do next (`guidance_for_agent`, `recommended_next_tools`)
* Fail closed: when context is thin, recommend more RepoNerve queries — not guessing

## Agents must not

* Edit code when `agent.completeness` is `retrieval_only`
* Invent file paths, types, or ADRs not in `structured`
* Pick one homonym without comparing briefings
* Re-read the whole repo when briefings already contain `defined_in` and relationships
* Narrate architecture without `related_decisions` or evidence when those exist

When evidence is missing, say **"RepoNerve has no indexed evidence for X — run scan or narrow the question"** — not a fabricated answer.

---

# Weak-Model Friendly Context

Weaker models succeed when context is **structured, bounded, and prescriptive** — not when they are asked to explore.

| Technique | Why it helps |
| --- | --- |
| `entity_briefings` with `defined_in`, `fields`, `role` | No need to open files to know what a type is |
| `plan.starting_points` | Bounded file set for first edits |
| `agent.must_use_before_edit` | Hard gate before changes |
| `agent.completeness` | Model knows if it can answer or must query more |
| Deterministic ordering | Same repo → same context → less drift |
| Short `formatted`, rich `structured` | Model reads JSON fields, not prose to parse |

RepoNerve does the hard work once at `scan`. Models spend tokens on **reasoning and edits**, not rediscovery.

---

# Token Discipline

Confusion is expensive. Wrong exploration paths burn context windows.

```text
EXPENSIVE:  grep → read 20 files → guess → wrong file → read 10 more → fix
CHEAP:      plan → briefings → edit 2 scoped files → review
```

Rules:

1. **One RepoNerve pass before bulk file reads**
2. **Stop exploring** when `completeness: full` and briefings cover the symbol
3. **Escalate via RepoNerve tools**, not ad-hoc search, when completeness is partial
4. **Never implement** from retrieval-only context

See `docs/product/token-economics.md` for the full cost model.

---

# Success Criteria

Universal understanding is achieved when:

* A pasted task description → `plan` → scoped, evidence-backed starting points
* Day-one `ask` / `explain` answers without requiring contributor access
* Weak and strong models receive the same structured contract
* Hallucination paths are blocked by completeness gates and evidence requirements
* Token spend shifts from exploration to implementation

Implementation tracking: ISSUE-057 (Development Experience), agent context contract, entity briefings, MCP `agent` envelope.
