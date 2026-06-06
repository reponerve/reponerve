# ISSUE-026 — Context Export

## Objective

Export generated Repository Context to reusable files.

This enables:

* AI agent bootstrapping
* Repository onboarding
* Documentation generation
* MCP integrations
* Context sharing

---

# Background

The Context Generator creates:

RepositoryContext

The Template Renderer creates:

Markdown

This issue is responsible for persisting rendered context.

---

# Commands

## Generate and Export

```bash
reponerve context export
```

---

## Export Markdown

```bash
reponerve context export \
  --format markdown
```

---

## Export To File

```bash
reponerve context export \
  --output repository-context.md
```

---

# Default Behaviour

When no output path is provided:

```bash
reponerve context export
```

write:

```text
./repository-context.md
```

---

# Initial Formats

Supported:

```text
markdown
```

Only.

Future versions may add:

```text
json
html
yaml
```

---

# Export Flow

Readers
↓
Generator
↓
RepositoryContext
↓
Renderer
↓
Markdown
↓
File

---

# Package Structure

Recommended:

```text
internal/context/export/

exporter.go
```

Responsibilities:

* Execute renderer
* Create destination file
* Write output safely

---

# Error Handling

Handle:

* Invalid output paths
* Permission failures
* Missing workspace
* Missing repository context

Gracefully.

---

# File Naming

Default:

```text
repository-context.md
```

Custom:

```bash
--output my-context.md
```

---

# Constraints

Do NOT implement:

* Cloud storage
* MCP transport
* HTTP APIs
* AI summarization
* Ownership extraction

Only file export.

---

# Unit Tests

Cover:

* Successful export
* Custom output paths
* Default output paths
* Permission failures
* Empty context

---

# Integration Tests

Verify:

Readers
↓
Generator
↓
Renderer
↓
Exporter
↓
File

End-to-end.

---

# Acceptance Criteria

Users can generate a repository context document and export it to a markdown file.

All tests pass.
