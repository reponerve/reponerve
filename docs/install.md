# Installing RepoNerve

No Go toolchain required for the recommended install path. You only need the `reponerve` binary on your `PATH`, then `init` and `scan` in each repository.

**Latest release:** [GitHub Releases](https://github.com/reponerve/reponerve/releases)

---

## npm (Node.js users)

If you already use Node 18+:

```bash
npm install -g reponerve
```

Per project (recommended for JS/TS repos):

```bash
npm install -D reponerve
npx reponerve init
npx reponerve scan
```

`package.json` scripts example:

```json
{
  "scripts": {
    "reponerve:scan": "reponerve scan",
    "reponerve:init": "reponerve init"
  },
  "devDependencies": {
    "reponerve": "^1.4.0"
  }
}
```

The npm package downloads the same prebuilt binary as GitHub Releases during `postinstall`. See `docs/rfc/RFC-006-npm-distribution.md`.

Skip download (e.g. CI): `REPONERVE_SKIP_POSTINSTALL=1 npm install`

---

## Install script (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/reponerve/reponerve/main/scripts/install.sh | bash
```

Options:

```bash
# Install a specific version
curl -fsSL https://raw.githubusercontent.com/reponerve/reponerve/main/scripts/install.sh | REPONERVE_VERSION=v1.3.1 bash

# Custom install directory (default: ~/.local/bin)
curl -fsSL https://raw.githubusercontent.com/reponerve/reponerve/main/scripts/install.sh | REPONERVE_INSTALL_DIR=/usr/local/bin bash

# Verify SHA256 checksum from release artifacts
curl -fsSL https://raw.githubusercontent.com/reponerve/reponerve/main/scripts/install.sh | REPONERVE_VERIFY=1 bash
```

Ensure the install directory is on your `PATH` (add to `~/.zshrc` or `~/.bashrc` if needed):

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Verify:

```bash
reponerve --help
```

---

## Manual: download a release archive

Pick the archive for your platform from [Releases](https://github.com/reponerve/reponerve/releases):

| Platform | Archive |
| --- | --- |
| macOS Apple Silicon | `reponerve_<version>_darwin_arm64.tar.gz` |
| macOS Intel | `reponerve_<version>_darwin_amd64.tar.gz` |
| Linux amd64 | `reponerve_<version>_linux_amd64.tar.gz` |
| Linux arm64 | `reponerve_<version>_linux_arm64.tar.gz` |
| Windows amd64 | `reponerve_<version>_windows_amd64.zip` |
| Windows arm64 | `reponerve_<version>_windows_arm64.zip` |

### macOS / Linux

```bash
VERSION=v1.3.1
curl -fsSL -o /tmp/reponerve.tgz \
  "https://github.com/reponerve/reponerve/releases/download/${VERSION}/reponerve_${VERSION}_darwin_arm64.tar.gz"
tar -xzf /tmp/reponerve.tgz -C /tmp
install -m 755 /tmp/reponerve ~/.local/bin/reponerve
```

### Windows (PowerShell)

Download the `.zip` from Releases, extract `reponerve.exe`, and add that folder to your user `PATH`.

Optional: verify with `reponerve_<version>_checksums.txt` from the same release page.

---

## After install (every user)

From the root of a git repository:

```bash
reponerve init    # workspace, MCP config, Cursor skill, discipline rules
reponerve scan    # build repository memory
```

Optional:

```bash
reponerve hook install   # refresh memory after each commit
reponerve integrate      # refresh IDE integration files
```

MCP hosts (Cursor, VS Code, …) should use:

```json
{
  "command": "reponerve",
  "args": ["mcp"],
  "env": {
    "REPONERVE_WORKSPACE": "${workspaceFolder}/.reponerve"
  }
}
```

See `docs/ai-chat-integration.md` and `docs/cursor-integration.md`.

---

## Go developers (optional)

If you already have Go 1.26+:

```bash
# From a clone of this repository
make install

# Or without cloning
go install github.com/reponerve/reponerve/cmd/reponerve@v1.4.0
```

Contributors typically use `make build` or `make install` from a local clone.

---

## Local graph UI

After `reponerve init` and `reponerve scan`:

```bash
reponerve explore --serve
# open http://127.0.0.1:8765/
```

Export static HTML:

```bash
reponerve explore -o reponerve-graph.html
```

See `docs/rfc/RFC-009-local-explore-ui.md`.

---

## Homebrew

Create a tap repository (e.g. `github.com/reponerve/homebrew-tap`) and copy `packaging/homebrew/reponerve.rb` to `Formula/reponerve.rb`. After each release, update checksums:

```bash
./scripts/homebrew-sha256.sh v1.4.0
```

Then:

```bash
brew tap reponerve/tap
brew install reponerve
```

Until the tap is published, use the install script or release archives.

---

## Troubleshooting

| Problem | Fix |
| --- | --- |
| `reponerve: command not found` | Add install dir to `PATH`; open a new shell |
| Download 404 | Set `REPONERVE_VERSION` to a release that lists archives on GitHub (e.g. `v1.2.0`) |
| MCP does not start | Run `reponerve init` in the project; set `REPONERVE_WORKSPACE` in MCP config |
| Old binary | Re-run install script with a newer `REPONERVE_VERSION` |

---

## Related

* `docs/ai-chat-integration.md`
* `docs/cursor-integration.md`
* `docs/releases/versioning.md`
