# ISSUE-003

Title: Implement Init Command

Priority: P0

Status: Ready

---

## Description

Implement:

```bash
reponerve init
```

---

## Responsibilities

Create:

```text
.reponerve/
```

Create:

```text
.reponerve/config.yaml
```

Create:

```text
.reponerve/memory.db
```

Run migrations.

---

## Acceptance Criteria

Running:

```bash
reponerve init
```

Produces:

```text
✓ Workspace created

✓ Configuration created

✓ Database initialized

✓ RepoNerve ready
```

---

## References

- event-flows.md
- architecture-overview.md