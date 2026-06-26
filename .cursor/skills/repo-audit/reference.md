# Agent findings schema

Append qualitative findings to `.reponerve/audit-findings-agent.json` before re-running `repo-audit.sh`.

## File shape

```json
{
  "findings": [
    {
      "id": "unique-kebab-id",
      "severity": "high | medium | low | info",
      "category": "security | test | health | quality | tech-debt | vision",
      "title": "Short imperative title (no prefix)",
      "body": "Markdown with evidence: paths, symbols, RepoNerve envelope excerpts",
      "labels": ["repo-audit", "bug"],
      "source": "cursor-agent"
    }
  ]
}
```

## Severity guide

| Level | Examples |
| --- | --- |
| **high** | Test failures, govulncheck CVEs, doctor `fail`, data loss risk |
| **medium** | doctor `warn`, go vet, stale scan, missing error handling with evidence |
| **low** | Docs drift, minor tech debt, optional hook not installed |
| **info** | Suggestions; usually skip issue filing |

## Labels

Always include `repo-audit`. Add one of: `bug`, `enhancement`, `documentation`, `security` (create on GitHub if missing).

## Body template

```markdown
## Summary
<one paragraph>

## Evidence
- `path/to/file.go` — symbol or behavior
- RepoNerve: `reponerve explain-function "Foo" --package bar`

## Suggested fix
<concrete next step>

## Vision alignment
<one line tying to software understanding / local-first / evidence>

_Agent finding from /repo-audit._
```

## Do not file

- Out-of-scope items from `docs/roadmap/v1.x-backlog.md`
- Duplicates of open `[audit:*]` issues
- Speculative issues without repository evidence
