#!/usr/bin/env bash
# Fetch open GitHub issues as JSON for Cursor product-triage workflow.
set -euo pipefail

REPO="${REPONERVE_GITHUB_REPO:-reponerve/reponerve}"
LIMIT="${1:-50}"

if ! command -v gh >/dev/null 2>&1; then
  echo '{"error":"gh CLI not found; install GitHub CLI and run gh auth login"}' >&2
  exit 1
fi

gh issue list \
  --repo "$REPO" \
  --state open \
  --limit "$LIMIT" \
  --json number,title,labels,body,url,createdAt,author
