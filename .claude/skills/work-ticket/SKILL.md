---
name: work-ticket
description: Use when picking up, researching, planning, building, reviewing or shipping a card from BOARD.md in the pharmacy repo — a T-NNN ticket named by id or title, "the next ticket", or "what should I work on next".
---

# Work a ticket

Internal skill for this repo. One card → one branch → one plan → one PR. The owner works alone, so you also do the second person's job: question the card, review your own diff hard, and leave a PR they can approve in minutes.

The board format (columns, fields, ids, how to move and close cards) is defined in the "How this board works" section at the top of `BOARD.md`. Read it there; do not work from memory of it.

## Phases

Phases loop rather than run once: research can reopen during implementation, review sends you back to implement, and any change of scope sends you back to the plan for fresh approval.

| Phase | Done when | Load first |
|---|---|---|
| 1. Pick & claim | Card chosen, depends met, branch created, card moved to In Progress | — (below) |
| 2. Research | Card claims verified in code; gaps to "feature complete" listed with sources | `references/research.md` |
| 3. Plan | Plan file written and **approved by the owner in chat** | `references/plan-template.md` |
| 4. Implement | Plan steps done, tests green, nothing committed yet | — (below) |
| 5. Review | Review checklist passed, findings fixed or listed | `references/review-checklist.md` |
| 6. Ship | One approval for all commits, push and PR; PR open with the structured body, board updated in the PR | `references/pr-template.md` |

Load each reference when its phase starts, every time, even if you remember it. If you notice a phase ran without its reference, stop and load it.

## Pick & claim

1. Given an id or title, find that card. Given nothing, take the top card of In Progress, else the top card of Up Next whose `depends` are all in `DONE.md`. Show the card and say why it was chosen.
2. Cards that need the owner's hands (a deploy, a decision, a manual check) are not coded: present the checklist, help with each item, and update the card. Do not open a branch for them.
3. Stop and ask if the working tree is dirty, a dependency is not done, or the card is really several features (propose a split into new cards; work the first).
4. `git fetch origin`, then `git switch -c <feat|fix|chore>/T-NNN-short-slug origin/main` and `git branch --unset-upstream`.
5. Move the card to In Progress on that branch.

## Implement

- Follow `CLAUDE.md` and `AGENTS.md`. Money stays in kobo; permission checks go in both the route/service and the UI; confirmations use the shared dialog.
- **REQUIRED SUB-SKILL:** superpowers:test-driven-development for backend logic.
- Don't commit or ask about commits while implementing or reviewing; the owner approves them all at once when shipping (see Commits and pushes). Tick the card's checklist items as they land.
- Something the plan didn't cover? Small and obviously in scope: do it and note it for the PR. Otherwise stop, explain, and get approval before continuing.

## Commits and pushes (protected)

Committing and pushing are protected: nothing is committed or pushed without the owner's yes in chat. `.claude/settings.json` backs this up with `ask` rules for `git commit`, `git push` and `gh pr merge`, so Claude Code also shows a permission prompt. That prompt adds to the chat approval; it doesn't replace it.

**One approval, at the end.** Implement and review the whole ticket with the work uncommitted. Don't stop to ask about commits along the way; one-by-one approvals are tiring for the owner. When review has passed, show everything in a single message (the shipping bundle in `references/pr-template.md`):

- the proposed commits, in logical groups by default (e.g. migration + repository, then service, then UI, then docs and board), each with its files staged by name (never `git add -A` or `git add .`) and its message ending with the repo's attribution trailer
- the tests and checks run, with their results
- the push (branch and commits going up) and the PR title

One yes covers all of it. The owner can also approve the commits only, regroup them, or change anything. Run the approved commits in one shell command, so Claude Code's permission prompt appears once.

**After that**, a further change (another review round, a fix the owner asks for) gets one approval covering its commit and its push together.

Never push to `main` unless the owner says so. Never force-push unless the owner asks for that push.

## Rules that apply every time

- **No implementation before the owner approves the plan.** Approval means they said so in chat ("approved", "go", "ship it"). Silence, a question, or approval of an earlier version is not approval.
- Approving the plan is not approval to commit or push. An approved bundle covers exactly what it listed.
- A card's description can be stale or wrong. Every claim on it is checked against the code before the plan relies on it; wrong claims get fixed on the card.
- Board edits for the ticket ride in the ticket's PR, not on `main`.
- Never merge the PR, push to `main`, or deploy unless the owner says so in this conversation.
- Follow-ups found along the way become new cards (next id from the board's `next-id` comment), not silent scope creep.
