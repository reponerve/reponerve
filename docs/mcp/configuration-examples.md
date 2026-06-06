# RepoNerve MCP Configuration Examples & Workflows

This document lists setup configuration templates and sample workflows for running RepoNerve MCP.

---

## Configuration Examples

### 1. Claude Desktop Config
Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"]
    }
  }
}
```

### 2. Cursor Settings
Configure via **Settings** -> **Features** -> **MCP**:
1. Click **+ Add New MCP Server**.
2. Name: `reponerve`.
3. Type: `command`.
4. Command: `/path/to/reponerve mcp`.

### 3. Windsurf
Add to `~/.codeium/windsurf/mcp.json`:
```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "enabled": true
    }
  }
}
```

### 4. Cline
Add to `~/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json`:
```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "disabled": false
    }
  }
}
```

### 5. Roo Code
Add to `~/Library/Application Support/Code/User/globalStorage/roocode.roo-cline/settings/roo_mcp_settings.json`:
```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "disabled": false
    }
  }
}
```

---

## Example Workflow Interactions

AI agents can execute standard repository intelligence workflows via these JSON-RPC calls:

### 1. List Decisions
Returns a list of all decisions:
* **Tool Name**: `list_decisions`
* **Arguments**:
  ```json
  {}
  ```

### 2. Trace a Decision
Traverse relationships to discover intents, supporting facts, and resulting events for a specific decision:
* **Tool Name**: `trace_decision`
* **Arguments**:
  ```json
  {
    "decision_id": "dec_1"
  }
  ```

### 3. Explain an Event
Obtain cause, intent, and decision trace for a specific event:
* **Tool Name**: `explain_event`
* **Arguments**:
  ```json
  {
    "event_id": "evt_1"
  }
  ```

### 4. Generate Repository Context
Retrieve structured repository briefing context:
* **Tool Name**: `generate_context`
* **Arguments**:
  ```json
  {
    "repository_id": "repo_xxx"
  }
  ```

### 5. Export Repository Context
Retrieve context formatted as rendered markdown string:
* **Tool Name**: `export_context`
* **Arguments**:
  ```json
  {
    "repository_id": "repo_xxx"
  }
  ```
