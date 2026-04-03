# Jira CLI - Next Steps Implementation Plan

## Features to Implement

### 1. Sprint API (High Priority)
- [ ] List sprints (active, future, closed)
- [ ] Create new sprint
- [ ] Start sprint
- [ ] Complete sprint
- [ ] View sprint board (kanban view in TUI)
- [ ] Move issues between sprints

### 2. Epic Management (High Priority)
- [ ] List epics
- [ ] Create epic
- [ ] View epic with child issues
- [ ] Add issues to epic
- [ ] Remove issues from epic
- [ ] Show epic progress

### 3. Glow Markdown Rendering (Medium Priority)
- [ ] Integrate Glow for rendering issue descriptions
- [ ] Render ADF (Atlassian Document Format) as Markdown
- [ ] Add markdown preview in TUI

### 4. Advanced TUI - Kanban Board (Medium Priority)
- [ ] Kanban board view for sprints
- [ ] Drag-and-drop to change status
- [ ] Swimlane view by assignee
- [ ] Column customization

### 5. Full TUI Ceremonies (Medium Priority)
- [ ] Sprint Planning TUI
  - Backlog view
  - Drag issues to sprint
  - Story point estimation
  - Export planning notes
- [ ] Retrospective TUI
  - Three-column board (Went Well, Improve, Action Items)
  - Anonymous card submission
  - Voting system
  - Export action items
- [ ] Daily Standup TUI
  - Team checklist
  - Blocker highlighting
  - Timer
  - Export summary

### 6. Export Features (Low Priority)
- [ ] Export ceremonies to Markdown
- [ ] Export issues to Markdown
- [ ] Generate sprint reports

---

## Implementation Order

1. Sprint API (core functionality)
2. Epic Management (core functionality)
3. Glow Markdown (quick win for UX)
4. Kanban Board TUI (enhanced UX)
5. Ceremonies TUI (advanced features)
6. Export features (final polish)
