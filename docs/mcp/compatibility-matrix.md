# RepoNerve MCP Compatibility Matrix

This document provides a compatibility matrix and integration details for connecting RepoNerve with various Model Context Protocol (MCP) clients and AI coding assistants.

---

## Compatibility Summary

| Client | Connection Mode | Discovery | Tool execution | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- |
| **Claude Code** | STDIO | Yes | Yes | Fully supported. Uses global or project-level config. |
| **Cursor** | STDIO | Yes | Yes | Configured via Settings UI. Requires absolute path to binary. |
| **Windsurf** | STDIO | Yes | Yes | Configured via `~/.codeium/windsurf/mcp.json`. |
| **Cline** | STDIO | Yes | Yes | Configured via `cline_mcp_settings.json`. |
| **Roo Code** | STDIO | Yes | Yes | Configured via `roo_mcp_settings.json`. |
| **Codex** | STDIO | Yes | Yes | Generic configuration via custom workspace JSON. |

---

## Client Details

### 1. Claude Code
* **Launch Command**: `reponerve mcp`
* **Supported Tools**: All 14 memory and context tools.
* **Known Limitations**: Requires `reponerve` binary to be installed on system `PATH` or specify the absolute executable path in config.

### 2. Cursor
* **Launch Command**: `reponerve mcp`
* **Supported Tools**: All 14 memory and context tools.
* **Known Limitations**: Execution environment must have access to `.reponerve` workspace directory. Standard error logs must be clean (we route stderr safely so as not to pollute JSON-RPC stdout communication).

### 3. Windsurf
* **Launch Command**: `reponerve mcp`
* **Supported Tools**: All 14 memory and context tools.
* **Known Limitations**: Uses custom configuration formats in Codeium app data directories.

### 4. Cline
* **Launch Command**: `reponerve mcp`
* **Supported Tools**: All 14 memory and context tools.
* **Known Limitations**: Standard input must remain dedicated to JSON-RPC streams.

### 5. Roo Code
* **Launch Command**: `reponerve mcp`
* **Supported Tools**: All 14 memory and context tools.
* **Known Limitations**: Needs correct directory permissions for SQLite database access.

### 6. Codex
* **Launch Command**: `reponerve mcp`
* **Supported Tools**: All 14 memory and context tools.
* **Known Limitations**: Standard schema limits apply.
