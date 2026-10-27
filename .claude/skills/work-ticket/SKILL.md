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
| 4. Implement | Plan steps done, tests green, commits made | — (below) |
| 5. Review | Review checklist passed, findings fixed or listed | `references/review-checklist.md` |
| 6. Ship | PR open with the structured body, board updated in the PR | `references/pr-template.md` |

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
- Commit only at the points the chosen commit style sets, and only after the owner confirms (see Commits and pushes). Tick the card's checklist items as they land.
- Something the plan didn't cover? Small and obviously in scope: do it and note it for the PR. Otherwise stop, explain, and get approval before continuing.

## Commits and pushes (protected)

Committing and pushing are protected actions: each one needs the owner's yes in chat for that specific commit or push. `.claude/settings.json` backs this up with `ask` rules for `git commit`, `git push` and `gh pr merge`, so Claude Code also shows a permission prompt. The prompt is in addition to the chat confirmation, not a replacement for it.

**Commit style.** Ask once per ticket, together with the plan approval (use the question tool if available), and record the answer in the plan's Commits section:

1. In logical groups (recommended): one commit per coherent change, e.g. migration + repository, then service, then UI
2. After every plan step
3. At the end of implementation, before review
4. One commit for the whole ticket, after review

**At each commit point**, show the owner:
- the files to be committed (stage them by name; never `git add -A` or `git add .`)
- the proposed message, ending with the repo's attribution trailer
- the tests run since the last commit and their result

Then wait. Yes → commit. Anything else → adjust and ask again. If the owner said "commit without asking" for this ticket, commits at the agreed points may go ahead, but you still say what you committed. Pushes never relax like this.

**Before every push**, show:
- the branch
- the commits that will go up (`git log --oneline origin/<branch>..HEAD`, or the full branch log on its first push)
- the review checklist result

Then wait for an explicit yes. Never push to `main`. Never force-push unless the owner asks for that push.

## Rules that apply every time

- **No implementation before the owner approves the plan.** Approval means they said so in chat ("approved", "go", "ship it"). Silence, a question, or approval of an earlier version is not approval.
- Approving the plan is not approval to commit or push. Approving one commit or push is not approval of the next.
- A card's description can be stale or wrong. Every claim on it is checked against the code before the plan relies on it; wrong claims get fixed on the card.
- Board edits for the ticket ride in the ticket's PR, not on `main`.
- Never merge the PR, push to `main`, or deploy unless the owner says so in this conversation.
- Follow-ups found along the way become new cards (next id from the board's `next-id` comment), not silent scope creep.
