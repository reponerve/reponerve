---
name: product-triage
description: >-
  Triage GitHub issues and feature ideas against RepoNerve vision, mission, and
  out-of-scope policy. Use when reviewing open issues, prioritizing roadmap,
  evaluating enhancements, or asking what RepoNerve should build next.
trigger: /product-triage
---

# /product-triage

Vision-aligned triage for **reponerve/reponerve** issues and feature requests.

**Principle:** Understanding first. Evidence second. AI third.

---

## When to use

- "What issues should we prioritize?"
- "Is this feature aligned with our vision?"
- "Review open GitHub issues"
- "What enhancements do we actually need?"
- `/product-triage` or `/product-triage issue 42`

---

## Prerequisites

1. **RepoNerve MCP** connected (Settings → Tools & MCP → `reponerve`) **or** CLI on PATH
2. **Memory:** `test -f .reponerve/memory.db || (reponerve init && reponerve scan)`
3. **GitHub CLI:** `gh auth status` (for issue fetch)

---

## Workflow (follow in order)

### Step 1 — Load product ground truth

Read these (RepoNerve `ask` first; read files only if ask is insufficient):

```bash
reponerve ask "What is RepoNerve vision, mission, and what is explicitly out of scope for post-1.0?" --json --format compact --token-budget 2000
```

Authoritative docs (for rubric details):

- `docs/vision/vision.md` — vision, mission, differentiation
- `docs/roadmap/v1.x-backlog.md` — **out of scope** (reject unless new RFC)
- `docs/product/implementation-status.md` — already shipped
- `docs/governance/rfc-process.md` — RFC required for significant work
- `docs/releases/versioning.md` — semver + release line

### Step 2 — Fetch open issues

```bash
./scripts/product-triage-issues.sh
# or: gh issue list --repo reponerve/reponerve --state open --limit 50 --json number,title,labels,body,url,createdAt
```

For a single issue: `gh issue view <n> --repo reponerve/reponerve --json number,title,labels,body,url,comments`

### Step 3 — Score each item

Use the rubric in `.cursor/skills/product-triage/reference.md`. For every issue or pasted idea, answer:

| Question | Source |
| --- | --- |
| Does it advance **Software Understanding** (not just code retrieval)? | vision.md |
| Is it **local-first** and **evidence-backed**? | vision.md, ADRs |
| Does it extend **Development Experience** or core memory/code layers? | implementation-status |
| Is it **already shipped** or duplicate? | `reponerve ask`, implementation-status |
| Is it **explicitly out of scope**? | v1.x-backlog.md → **Reject** |
| Does it need an **RFC** before implementation? | rfc-process.md |

**Verdicts (pick one):**

| Verdict | Meaning |
| --- | --- |
| **ALIGN — Now** | Clear vision fit; fills a real gap; no RFC blocker for small work |
| **ALIGN — RFC** | Vision fit but significant; requires RFC + versioning row before code |
| **DEFER** | Reasonable but lower priority; document in backlog, not now |
| **DUPLICATE** | Already exists; link to command/doc/issue |
| **REJECT** | Out of scope, wrong layer, or conflicts with local-first / evidence principles |

### Step 4 — Check reuse before recommending new work

For items marked ALIGN, run:

```bash
reponerve reuse-check "<issue title or feature intent>" --json
reponerve ask "Do we already support <capability>?" --json --format compact --token-budget 1500
```

Prefer extending existing engines over new subsystems.

### Step 5 — Deliver triage report

Use this format:

```markdown
## Product triage — <date>

**Vision anchor:** <one sentence from vision.md>

### Recommended now (≤3)
| Issue | Verdict | Why | Next step |
| --- | --- | --- | --- |

### RFC required
| Issue | Why RFC | Suggested RFC title |

### Defer
| Issue | Reason |

### Reject / duplicate
| Issue | Verdict | Evidence |

### Gaps (no open issue, but vision-aligned)
<bullets only if evidence from implementation-status or ask supports a real gap>
```

Do **not** invent gaps. If RepoNerve has no evidence, say so.

---

## Anti-patterns

```text
BAD:  Prioritize by vote count or loudest request alone
BAD:  Recommend semantic/vector search without noting v1.x out-of-scope
BAD:  Suggest cloud SaaS core without RFC and vision exception
BAD:  Triage without reading v1.x-backlog out-of-scope list
GOOD: vision + implementation-status + gh issues + reuse-check → verdict table
```

---

## Reference

Full rubric and out-of-scope list: `.cursor/skills/product-triage/reference.md`

## Related

- **Mechanical audit + daily issue filing:** `.cursor/skills/repo-audit/SKILL.md` and `.github/workflows/repo-audit.yml`
- After audit issues are filed, prioritize with `/product-triage` (filter label `repo-audit`)
