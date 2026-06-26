#!/usr/bin/env bash
# Create GitHub issues from audit report JSON (skips duplicates).
#
# Usage:
#   ./scripts/create-audit-issues.sh [--dry-run|--yes] [report.json]
#   DRY_RUN=1 ./scripts/create-audit-issues.sh
#   CREATE_ISSUES=1 ./scripts/create-audit-issues.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

REPORT="${REPONERVE_AUDIT_REPORT:-.reponerve/audit-report.json}"
REPO="${REPONERVE_GITHUB_REPO:-reponerve/reponerve}"
DRY_RUN="${DRY_RUN:-0}"
CREATE_ISSUES="${CREATE_ISSUES:-0}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --yes)
      CREATE_ISSUES=1
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [--dry-run|--yes] [report.json]"
      exit 0
      ;;
    *)
      REPORT="$1"
      shift
      ;;
  esac
done

if [[ "$CREATE_ISSUES" != "1" && "$DRY_RUN" != "1" ]]; then
  echo "Pass --yes to file issues, or --dry-run to preview." >&2
  echo "Example: $0 --yes $REPORT" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI required (gh auth login)" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 1
fi

if [[ ! -f "$REPORT" ]]; then
  echo "Report not found: $REPORT (run ./scripts/repo-audit.sh first)" >&2
  exit 1
fi

gh label create "repo-audit" --repo "$REPO" --color "1D76DB" --description "Automated repository audit" 2>/dev/null || true

EXISTING_TITLES="$(gh issue list --repo "$REPO" --state open --label repo-audit --limit 200 --json title | jq -r '.[].title')"

CREATED=0
SKIPPED=0

while IFS= read -r row; do
  id="$(echo "$row" | jq -r '.id')"
  title="$(echo "$row" | jq -r '.title')"
  body="$(echo "$row" | jq -r '.body')"
  labels="$(echo "$row" | jq -r '[.labels[]?] | join(",")')"
  severity="$(echo "$row" | jq -r '.severity // "medium"')"

  issue_title="[audit:${id}] ${title}"

  if echo "$EXISTING_TITLES" | grep -Fxq "$issue_title"; then
    echo "SKIP (exists): $issue_title"
    SKIPPED=$((SKIPPED + 1))
    continue
  fi

  full_body="${body}

---
**Severity:** ${severity}
**Audit ID:** \`${id}\`
**Report:** \`${REPORT}\`
"

  if [[ "$DRY_RUN" == "1" ]]; then
    echo "DRY-RUN create: $issue_title"
    echo "  labels: $labels"
    CREATED=$((CREATED + 1))
    continue
  fi

  if [[ "$CREATE_ISSUES" == "1" ]]; then
    url="$(gh issue create --repo "$REPO" --title "$issue_title" --body "$full_body" --label "$labels")"
    echo "CREATED: $url"
    CREATED=$((CREATED + 1))
  fi
done < <(jq -c '.findings[]' "$REPORT")

echo "Done. created=$CREATED skipped=$SKIPPED"
