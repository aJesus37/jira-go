---
name: jira-go-epics
description: Epic management for jira-go — create epics, link tasks, create subtasks, view full hierarchy. Supports flat (Epic→Tasks) and 3-level (Epic→Tasks→Subtasks) models.
---

# jira-go — Epics

## Create an Epic

```bash
jira epic create --summary "Q3 Auth Overhaul" \
  --description "Replace legacy session system with OAuth2"
```

## List Epics

```bash
jira epic list
```

## View Epic with All Child Issues

```bash
jira epic view PROJ-1
```

## Add Existing Tasks to an Epic

```bash
jira epic add PROJ-1 PROJ-45 PROJ-46 PROJ-47
```

## Remove Tasks from an Epic

```bash
jira epic remove PROJ-45 PROJ-46
```

## Create Subtasks (3-level hierarchy)

```bash
# Parent must be a Task (not an Epic)
jira task create \
  --type Sub-task \
  --parent PROJ-45 \
  --summary "Write integration tests for OAuth callback"
```

## Full Epic Setup Workflow

```bash
# 1. Create the epic
jira epic create --summary "Payment Gateway v2"
# → Created PROJ-10

# 2. Create tasks (note the keys returned)
jira task create --summary "Design API contract" --assignee backend@example.com
# → PROJ-11
jira task create --summary "Implement Stripe integration" --assignee backend@example.com
# → PROJ-12
jira task create --summary "QA regression suite" --assignee qa@example.com
# → PROJ-13

# 3. Link all tasks to the epic
jira epic add PROJ-10 PROJ-11 PROJ-12 PROJ-13

# 4. Verify
jira epic view PROJ-10
```
