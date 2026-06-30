#!/usr/bin/env bash
# Regression tests for mechanical audit report classification.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMPDIR="$(mktemp -d)"
HAD_MEMORY=0

cleanup() {
  rm -rf "$TMPDIR"
  if [[ "$HAD_MEMORY" == "0" ]]; then
    rm -f "$ROOT/.reponerve/memory.db"
  fi
}
trap cleanup EXIT

mkdir -p "$TMPDIR/bin" "$ROOT/.reponerve"
if [[ -f "$ROOT/.reponerve/memory.db" ]]; then
  HAD_MEMORY=1
else
  touch "$ROOT/.reponerve/memory.db"
fi

cat >"$TMPDIR/bin/go" <<'EOF'
#!/usr/bin/env bash
case "${1:-}" in
  vet|test)
    exit 0
    ;;
  *)
    exit 0
    ;;
esac
EOF

cat >"$TMPDIR/bin/reponerve" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-}" == "doctor" ]]; then
  printf '{"structured":{"checks":[{"name":"memory","status":"ok","message":"fresh","evidence":{}}]}}\n'
fi
EOF

cat >"$TMPDIR/bin/govulncheck" <<'EOF'
#!/usr/bin/env bash
case "${REPONERVE_TEST_GOVULNCHECK_MODE:-}" in
  toolchain)
    printf 'govulncheck: loading packages:\n'
    printf '/workspace/internal/example.go:1:1: package requires newer Go version go1.26 (application built with go1.25)\n'
    exit 1
    ;;
  vulnerable)
    printf 'Vulnerability #1: GO-2026-9999\n'
    printf 'Example vulnerability\n'
    exit 1
    ;;
  *)
    exit 0
    ;;
esac
EOF

chmod +x "$TMPDIR/bin/go" "$TMPDIR/bin/reponerve" "$TMPDIR/bin/govulncheck"

PATH="$TMPDIR/bin:$PATH" REPONERVE_TEST_GOVULNCHECK_MODE=toolchain \
  "$ROOT/scripts/repo-audit.sh" "$TMPDIR/toolchain.json" >/dev/null

jq -e '
  .findings
  | any(
    .id == "govulncheck-execution-failed"
    and .severity == "medium"
    and .category == "health"
  )
' "$TMPDIR/toolchain.json" >/dev/null

jq -e '
  .findings
  | all(.id != "govulncheck")
' "$TMPDIR/toolchain.json" >/dev/null

PATH="$TMPDIR/bin:$PATH" REPONERVE_TEST_GOVULNCHECK_MODE=vulnerable \
  "$ROOT/scripts/repo-audit.sh" "$TMPDIR/vulnerable.json" >/dev/null

jq -e '
  .findings
  | any(
    .id == "govulncheck"
    and .severity == "high"
    and .category == "security"
  )
' "$TMPDIR/vulnerable.json" >/dev/null

echo "repo-audit classification tests passed"
