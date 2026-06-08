# RepoNerve MCP Server Troubleshooting Guide

This guide describes common error scenarios, root causes, and recovery procedures for running the RepoNerve MCP server.

---

## 1. Missing Database Error

### Symptom
When starting the MCP server via `reponerve mcp` (defined in [mcp.go](../../internal/cli/mcp/mcp.go)), the process exits immediately with an error:
```text
workspace not initialized; run 'reponerve init' first
```
Or if initialized but the SQLite file is missing:
```text
failed to open database: ...
```

### Cause
RepoNerve has not been initialized in the current repository or workspace directory, meaning no SQLite database exists yet.

### Resolution
1. Initialize the workspace:
   ```bash
   reponerve init
   ```
2. Scan the repository to ingest Git log events, ADRs, decisions, and build the sqlite database:
   ```bash
   reponerve scan
   ```
3. Restart the MCP server or the client using it.

---

## 2. Empty Repository or Context

### Symptom
Tool calls like `list_decisions` or `generate_context` succeed but return empty lists or empty JSON payloads.

### Cause
The scanner has run but did not extract any elements because the repository:
1. Has no commits.
2. Does not contain any markdown files matching Architecture Decision Record (ADR) naming conventions (e.g., in `docs/decisions/` or `doc/adr/`).
3. The extracted decisions, intents, facts, or events have not been linked yet.

### Resolution
1. Verify database state manually using the CLI:
   ```bash
   reponerve memory list decisions
   ```
2. If empty, ensure you have commits and status logs that meet the scanning rules.
3. Make sure to commit files or write ADR documents, and run `reponerve scan` again.

---

## 3. Invalid Repository IDs

### Symptom
When calling `generate_context` or `export_context`, you receive an empty result or error indicating the repository is not found or not registered.

### Cause
The `repository_id` passed to the tool does not match any active repository ID stored in the database.

### Resolution
1. Verify the current active workspace directory or target path.
2. Check the SQLite schema entries to ensure the repository has been indexed under the correct workspace ID.
3. You can retrieve the repository details by running:
   ```bash
   reponerve context generate
   ```
   from the command line in that directory to see the generated Repository ID.

---

## 4. MCP Startup Failures (Connection / STDIO Blockages)

### Symptom
AI client (e.g. Cursor, Claude Desktop) reports:
* `Failed to start MCP server: reponerve`
* Server connection timed out or process exited with status 1.

### Cause
The Model Context Protocol server operates exclusively over STDIO (standard input/output), orchestrated by [server.go](../../internal/mcp/server/server.go). Any unexpected logs, warnings, or standard output print statements that do not conform to the JSON-RPC 2.0 protocol will corrupt the stdout stream and cause the client to reject the handshake or disconnect.

### Resolution
1. **Ensure Clean Output**: Never write debug print statements directly to `os.Stdout`. In the CLI implementation of the MCP command, only the JSON-RPC transport stream handles `stdout`.
2. **Execute Locally**: Check that the `reponerve` command is globally accessible on your terminal `PATH`. In Claude Desktop or Cursor configuration, use the absolute path to the binary if it is not on the user `PATH` (e.g. `/usr/local/bin/reponerve`).
3. **Verify STDIO Loop**: Run the server manually in a terminal session:
   ```bash
   reponerve mcp
   ```
   Type the standard MCP initialization message. If the program terminates immediately, examine the printed error output on `stderr` to debug the underlying panic or crash.

---

## 5. Tool Execution Failures

### Symptom
Handshake completes successfully, but specific tools (e.g., `trace_decision`, `explain_event`, or `list_intents`) fail when invoked by the agent.

### Cause
This typically happens when:
* There are missing inputs or required parameters (e.g., calling `get_decision` without specifying `decision_id`).
* Internal database query fails due to SQLite locks or corruption.
* The query engine readers throw error responses.

### Resolution
1. Check the inputs sent by the agent in the agent client log file.
2. Verify that the requested ID (e.g., `decision_id`) exists in the database.
3. Ensure no parallel process holds an exclusive write lock on the SQLite file.
