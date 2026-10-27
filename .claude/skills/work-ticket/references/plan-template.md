# Plan a ticket

Load at the start of phase 3, and again whenever scope changes after approval.

Write the plan to `docs/plans/T-NNN-short-slug.md` on the ticket branch. It is committed with the PR so the decision record lives next to the code. Then post a short summary in chat (goal, scope, the decisions that need the owner, risks) and wait for approval. Revisions edit the same file; say what changed.

## Template

```markdown
# T-NNN Title

Card: BOARD.md T-NNN · Branch: feat/T-NNN-slug

## Goal
One or two sentences: what changes for the user of the shop, and why now.

## What I found
- Verified on the card: claim — `path:line`
- Wrong or stale on the card: claim — what's actually true (card will be corrected)
- Relevant existing patterns: …

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| … | link or "my recommendation" | yes / follow-up |

## Decisions for you
For each: the options, my recommendation, and why. Only real choices; no questions the code or conventions already answer.

## Design
Data model and migrations (with rollback), API changes, permissions (keys and who gets them by default), audit entries, UI changes, offline or idempotency concerns.

## Steps
Numbered, each small enough to commit on its own, each naming its test.

## Testing
Unit tests, integration tests (`PHARMACY_TEST_DSN`), manual checks in the running app, and what will not be tested and why.

## Deploy notes
Migrations to run, permissions to grant, users who must log in again, `make assets`.

## Follow-up cards
Title, priority, feature, one line each. Added to the board after approval.

## Sources
Links with one line each.
```

## Approval

Only the owner's explicit yes in chat counts. Approval covers this plan as written: a new migration, permission, dependency, or user-visible behaviour not in the plan needs the plan updated and approved again.
