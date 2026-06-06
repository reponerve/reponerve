# ISSUE-029 — Memory MCP Tools

## Objective

Expose Query Engine capabilities through MCP tools.

---

# Background

RepoNerve already supports:

* List
* Get
* Trace
* Explain

through the CLI.

These capabilities must now be exposed through MCP.

---

# MCP Tools

Implement:

list_decisions

get_decision

list_events

get_event

list_intents

get_intent

list_facts

get_fact

trace_decision

trace_event

explain_decision

explain_event

---

# Architecture

MCP Tool
↓
MCP Service
↓
Query Engine
↓
Memory Engine

---

# Constraints

Do NOT:

* Access SQLite directly
* Duplicate query logic
* Add AI features

Reuse existing Query Engine components.

---

# Unit Tests

Cover:

* Successful execution
* Missing entities
* Invalid arguments

---

# Acceptance Criteria

All Query Engine memory operations are accessible through MCP.
