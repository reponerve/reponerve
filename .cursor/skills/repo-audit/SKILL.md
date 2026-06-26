---
name: repo-audit
description: >-
  Audit reponerve repository code for health, security, and vision-aligned
  improvements; recommend next work and file GitHub issues. Use for vulnerability
  scans, what to build next, daily improvement, or /repo-audit.
trigger: /repo-audit
---

# /repo-audit

Continuous improvement loop: **analyze code → align with vision → file GitHub issues**.

Works with RepoNerve MCP/CLI + `gh`. Pair with `/product-triage` for prioritization.

---

## When to use

- "What should we do next on RepoNerve?"
- "Audit the codebase for issues and vulnerabilities"
- "Create GitHub issues from findings"
- "Daily improvement scan"
- `/repo-audit` or `/repo-audit --create-issues`

---

## Prerequisites

```bash
test -f .reponerve/memory.db || (reponerve init && reponerve scan)
gh auth status
go install golang.org/x/vuln/cmd/govulncheck@latest   # recommended
```

RepoNerve MCP connected in Cursor (Settings → Tools & MCP).

---

## Workflow

### Step 1 — Mechanical audit

```bash
./scripts/repo-audit.sh
```

Runs: `reponerve doctor --json`, `go vet`, `go test`, `govulncheck` (if installed).  
Output: `.reponerve/audit-report.json`

### Step 2 — RepoNerve intelligence (evidence-backed gaps)

```bash
reponerve review "repository health improvement opportunities" --json --format compact --token-budget 2000
reponerve ask "What shipped capabilities have gaps, missing tests, or docs drift?" --json --format compact --token-budget 1500
reponerve ship-check "repository maintenance" --json --format compact --token-budget 1500
```

Use `structured` + `agent.ship_blockers` / `agent.advisories` — not grep-first exploration.

### Step 3 — Vision filter (required before filing)

Read `.cursor/skills/product-triage/reference.md`. For each proposed finding:

| Filter | Action |
| --- | --- |
| Out of scope (semantic search, cloud core, …) | **Do not file** — note in report only |
| Already tracked open issue | **Skip** — link existing |
| Mechanical (tests, vulns, doctor) | **File** with `repo-audit` label |
| Product enhancement | **File** only if vision-aligned; prefer **RFC** label in body |

### Step 4 — Add agent findings (optional JSON)

Write qualitative findings to `.reponerve/audit-findings-agent.json`:

```json
{
  "findings": [
    {
      "id": "de-mcp-docs-drift",
      "severity": "low",
      "category": "tech-debt",
      "title": "Short title",
      "body": "Markdown body with evidence paths from RepoNerve output",
      "labels": ["repo-audit", "documentation", "enhancement"]
    }
  ]
}
```

Re-run merge:

```bash
./scripts/repo-audit.sh
```

### Step 5 — Present summary to user

```markdown
## Repo audit — <date>

### Critical (file issues)
| ID | Title | Source |

### Vision-aligned improvements
| Title | Why | RFC? |

### Skipped (out of scope / duplicate)
| Item | Reason |
```

### Step 6 — Create GitHub issues (only when user confirms)

**Preview:**

```bash
./scripts/create-audit-issues.sh --dry-run .reponerve/audit-report.json
```

**File issues in Cursor** (user must ask explicitly, e.g. "create the issues"):

```bash
./scripts/create-audit-issues.sh --yes .reponerve/audit-report.json
```

**Daily CI** files mechanical findings automatically — no approval needed (see Automation below).

Issues are titled `[audit:<id>] …` with label `repo-audit`. Duplicates are skipped automatically.

---

## Daily habit (Cursor)

1. `reponerve scan` (fresh memory)
2. `/repo-audit` — review report + add agent findings if needed
3. `/product-triage` — order the `repo-audit` backlog

Mechanical issues (tests, vulns, doctor) are filed daily by GitHub Actions.

---

## Automation

**Daily GitHub Action:** `.github/workflows/repo-audit.yml` — cron `0 6 * * *` UTC.

Each run:

1. Builds `reponerve` from source
2. `reponerve init --skip-ide` + `reponerve scan`
3. `./scripts/repo-audit.sh` (doctor, go vet, go test, govulncheck)
4. `./scripts/create-audit-issues.sh --yes` (deduplicated `[audit:<id>]` titles)

Manual re-run: **Actions → Daily repository audit → Run workflow**.

Pair with `/product-triage` to prioritize the backlog.

---

## Reference

- Vision rubric: `.cursor/skills/product-triage/reference.md`
- Issue triage: `.cursor/skills/product-triage/SKILL.md`
- Agent findings schema: `.cursor/skills/repo-audit/reference.md`
