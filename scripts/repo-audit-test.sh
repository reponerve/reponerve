#!/usr/bin/env bash
# Focused regression tests for scripts/repo-audit.sh.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

STUBBIN="$TMPDIR/bin"
mkdir -p "$STUBBIN"

cat >"$STUBBIN/go" <<'STUB'
#!/usr/bin/env bash
case "$1" in
  vet|test)
    exit 0
    ;;
  *)
    echo "unexpected go invocation: $*" >&2
    exit 1
    ;;
esac
STUB

cat >"$STUBBIN/govulncheck" <<'STUB'
#!/usr/bin/env bash
cat <<'OUT'
govulncheck: loading packages:
There are errors with the provided package patterns:
/workspace/example.go:1:1: package requires newer Go version go1.26 (application built with go1.25)
OUT
exit 1
STUB

chmod +x "$STUBBIN/go" "$STUBBIN/govulncheck"

OUT="$TMPDIR/audit-report.json"
(
  cd "$ROOT"
  PATH="$STUBBIN:$PATH" REPONERVE_AGENT_FINDINGS="$TMPDIR/no-agent-findings.json" \
    ./scripts/repo-audit.sh "$OUT" >/dev/null
)

jq -e '.findings[] | select(.id == "govulncheck-execution" and .severity == "medium" and .category == "health")' "$OUT" >/dev/null
if jq -e '.findings[] | select(.id == "govulncheck" and .severity == "high" and .category == "security")' "$OUT" >/dev/null; then
  echo "govulncheck package-loading failure was misclassified as a vulnerability" >&2
  exit 1
fi

echo "repo-audit tests passed"
