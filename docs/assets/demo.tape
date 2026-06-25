# RepoNerve README demo — setup & use (Steps 1–5)
# Regenerate from repo root: vhs docs/assets/demo.tape
Output docs/assets/reponerve-demo.gif

Set Width 1280
Set Height 720
Set FontSize 19
Set Theme "Catppuccin Mocha"
Set Padding 32
Set Framerate 24

Type@30ms "# RepoNerve — setup and use (exact steps)"
Enter
Sleep 1s

Type@30ms "# Step 1 — Install: verify binary"
Enter
Sleep 600ms
Type@30ms "reponerve --version"
Enter
Sleep 2s

Type@30ms "# Step 2 — Set up your repository"
Enter
Sleep 600ms
Type@30ms "reponerve init"
Enter
Sleep 2s

Hide
Type@30ms "reponerve scan"
Enter
Sleep 4s
Show
Sleep 1s

Type@30ms "# Step 3 — Verify setup"
Enter
Sleep 600ms
Type@30ms "reponerve doctor"
Enter
Sleep 3s

Type@30ms "# Step 4 — Understand the repository"
Enter
Sleep 600ms
Hide
Type@30ms "reponerve onboard --format compact --token-budget 380"
Enter
Sleep 8s
Show
Sleep 1s

Type@30ms "# Step 5 — Plan and ship a change"
Enter
Sleep 600ms
Hide
Type@30ms "reponerve plan 'Add webhook notifications' --format compact --token-budget 300"
Enter
Sleep 19s
Show
Sleep 1s

Hide
Type@30ms "reponerve reuse-check 'add webhook' --format compact --token-budget 300"
Enter
Sleep 7s
Show
Sleep 1s

Hide
Type@30ms "reponerve review 'webhook notifications' --format compact --token-budget 260"
Enter
Sleep 13s
Show
Sleep 1s

Type@30ms "# Step 6 — AI chat: restart IDE MCP, ask in natural language"
Enter
Sleep 2s
