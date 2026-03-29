# Phase 12: Remaining Mockup Pages + Unmocked Pages - Context

**Gathered:** 2026-03-28
**Status:** Ready for planning
**Mode:** Auto-generated (autonomous mode)

<domain>
## Phase Boundary

Все оставшиеся 7 страниц с макетами pixel-perfect ��оответствуют HTML-макетам, а 6 страниц без макетов приведены к единому дизайн-языку — вся функциональность работает без регрессий.

Requirements: ISSU-01..04, CALL-01..04, TMSP-01..04, STNG-01..05, ADMN-01..05, ONBD-01..05, TRSH-01..03, XTRA-01..07

</domain>

<decisions>
## Implementation Decisions

### Pages WITH Mockups (7 pages)

#### Issue Detail (ISSU-01..04)
- Two-column layout: main content + right sidebar meta
- Description, comments, activity section restyled
- Right sidebar: status, assignee, deadline, related issues
- Mockup: редизайн/05-issue-detail.html

#### Calls (CALL-01..04)
- Participant grid restyled
- Call controls (mute, camera, share, end call) restyled
- Participant status indicators
- Mockup: редизайн/06-calls (1).html

#### TMS (TMSP-01..04)
- Test management interface restyled
- Test case grid/list with new card styling
- Execution status tracking
- Mockup: редизайн/07-tms (3).html

#### Settings (STNG-01..05)
- Sidebar navigation + content layout
- Profile section, appearance section, notification preferences
- Mockup: редизайн/08-settings.html

#### Admin (ADMN-01..05)
- KPI cards, tabbed interface, user management table, audit log
- Mockup: редизайн/09-admin.html

#### Onboarding (ONBD-01..05)
- Centered modal flow, progress indicator
- Step 1: profile form, Step 2: feature tour (6 cards), Step 3: theme + accent
- Mockup: редизайн/10-onboarding.html

#### Trash (TRSH-01..03)
- Table with deleted items, restore/delete buttons, type badges
- Mockup: редизайн/11-trash.html

### Pages WITHOUT Mockups (6 page groups — XTRA-01..07)
- Projects list page → use Dashboard widget grid patterns
- Project overview → use Dashboard + Card patterns
- Sprint management (backlog, sprint detail, burndown, velocity, planning poker) → use Kanban column + Card patterns
- JQL search → use Settings form + Card patterns
- Project settings (workflow, labels, integrations) → use Settings sidebar + form patterns
- Project dashboard (widget grid) → use Dashboard widget grid patterns

### Claude's Discretion
- Exact layout details for unmocked pages (compose from established patterns)
- Any CSS utilities needed for page-specific patterns
- Responsive adjustments

</decisions>

<code_context>
## Existing Code Insights

### Mockup Files
- редизайн/05-issue-detail.html
- редизайн/06-calls (1).html
- редизайн/07-tms (3).html
- редизай��/08-settings.html
- редизайн/09-admin.html
- редизайн/10-onboarding.html
- редизайн/11-trash.html

### Key Source Files
- Issue: frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx + modules/tracker/
- Calls: frontend/src/routes/_authenticated/calls.tsx + modules/calls/
- TMS: frontend/src/routes/_authenticated/projects/$projectId/tms/ + modules/tms/
- Settings: frontend/src/routes/_authenticated/settings/ + modules/settings/
- Admin: frontend/src/routes/_authenticated/admin/ + modules/admin/
- Onboarding: frontend/src/routes/onboarding.tsx
- Trash: frontend/src/routes/_authenticated/trash.tsx
- Projects: frontend/src/routes/_authenticated/projects/index.tsx
- Sprint: frontend/src/routes/_authenticated/projects/$projectId/sprints/

</code_context>

<specifics>
## Specific Ideas

- Only CSS class changes — NO modifications to event handlers, state, API calls
- Pages without mockups: apply design system tokens consistently (bg-base, text-t1/t2, border-b1, card patterns from Phase 10/11)

</specifics>

<deferred>
## Deferred Ideas

None.

</deferred>
