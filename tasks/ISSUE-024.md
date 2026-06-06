# ISSUE-024 — Context CLI Command

## Objective

Expose Repository Context generation through the CLI.

This command provides a single entry point for understanding repository knowledge.

---

# Background

The Context Generator already produces:

```go
RepositoryContext
```

This issue is responsible for:

* Executing the generator
* Rendering the result
* Displaying repository context to users

---

# Command

Create:

```bash
reponerve context generate
```

---

# Command Structure

Root:

```bash
reponerve context
```

Subcommand:

```bash
reponerve context generate
```

---

# Responsibilities

The command must:

1. Load the active workspace.
2. Open configured storage.
3. Create Query Engine readers.
4. Execute the Context Generator.
5. Render the generated RepositoryContext.

---

# Output

Initial output should be human-readable.

Example:

```text
Repository Context

Repository:
repo_xxx

Generated:
2026-06-06T12:00:00Z

Key Decisions

- Use Redis Cache
- Adopt gRPC

Key Intents

- Reduce Latency
- Improve Reliability

Key Facts

- Auth Service USES Redis
- API Gateway DEPENDS_ON Auth Service

Recent Events

- Introduce Redis Cache
- Refactor Authentication Flow
```

---

# Rendering Rules

Use deterministic rendering.

No AI.

No summarization.

No generated prose.

Only structured output.

---

# Package Structure

Recommended:

```text
cmd/reponerve/contextcmd/

context.go

generate.go
```

---

# Error Handling

Handle:

* Missing workspace
* Missing database
* Empty repository
* Generator failures

Gracefully.

---

# Empty Context

Display:

```text
No repository context available.
```

Return success.

Do not return an error.

---

# Unit Tests

Cover:

* Command registration
* Successful generation
* Empty context
* Missing workspace
* Generator failures

---

# Integration Tests

Verify:

CLI
↓
Context Generator
↓
Readers
↓
SQLite

End-to-end.

---

# Constraints

Do NOT implement:

* Markdown export
* AI summaries
* MCP integration
* Embeddings
* Ownership extraction

Only generate and display repository context.

---

# Acceptance Criteria

Users can run:

```bash
reponerve context generate
```

and receive a deterministic repository context generated from repository memories.

All tests pass.
