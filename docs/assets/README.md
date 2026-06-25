# RepoNerve assets

## Demo GIF (`reponerve-demo.gif`)

The README demo follows the **Setup and use (exact steps)** section — install → init → scan → doctor → understand → plan → reuse → review.

| File | Purpose |
| --- | --- |
| `reponerve-demo.gif` | README demo (1280×720) |
| `demo.tape` | [VHS](https://github.com/charmbracelet/vhs) script |
| `reponerve-demo-placeholder.svg` | Fallback if GIF is removed |

### Regenerate

Prerequisites: `ttyd`, `vhs`, `reponerve` on `PATH`, repository with git history.

```bash
brew install ttyd
go install github.com/charmbracelet/vhs@latest

# From repository root (after reponerve scan)
vhs docs/assets/demo.tape
```

### Demo script (matches README)

```bash
# Step 1 — Install (verify)
reponerve --version

# Step 2 — Set up repository
reponerve init
reponerve scan

# Step 3 — Verify
reponerve doctor

# Step 4 — Understand
reponerve onboard --format compact --token-budget 400

# Step 5 — Plan and ship a change
reponerve plan "Add webhook notifications" --format compact --token-budget 320
reponerve reuse-check "add webhook" --format compact --token-budget 320
reponerve review "webhook notifications" --format compact --token-budget 280

# Step 6 — AI chat: restart IDE MCP after init, then ask naturally
```

Slow commands use VHS `Hide`/`Show` while waiting for real CLI output.
