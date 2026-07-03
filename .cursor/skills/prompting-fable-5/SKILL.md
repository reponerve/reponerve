---
name: prompting-fable-5
description: >-
  Prompting patterns and scaffolding guidance for Claude Fable 5 and Claude
  Mythos 5: effort, instruction following, long autonomous runs, memory,
  subagents, and migration from Claude Opus 4.8. Use when writing or reviewing
  system prompts, agent harnesses, or skills that target Fable 5.
trigger: /prompting-fable-5
---

# /prompting-fable-5

Behavioral differences and prompting patterns for **Claude Fable 5** and **Claude Mythos 5**.

Fable 5 takes on problems that were previously too complex, long-running, or ambiguous for prior models. The teams seeing the best outcomes apply it to their hardest unsolved problems; testing it only on simpler workloads undersells its capability range. Capability improvements at this level are also a good prompt to re-evaluate which instructions, tools, and guardrails are still needed.

---

## When to use

- "Write a system prompt for an agent running on Fable 5"
- "Why is my Fable 5 agent overplanning / too verbose / stopping early?"
- "Migrate this prompt/skill/harness from Opus 4.8 to Fable 5"
- "Review this scaffolding for a long-running autonomous agent"
- `/prompting-fable-5` or `/prompting-fable-5 <symptom or prompt to review>`

All copy-paste snippets live in `.cursor/skills/prompting-fable-5/reference.md`.

---

## What changed vs. Claude Opus 4.8

| Area | Fable 5 behavior | Prompting consequence |
| --- | --- | --- |
| Long-horizon autonomy | Sustains multi-day goal-directed runs with strong instruction retention | Ground progress claims in tool evidence; add memory system |
| Turn length | Single requests can run many minutes at high effort; autonomous runs for hours | Raise client timeouts, stream, check runs asynchronously |
| Effort control | `high` default; `xhigh` for capability-sensitive work; `medium`/`low` still strong | Effort is the primary intelligence/latency/cost dial |
| Instruction following | One brief instruction steers a behavior class; no need to enumerate cases | Trim prescriptive prompts; over-specification degrades output |
| First-shot correctness | Single-pass implementations of previously multi-day systems | Assign harder, well-specified tasks; let it scope and clarify |
| Subagents | Dispatches parallel subagents readily; manages long-lived peers | Encourage delegation; prefer async over blocking |
| Vision | Dense technical images and screenshots read accurately; uses bash/crop tools | Fewer image-handling workarounds needed |
| Refusals | Safety classifiers for offensive cyber, bio/life sciences, reasoning extraction | Configure fallback to Opus 4.8; audit "show your thinking" prompts |

API changes (adaptive thinking only, summarized-only thinking output, no extended thinking budgets, the `refusal` stop reason) are platform docs, not prompt patterns — see the Anthropic model introduction page.

---

## Symptom → pattern

Diagnose the observed behavior, then apply the matching snippet from `reference.md`.

| Symptom | Pattern | Snippet |
| --- | --- | --- |
| Overplanning on ambiguous tasks; narrating options it won't pursue | Act-when-ready instruction | §1 Bias to action |
| Unrequested tidying, refactoring, speculative abstractions at high effort | Scope constraint | §2 Simplicity constraint |
| Verbose output: option surveys, long root-cause essays, narrating comments | Brief brevity instruction (do not enumerate each pattern) | §3 Lead with the outcome |
| Stops at unnecessary checkpoints in long workflows | Define when pausing is genuinely required | §4 Checkpoint rule |
| Fabricated or unverified status reports on long runs | Audit claims against tool results | §5 Ground progress claims |
| Unrequested actions (drafting emails, defensive git backups) | Explicit action boundaries | §6 State the boundaries |
| Underuses subagents, or blocks on each one | Delegation guidance + async communication | §7 Parallel subagents |
| Repeats mistakes across runs | Markdown-file memory system + bootstrap prompt | §8 Memory system |
| Ends turn with "I'll now run X" and no tool call; asks permission it doesn't need | Autonomy reminder | §9 Early stopping |
| Suggests new session / trims work in very long sessions | Hide token countdowns; add reassurance | §10 Context-budget concern |
| Generic output despite a clear request | Give the reason, not only the request | §11 Intent framing |
| Final summaries full of arrow chains, invented labels, working shorthand | Communication-style addendum | §12 Readability |
| Mid-task deliverables get summarized or lost in long async runs | `send_to_user` tool + elicitation instruction | §13 Send-to-user tool |

---

## Migration checklist (Opus 4.8 → Fable 5)

1. **Start at the top of your difficulty range.** Pick a task harder than what you'd assign prior models; have Fable 5 scope it, ask clarifying questions, and execute.
2. **Raise timeouts and go async.** Adjust client timeouts, streaming, and progress indicators; restructure harnesses to check runs via scheduled jobs rather than blocking.
3. **Re-tune effort.** Default `high`; reserve `xhigh` for capability-sensitive work; drop to `medium`/`low` for routine tasks — lower effort on Fable 5 often exceeds `xhigh` on prior models.
4. **Refactor existing prompts and skills.** Instructions written for prior models are often too prescriptive and degrade Fable 5 output. Remove instructions where default behavior is now better. Fable 5 also updates skills on the fly from task learnings.
5. **Make self-verification explicit in long-run prompts.** Fresh-context verifier subagents outperform self-critique — see reference §14.
6. **Audit for reasoning-echo instructions.** Prompts that tell the model to echo or transcribe its internal reasoning trigger the `reasoning_extraction` refusal category and elevate fallbacks. Read structured `thinking` blocks instead; surface progress via `send_to_user`.
7. **Configure refusal fallback.** Benign cybersecurity and life-sciences work can trip safety classifiers; set server-side or client-side fallback to Claude Opus 4.8.
8. **Add a memory file and a send-to-user tool** for long-running or asynchronous agents (reference §8, §13).

---

## Anti-patterns

```text
BAD:  Enumerate every verbose behavior by name ("don't survey options, don't
      write long root-cause essays, don't...") — one brief instruction suffices
BAD:  Keep prescriptive Opus-era step-by-step instructions that fight better defaults
BAD:  Show a remaining-token countdown to the model
BAD:  Instruct the model to reproduce its reasoning in the response
BAD:  Define send_to_user without an elicitation instruction (it will rarely call it)
BAD:  Route narration or internal reasoning through send_to_user
BAD:  Test Fable 5 only on simple workloads and conclude it is unremarkable
GOOD: Short behavioral instructions + effort tuning + evidence-grounded
      progress + async harness + memory file
```

---

## Reference

Full copy-paste snippet library: `.cursor/skills/prompting-fable-5/reference.md`
