# RFC-006: npm Distribution

Status: Accepted  
Date: 2026-06-24

Related:

* `docs/install.md`
* `scripts/install.sh`
* `.goreleaser.yaml`

---

## Problem

JavaScript/TypeScript teams and monorepos expect `npm install` or `npx`. The shell install script and `go install` path exclude many non-Go users who already have Node.

## Decision

Publish an npm package **`reponerve`** at `npm/reponerve/` that:

1. Runs `postinstall` to download the matching GitHub Release archive (same artifacts as GoReleaser).
2. Exposes `reponerve` via `package.json` `bin` → `bin/reponerve.js` (spawns native binary).
3. Publishes to npm registry on git tag via GitHub Actions (`NPM_TOKEN` secret).

No Node port of RepoNerve. npm is an **install channel only**.

---

## Package layout

```text
npm/reponerve/
  package.json
  postinstall.js
  bin/reponerve.js
  lib/platform.js
  lib/install.js
  vendor/          # created at install time (gitignored)
```

Version in `package.json` tracks the GitHub Release binary set (`reponerve.version` field mirrors tag).

---

## Environment variables

| Variable | Purpose |
| --- | --- |
| `REPONERVE_SKIP_POSTINSTALL=1` | Skip download (CI smoke, offline) |
| `REPONERVE_FORCE_POSTINSTALL=1` | Re-download even if vendor exists |
| `REPONERVE_VERSION` | Override binary version (default: package version) |

---

## CI

Workflow `.github/workflows/npm-publish.yml`:

- Trigger: `push` tags `v*`
- Set `npm/reponerve` version from tag
- `npm publish --access public` with `NODE_AUTH_TOKEN`

Maintainers must configure `NPM_TOKEN` in GitHub repository secrets.

---

## Non-goals

- WASM or pure-JS implementation
- Bundling RepoNerve into application JS bundles
- Replacing `install.sh` or GitHub Releases

---

## Success criteria

| Verify |
| --- |
| `npm install` in `npm/reponerve` downloads binary and `npx reponerve --help` works |
| `docs/install.md` documents npm install |
| Tag push publishes to npm when `NPM_TOKEN` is set |

---

## v1.3.2 bundle (planned)

| Item | Surface |
| --- | --- |
| npm package | `npm install -g reponerve` |
| Docs | `docs/install.md` |
