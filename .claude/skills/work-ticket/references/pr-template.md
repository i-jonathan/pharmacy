# Ship a ticket

Load at the start of phase 6.

## Board, in the PR

1. Move the card from `BOARD.md` to the top of the current month in `DONE.md`: drop `priority` and `depends`, add `done:` (today) and `prs:`.
2. If the card's feature already has docs in `docs/features/`, refresh them with the `document-feature` skill (update mode) and include them in the PR. If it has none, offer to create them; don't create them unasked. A hook blocks `gh pr create` when a documented feature's code changed without its docs; if the owner confirms the docs are still accurate, record that with a commit trailer `Feature-Docs-Unchanged: <slug>` instead of editing them.
3. Add approved follow-up cards to the Backlog at their priority, using and bumping `next-id`.
4. Ask the owner once to approve the shipping bundle. Show every proposed commit (code in logical groups, then docs and board; files and message for each), the tests and checks run, the branch, and the PR title. The bundle is: make those commits in one command, push the branch, open the PR, then commit and push the `prs:` number. On yes, do exactly that. Anything beyond it needs its own approval.

## PR

Title: `T-NNN: <card title>`. Body:

```markdown
## Summary
What changes for the people using the shop, in 2–4 bullets.

Card: T-NNN (moved to DONE.md) · Plan: docs/plans/T-NNN-slug.md

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
