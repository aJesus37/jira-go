# jira-ticket → jira-go Skills Merge — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Absorb all valuable knowledge from the `jira-ticket` skill into the `jira-go-*` skill family, remove all project-specific references so the skills are publicly usable, and delete the now-redundant `jira-ticket` skill.

**Architecture:** Pure documentation changes — no code, no binaries. Each task edits one skill file or deletes one file. The content from `jira-ticket` is decomposed and routed to the most semantically appropriate `jira-go-*` skill. No SWCSIRT-specific URLs, project keys, custom fields, or language-specific strings survive.

**Tech Stack:** Markdown skill files, `jira` CLI binary (jira-go), bash examples.

---

## Files Involved

| File | Action |
|---|---|
| `~/.agents/skills/jira-go/jira-go/SKILL.md` | Modify — add reporting standard section |
| `~/.agents/skills/jira-go/jira-go-tasks/SKILL.md` | Modify — add 5 new sections |
| `~/.agents/skills/jira-go/jira-go-sprints/SKILL.md` | Modify — add sprint intent inference section |
| `~/.claude/skills/jira-ticket/SKILL.md` | Delete |
| `~/.claude/skills/jira-ticket/evals/evals.json` | Delete |

---

## Task 1: Add Ticket Creation Reporting Standard to `jira-go` entry point

**Files:**
- Modify: `~/.agents/skills/jira-go/jira-go/SKILL.md` after line 12 (end of `## Prerequisites`)

**What to add** — insert new section between `## Prerequisites` and `## Route to Sub-Skills`:

```markdown
## Ticket Creation Reporting Standard

After creating any ticket, always report:
- The ticket key (e.g., `PROJ-123`)
- The ticket URL: construct as `<jira_base_url>/browse/<KEY>` using the base URL from `~/.config/jira-go/config.yaml`
- Where it was placed: sprint name, or "backlog"
- One-line summary of what was created

Tell the user they can ask to update, add subtasks, assign, or move the ticket between sprint/backlog.
```

**Verify:** Read the file after editing. Confirm:
- No mention of `ifood`, `SWCSIRT`, or any hard-coded URL
- Section appears between `## Prerequisites` and `## Route to Sub-Skills`
- Existing content is untouched

---

## Task 2: Add Structured Ticket Description Template to `jira-go-tasks`

**Files:**
- Modify: `~/.agents/skills/jira-go/jira-go-tasks/SKILL.md` — append after line 100 (end of file)

**What to add:**

~~~markdown
## Structured Ticket Description Template

When creating tickets that require clear intent and acceptance criteria, use this description structure:

```
**Goal**
<what this task achieves and why it matters>

--------

**Definition of Done**
• <concrete, verifiable criterion>
• <another criterion>

--------

**Tech Inputs**
<tools, systems, links, docs involved — or "N/A" if none>
```

Build the body as a variable to safely handle newlines:

```bash
BODY="**Goal**
<goal text>

--------

**Definition of Done**
• <criterion 1>
• <criterion 2>

--------

**Tech Inputs**
<tech inputs or N/A>"

jira task create \
  --summary "<summary>" \
  --description "$BODY" \
  --no-interactive
```
~~~

**Verify:** Confirm the template block renders correctly, no project-specific references.

---

## Task 3: Add Info-Gathering Discipline to `jira-go-tasks`

**Files:**
- Modify: `~/.agents/skills/jira-go/jira-go-tasks/SKILL.md` — append after Task 2's addition

**What to add:**

```markdown
## Gathering Info Before Creating

Extract from user input before running any command:

- **Summary**: concise, action-oriented title
- **Goal**: the "why" — what problem does this solve or what outcome does it enable?
- **Definition of Done**: at least one concrete, verifiable criterion
- **Tech Inputs**: tools, systems, APIs, or links involved — or "N/A"
- **Assignee / Owners**: only if the user explicitly names someone
- **Parent ticket**: if the user mentions an existing ticket key → use `--type Sub-task --parent PROJ-XXX`

Ask only what is missing, one question at a time. Do NOT ask if the answer is clearly inferable from context. Do NOT ask all questions at once.

Example clarifying questions:
- Goal is vague: "What outcome does this achieve? What breaks if it isn't done?"
- No DoD: "How will we know this is complete? What can we verify?"
- Tech Inputs unclear but likely exist: "Are there specific tools, APIs, or systems this touches?"
```

**Verify:** No mention of `SWCSIRT`, `Tarefa`, `Subtarefa`, `ifood`, or any custom field IDs.

---

## Task 4: Add Sprint Placement After Creation to `jira-go-tasks`

**Files:**
- Modify: `~/.agents/skills/jira-go/jira-go-tasks/SKILL.md` — append after Task 3's addition

**What to add:**

```markdown
## Sprint Placement After Creation

After creating a ticket, move it to the appropriate sprint (see `jira-go-sprints` for intent inference):

```bash
jira sprint move active PROJ-XXX    # default — active sprint
jira sprint move future PROJ-XXX    # user said "next sprint" / "future sprint"
# omit entirely if user said "backlog" / "for later" / "not this sprint"
```

If no active sprint exists, leave the ticket in the backlog and inform the user.
```

**Verify:** Uses generic `PROJ-XXX` placeholder, correctly cross-references `jira-go-sprints`.

---

## Task 5: Add Unplanned Tasks and Common Mistakes to `jira-go-tasks`

**Files:**
- Modify: `~/.agents/skills/jira-go/jira-go-tasks/SKILL.md` — append after Task 4's addition

**What to add:**

```markdown
## Handling Unplanned Tasks

For ad hoc requests, incidents, or tasks raised mid-conversation, still apply the full template. The absence of prior planning makes the Goal and DoD *more* important, not less. If the user is in a hurry, create with a minimal DoD and flag it for refinement — but always capture at least one verifiable criterion and a real Goal.

## Common Mistakes to Avoid

- **Don't create a ticket without a Goal** — a title alone has no value.
- **Don't over-ask** — if the goal is evident from context, don't ask for it.
- **Don't invent assignees** — only set `--assignee` or `--owners` if explicitly named by the user.
- **Don't default to backlog** — unless the user signals otherwise, place in the active sprint.
- **Retro action items are an exception** — always place in backlog (they belong to a future sprint).
- **Match the user's language** — if the user wrote in a non-English language, use that language for the summary. Template section headers (Goal, Definition of Done, Tech Inputs) stay in English.
```

**Verify:** No project-specific wording. Portuguese example (`próxima sprint`) is generalized to "non-English language".

---

## Task 6: Add Sprint Intent Inference to `jira-go-sprints`

**Files:**
- Modify: `~/.agents/skills/jira-go/jira-go-sprints/SKILL.md` — insert after line 7 (title `# jira-go — Sprints`), before `## Create a Sprint`

**What to add:**

```markdown
## Sprint Intent Inference

Map user language to the correct sprint target before running any `sprint move` command:

| User says | CLI target | Notes |
|---|---|---|
| nothing / "this sprint" / "current sprint" | `active` | Default behavior |
| "next sprint" / "future sprint" | `future` | First sprint in FUTURE state |
| "backlog" / "for later" / "not this sprint" | *(omit sprint move)* | Leave in backlog |
| Retro action items | *(omit sprint move)* | Always backlog — belongs to a future sprint |

If no active sprint exists (`jira sprint list --state active` returns empty), skip `sprint move` entirely and inform the user the ticket was created in the backlog.
```

**Verify:** Section appears before `## Create a Sprint`. No project-specific language. Existing content untouched.

---

## Task 7: Delete `jira-ticket` skill files

**Files:**
- Delete: `~/.claude/skills/jira-ticket/evals/evals.json`
- Delete: `~/.claude/skills/jira-ticket/SKILL.md`
- Delete empty dirs: `~/.claude/skills/jira-ticket/evals/` and `~/.claude/skills/jira-ticket/`

**Commands:**
```bash
rm ~/.claude/skills/jira-ticket/evals/evals.json
rmdir ~/.claude/skills/jira-ticket/evals
rm ~/.claude/skills/jira-ticket/SKILL.md
rmdir ~/.claude/skills/jira-ticket
```

**Verify:** `ls ~/.claude/skills/` no longer contains `jira-ticket/`.

---

## Final Verification Checklist

After all tasks complete, confirm:

- [ ] `jira-go/SKILL.md` has `## Ticket Creation Reporting Standard` with no hard-coded URLs or project keys
- [ ] `jira-go-tasks/SKILL.md` has all 5 new sections: Template, Gathering Info, Sprint Placement, Unplanned Tasks, Common Mistakes
- [ ] `jira-go-sprints/SKILL.md` has `## Sprint Intent Inference` as the first content section after the title
- [ ] `~/.claude/skills/jira-ticket/` directory no longer exists

Run this grep after completion — expected output is no matches:
```bash
rg "SWCSIRT|ifood|Tarefa|Subtarefa|customfield_26030|atlassian\.net" ~/.agents/skills/jira-go/
```
