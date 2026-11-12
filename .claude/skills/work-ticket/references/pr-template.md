# Ship a ticket

Load at the start of phase 6.

## Board and docs

The board lives outside the repo (Waypoint, project `Primocrest`), so nothing about it goes in the PR's files.

1. If the card's feature already has docs in `docs/features/`, refresh them with the `document-feature` skill (update mode) and include them in the PR. If it has none, offer to create them; don't create them unasked. A hook blocks `gh pr create` when a documented feature's code changed without its docs; if the owner confirms the docs are still accurate, record that with a commit trailer `Feature-Docs-Unchanged: <slug>` instead of editing them.
2. Tick the card's remaining checklist items (`waypoint check`) and add approved follow-up cards with `waypoint add` at their priority and group; the ids are assigned for you.
3. Ask the owner once to approve the shipping bundle. Show the commits that will be pushed and the PR title. The bundle is: push the branch and open the PR. On yes, do exactly that. Anything beyond it needs its own approval.
4. After the PR is open, record it on the card: `waypoint set <ID> prs=#N --project Primocrest`. Move the card to Done (`waypoint done <ID> --project Primocrest`) when the owner merges the PR, not before.

## PR

Title: `T-NNN: <card title>`. Body:

```markdown
## Summary
What changes for the people using the shop, in 2–4 bullets.

Card: T-NNN · Plan: docs/plans/T-NNN-slug.md

## Changes
- Backend: …
- Frontend: …
- Migrations: NNNNNN_name (what it does; rollback behaviour)
- Permissions: new keys and their default roles

## Before merging / deploying
- [ ] Migrations to run, permissions to grant, users who must log in again
- [ ] `make assets` before building the image

## Testing
Commands run and their results; manual checks in the app; what was not tested and why.

## Review
Code-review findings fixed; any left, with the reason.

## Follow-ups
New card ids and titles.

## Sources
Research links that shaped the design.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

Then stop. Merging is the owner's call; merge only when they say so in this conversation.
