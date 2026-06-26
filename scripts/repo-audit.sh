#!/usr/bin/env bash
# Run mechanical repository health and security checks; emit JSON findings.
# Usage: ./scripts/repo-audit.sh [output.json]
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

OUT="${1:-${REPONERVE_AUDIT_REPORT:-.reponerve/audit-report.json}}"
mkdir -p "$(dirname "$OUT")"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required (brew install jq / apt install jq)" >&2
  exit 1
fi

FINDINGS='[]'
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

add_finding() {
  local id="$1" severity="$2" category="$3" title="$4" body="$5" labels_json="$6"
  local entry
  entry="$(jq -n \
    --arg id "$id" \
    --arg severity "$severity" \
    --arg category "$category" \
    --arg title "$title" \
    --arg body "$body" \
    --argjson labels "$labels_json" \
    '{id: $id, severity: $severity, category: $category, title: $title, body: $body, labels: $labels, source: "repo-audit.sh"}')"
  FINDINGS="$(echo "$FINDINGS" | jq --argjson e "$entry" '. + [$e]')"
}

# --- RepoNerve doctor ---
if command -v reponerve >/dev/null 2>&1 && [[ -f .reponerve/memory.db ]]; then
  DOCTOR_JSON="$(reponerve doctor --json 2>/dev/null || true)"
  if [[ -n "$DOCTOR_JSON" ]]; then
    while IFS= read -r row; do
      name="$(echo "$row" | jq -r '.name')"
      status="$(echo "$row" | jq -r '.status')"
      msg="$(echo "$row" | jq -r '.message')"
      evidence="$(echo "$row" | jq -c '.evidence // {}')"
      [[ "$status" == "ok" ]] && continue
      sev="medium"
      [[ "$status" == "fail" ]] && sev="high"
      labels='["repo-audit","bug"]'
      [[ "$status" == "warn" ]] && labels='["repo-audit","enhancement"]'
      body="**Doctor check:** \`$name\` — $msg

**Status:** $status

**Evidence:**
\`\`\`json
$evidence
\`\`\`

**Suggested action:** Run \`reponerve doctor\` and follow recommendations (often \`reponerve scan\`).

_Automated finding from \`scripts/repo-audit.sh\`._"
      add_finding "doctor-${name}" "$sev" "health" "Doctor: $name — $msg" "$body" "$labels"
    done < <(echo "$DOCTOR_JSON" | jq -c '.structured.checks[]?')
  fi
else
  add_finding "doctor-skipped" "info" "health" "RepoNerve doctor not run" \
    "Run \`reponerve init && reponerve scan\` then re-run audit." '["repo-audit"]'
fi

# --- go vet ---
VET_LOG="$(mktemp)"
if go vet ./... >"$VET_LOG" 2>&1; then
  :
else
  body="**go vet** reported issues:

\`\`\`
$(head -c 12000 "$VET_LOG")
\`\`\`

_Automated finding from \`scripts/repo-audit.sh\`._"
  add_finding "go-vet" "medium" "quality" "go vet failures" "$body" '["repo-audit","bug"]'
fi
rm -f "$VET_LOG"

# --- go test ---
TEST_LOG="$(mktemp)"
if go test ./... -count=1 >"$TEST_LOG" 2>&1; then
  :
else
  body="**Tests failed:**

\`\`\`
$(tail -c 12000 "$TEST_LOG")
\`\`\`

_Automated finding from \`scripts/repo-audit.sh\`._"
  add_finding "go-test" "high" "test" "Test failures in CI/local audit" "$body" '["repo-audit","bug"]'
fi
rm -f "$TEST_LOG"

# --- govulncheck ---
VULN_LOG="$(mktemp)"
if command -v govulncheck >/dev/null 2>&1; then
  MODULE_GO_VERSION="$(go list -m -f '{{.GoVersion}}')"
  if GOTOOLCHAIN="go${MODULE_GO_VERSION}" go run golang.org/x/vuln/cmd/govulncheck@latest ./... >"$VULN_LOG" 2>&1; then
    :
  else
    body="**govulncheck** reported vulnerabilities:

\`\`\`
$(head -c 12000 "$VULN_LOG")
\`\`\`

Review at https://go.dev/security/vuln/ and upgrade affected modules.

_Automated finding from \`scripts/repo-audit.sh\`._"
    add_finding "govulncheck" "high" "security" "Go vulnerability scan findings" "$body" '["repo-audit","bug"]'
  fi
else
  add_finding "govulncheck-missing" "low" "security" "Install govulncheck for vulnerability scanning" \
    "Run: \`go install golang.org/x/vuln/cmd/govulncheck@latest\`" '["repo-audit","enhancement"]'
fi
rm -f "$VULN_LOG"

# --- Merge optional agent findings ---
AGENT_FILE="${REPONERVE_AGENT_FINDINGS:-.reponerve/audit-findings-agent.json}"
if [[ -f "$AGENT_FILE" ]]; then
  AGENT_FINDINGS="$(jq -c '.findings // []' "$AGENT_FILE" 2>/dev/null || echo '[]')"
  FINDINGS="$(echo "$FINDINGS" "$AGENT_FINDINGS" | jq -s 'add')"
fi

jq -n \
  --arg generated_at "$GENERATED_AT" \
  --arg repo "reponerve/reponerve" \
  --argjson findings "$FINDINGS" \
  '{generated_at: $generated_at, repository: $repo, findings: $findings}' >"$OUT"

echo "Wrote $OUT ($(echo "$FINDINGS" | jq 'length') findings)"
echo "$OUT"
