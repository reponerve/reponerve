# GitHub Copilot Chat Integration

RepoNerve can be integrated with GitHub Copilot Chat via the Model Context Protocol (MCP). This allows Copilot Chat to query repository memory, decisions, events, intents, facts, and more directly from RepoNerve.

## Prerequisites

- Visual Studio Code 1.99 or later
- GitHub Copilot extension installed
- RepoNerve CLI installed and available on your `PATH`

## Configuration

### Step 1: Ensure RepoNerve is Initialized

In the repository you want to analyze, make sure RepoNerve has been initialized and scanned:

```bash
reponerve init
reponerve scan
```

### Step 2: Configure MCP Server

The `.vscode/mcp.json` file in this repository already contains the MCP server configuration for RepoNerve:

```json
{
  "servers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "env": {
        "REPOERVE_WORKSPACE": "${workspaceFolder}"
      }
    }
  }
}
```

### Step 3: Start the MCP Server

1. Open the `.vscode/mcp.json` file in VS Code.
2. Click the **Start** button at the top of the file.
3. VS Code will discover the RepoNerve tools and make them available in Copilot Chat.

### Step 4: Use in Copilot Chat

1. Open Copilot Chat by clicking the Copilot icon in the title bar.
2. Select **Agent** from the dropdown menu.
3. Click the tools icon in the top left corner of the chat box to see available RepoNerve tools.
4. Ask questions like:
   - "What decisions have been made about the authentication system?"
   - "Who are the key contributors to this repository?"
   - "Explain the event related to database migration."
   - "List all facts about the API gateway."

## Available Tools

When connected, Copilot Chat can use the following RepoNerve tools:

- `list_decisions` - List all architectural decisions
- `get_decision` - Retrieve a specific decision
- `list_events` - List all repository events
- `get_event` - Retrieve a specific event
- `list_intents` - List all intents
- `get_intent` - Retrieve a specific intent
- `list_facts` - List all facts
- `get_fact` - Retrieve a specific fact
- `trace_decision` - Trace relationships for a decision
- `trace_event` - Trace relationships for an event
- `explain_decision` - Explain a decision
- `explain_event` - Explain an event
- `generate_context` - Generate repository context
- `export_context` - Export context as markdown
- `list_contributors` - List contributors
- `get_contributor` - Get contributor details
- `list_expertise` - List expertise records
- `trace_contributor` - Trace contributor activity
- `recommend_reviewers` - Recommend reviewers
- `discover_knowledge` - Discover repository knowledge
- `generate_learning_path` - Generate learning paths
- `generate_change_plan` - Generate change plans
- `trace_graph` - Traverse knowledge graph
- `trace_path` - Find paths between nodes
- `find_dependencies` - Find dependencies
- `find_dependents` - Find dependents
- `analyze_impact` - Analyze impact of changes

## Troubleshooting

### MCP Server Not Starting

- Ensure `reponerve` is on your PATH: `which reponerve`
- Check that the workspace is initialized: `reponerve init`
- Verify the database exists: `ls .reponerve/memory.db`

### Tools Not Appearing in Copilot Chat

- Make sure the MCP server is running (check the status in `.vscode/mcp.json`)
- Restart VS Code after starting the server
- Check the VS Code Output panel for MCP-related errors

### Empty Results

- Run `reponerve scan` to populate the database
- Verify the repository has commits and/or ADR files

## Further Reading

- [GitHub Docs: Extending Copilot Chat with MCP](https://docs.github.com/en/copilot/customizing-copilot/extending-copilot-chat-with-mcp)
- [VS Code Docs: Use MCP servers](https://aka.ms/vscode-add-mcp)
- [MCP Protocol Documentation](https://modelcontextprotocol.io/)
