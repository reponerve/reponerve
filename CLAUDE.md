## RepoNerve

RepoNerve provides evidence-backed repository context in **AI chat without MCP**.

Before answering questions about this codebase or making edits:

1. Ensure memory exists:
   `test -f .reponerve/memory.db || (reponerve init && reponerve scan)`
2. Load context (pick one):
   - `reponerve ask "<question>" --json`
   - `reponerve plan "<task>" --json` (pasted tickets)
   - `reponerve onboard --json` (day one)
   - `reponerve explain-function "<name>" --package <pkg> --json` (verify a fix / one symbol)
   - `reponerve explain-file "<path>" --json` (verify a file)
3. Read the JSON envelope in order: `structured` → `agent` → `formatted`
4. Answer and edit **only** from RepoNerve evidence. Do not grep the repo first.

For verification ("is this fix correct?"), prefer `explain-function` / `explain-struct` / `explain-file` over broad `ask` or full `plan` JSON.

Chat triggers: `/reponerve ask "..."` | Skill: `.cursor/skills/reponerve/SKILL.md`
