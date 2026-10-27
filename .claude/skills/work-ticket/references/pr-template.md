# Ship a ticket

Load at the start of phase 6.

## Board, in the PR

1. Move the card from `BOARD.md` to the top of the current month in `DONE.md`: drop `priority` and `depends`, add `done:` (today) and `prs:`.
2. Add approved follow-up cards to the Backlog at their priority, using and bumping `next-id`.
3. Commit, push, open the PR, then set `prs:` to the PR number in a follow-up commit.

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
