# Contributing to RepoNerve

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Welcome

Thank you for your interest in contributing to RepoNerve.

RepoNerve is the intelligence layer for software understanding.

Our mission is to preserve, organize, and transfer software knowledge so that understanding survives beyond individual contributors and remains accessible to both humans and AI systems.

We welcome contributions from:

* Software Engineers
* Open Source Maintainers
* Technical Writers
* Researchers
* AI Engineers
* Students
* Community Members

Whether you contribute code, documentation, design ideas, bug reports, or feedback, your contributions are valuable.

---

# Before You Contribute

Please read:

```text
/docs/vision/vision.md
/docs/vision/mission.md
/docs/vision/project-charter.md

/docs/product/problem-analysis.md

/docs/architecture/architecture-overview.md
```

These documents explain:

* Why RepoNerve exists
* What problems it solves
* Architectural direction
* Project principles

---

# Project Values

Every contribution should support the project's core values.

---

## Understanding First

RepoNerve exists to reduce the cost of software understanding.

Contributions should help humans and AI understand, change, review, and evolve software with less repeated repository exploration.

---

## Memory First

Repository memory is the technical foundation for knowledge preservation.

Features that improve memory quality are generally preferred over features that increase complexity.

---

## Evidence Over Assumptions

Repository memory should be traceable.

Whenever possible:

* Cite evidence
* Preserve sources
* Avoid unsupported conclusions

---

## Local First

Users should control their repository memory.

Avoid introducing unnecessary cloud dependencies.

---

## Simplicity Over Complexity

Prefer simple, understandable solutions.

Complexity must be justified.

---

## Open Source First

Decisions should prioritize community benefit.

---

# Ways To Contribute

---

## Report Bugs

Examples:

* Incorrect memory extraction
* Scan failures
* Query failures
* Performance issues

---

## Improve Documentation

Examples:

* Tutorials
* Examples
* Architecture diagrams
* Guides

---

## Implement Features

Examples:

* Parsers
* Memory extraction
* Query engine
* Context engine
* MCP integrations

---

## Improve Tests

Examples:

* Unit tests
* Integration tests
* End-to-end tests

---

## Research

Examples:

* Repository mining
* Knowledge extraction
* Context optimization
* AI token reduction

---

# Getting Started

---

## Fork Repository

```bash
git clone https://github.com/reponerve/reponerve.git
```

---

## Create Branch

Use descriptive branch names.

Examples:

```bash
feature/git-scanner

feature/decision-extraction

fix/memory-linking

docs/context-engine
```

---

## Keep Changes Focused

A pull request should solve a single problem.

Avoid combining unrelated changes.

---

# Development Environment

---

## Requirements

```text
Go 1.25+

Git

SQLite
```

---

## Setup

```bash
make setup
```

---

## Run Tests

```bash
make test
```

---

## Run Linter

```bash
make lint
```

---

## Build

```bash
make build
```

---

# Coding Standards

---

## Follow Go Standards

Use standard Go formatting.

```bash
gofmt
```

must always pass.

---

## Keep Packages Focused

Packages should have a single responsibility.

Bad:

```text
parser/
  ├── parser.go
  ├── scanner.go
  ├── storage.go
  ├── api.go
```

Good:

```text
parser/
scanner/
storage/
api/
```

---

## Favor Interfaces

Depend on abstractions.

Example:

```go
type MemoryStore interface {
    Save(ctx context.Context, memory Memory) error
}
```

---

## Avoid Global State

Prefer dependency injection.

---

## Keep Functions Small

Functions should be easy to test and understand.

---

# Testing Standards

All significant changes should include tests.

---

## Unit Tests

Required for:

* Parsers
* Extractors
* Query logic

---

## Integration Tests

Required for:

* Database operations
* Repository scanning
* Memory generation

---

## End-to-End Tests

Required for:

* CLI commands
* Full ingestion flows

---

# Documentation Standards

Documentation is treated as a first-class artifact.

---

## Documentation Updates

If a feature changes architecture or behavior:

Documentation should be updated in the same pull request.

---

## Markdown Guidelines

Use:

```markdown
# Heading

## Section

### Subsection
```

Keep documents clear and concise.

---

# Pull Request Process

---

## Step 1

Open an issue if one does not already exist.

---

## Step 2

Create a branch.

---

## Step 3

Implement changes.

---

## Step 4

Add tests.

---

## Step 5

Update documentation.

---

## Step 6

Open pull request.

---

# Pull Request Checklist

Before submitting:

* [ ] Code builds successfully
* [ ] Tests pass
* [ ] Documentation updated
* [ ] No unrelated changes included
* [ ] New functionality tested
* [ ] Commit messages are meaningful

---

# Commit Message Guidelines

Recommended format:

```text
type(scope): description
```

Examples:

```text
feat(scanner): add git commit parser

feat(memory): implement decision extraction

fix(query): handle empty search results

docs(architecture): update context engine design
```

---

# Design Proposals

Large changes should not begin with code.

Start with an RFC.

Examples:

* New memory type
* Storage redesign
* MCP protocol changes
* AI integration architecture

---

# Community Expectations

All contributors are expected to:

* Be respectful
* Be constructive
* Assume good intent
* Welcome newcomers
* Focus on ideas, not individuals

---

# Maintainer Responsibilities

Maintainers are responsible for:

* Protecting project direction
* Reviewing contributions
* Ensuring quality
* Preserving architectural consistency

---

# Decision Making

RepoNerve initially follows a Benevolent Dictator For Life (BDFL) governance model.

The project founder is responsible for:

* Product direction
* Architectural decisions
* Roadmap prioritization

Community input is strongly encouraged.

---

# Recognition

Contributors should be recognized for meaningful contributions.

Examples:

* Release notes
* Contributors page
* Documentation acknowledgments

---

# What Makes a Good Contribution

A good contribution:

* Solves a real problem
* Aligns with project goals
* Preserves simplicity
* Improves repository memory
* Includes tests and documentation

---

# What Makes a Great Contribution

A great contribution:

* Improves memory quality
* Improves context quality
* Reduces complexity
* Reduces AI token consumption
* Improves repository understanding

while remaining consistent with the project's vision.

---

# Guiding Principle

Every contribution should help developers and AI systems understand software systems without rediscovering knowledge that already exists.
