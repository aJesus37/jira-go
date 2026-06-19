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

Epics are created as issue type `Epic`. Use `--project KEY` to override the default project.

## List Epics

```bash
jira epic list
```

## View Epic with All Child Issues

```bash
jira epic view EPIC-10
```

Shows epic details, progress (% done), and a table of child issues.

## Link a Task to an Epic (two approaches)

**Approach A — One-step (create + link):**
```bash
jira task create --epic EPIC-10 --summary "Implement Stripe integration"
```

**Approach B — Two-step (create, then link existing):**
```bash
jira task create --summary "QA regression suite"   # → PROJ-13
jira epic add EPIC-10 PROJ-13
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
# → EPIC-10

# 2. Create tasks directly linked to epic
jira task create --epic EPIC-10 --summary "Design API contract" --assignee backend@example.com
# → PROJ-11
jira task create --epic EPIC-10 --summary "Implement Stripe integration" --assignee backend@example.com
# → PROJ-12
jira task create --epic EPIC-10 --summary "QA regression suite" --assignee qa@example.com
# → PROJ-13

# 3. Create subtasks under a task (still in same epic)
jira task create --type Sub-task --parent PROJ-11 --epic EPIC-10 --summary "API spec review"

# 4. Create an unlinked task then add it later
jira task create --summary "Write docs"
jira epic add EPIC-10 PROJ-14

# 5. Verify everything
jira epic view EPIC-10
```

## Configuration for Epic Linking

Epic linking works differently depending on the project type:

| Project Type | Field Used | Config Required |
|---|---|---|
| **Team-managed** (next-gen) | `parent` | None (works automatically) |
| **Company-managed** (classic) | Epic Link custom field | `epic_link_field: customfield_10014` in project config |

```yaml
# ~/.config/jira-go/config.yaml
projects:
  PROJ:
    jira_url: https://company.atlassian.net
    epic_link_field: customfield_10014  # only needed for company-managed
```

If `epic_link_field` is not set, the CLI uses the `parent` field (team-managed approach).
