# Research a ticket

Load at the start of phase 2, and again whenever implementation raises a question research should answer.

The output of research is two lists that go straight into the plan: **verified facts** (each with where it was checked) and **gaps** (what the feature needs to be complete, each marked in-scope or follow-up).

## 1. Code

- Read the card, its `depends` and `merged` cards, and any card it mentions.
- For every claim on the card (a file, a function, "X is not checked", "Y is hard-coded"), find the code and confirm it. Record `path:line` for each. A claim you can't confirm is wrong until shown otherwise; say so in the plan and fix the card.
- Trace the request path the change touches: router → controller → service → repository → migration → Vue view/composable. Note existing patterns to copy (e.g. how purchase orders do versioning, audit entries and permissions).
- Look for the same concept elsewhere (other screens, other services) so the change is made everywhere it applies, not in one copy.
- Check `DONE.md` and git history for earlier work on the same area.

## 2. The web

Search when the card is a product feature, touches security or money, or involves regulation or a library choice. Skip it for purely internal refactors and say you skipped it.

Look for, in this order:

1. **What complete looks like**: how established retail and pharmacy POS / inventory products handle this feature, which options and edge cases they cover, and what users complain about when it's missing.
2. **Rules that apply**: Nigerian pharmacy and retail rules where relevant (PCN, NAFDAC, FIRS/tax), data protection (NDPA), and security standards (OWASP ASVS / cheat sheets) for anything touching auth, sessions, money or personal data.
3. **How to build it**: library docs and well-regarded write-ups for the approach, checked against this stack (Go stdlib `net/http`, sqlx, Postgres, Vue 3, shadcn-vue).

Record every source as a link with one line on what it supports. Keep facts (with a source) separate from your recommendations (no source needed, labelled as yours). If sources disagree or you can't find one for a claim, say so; never fill the gap with a confident guess.

## 3. Turn findings into gaps

For each gap: one line, and **in scope** (needed for this card to be useful and small enough to fit) or **follow-up** (worth doing, separate card). Prefer a smaller complete feature over a large half-done one. Follow-ups are proposed in the plan and added to the board only after the owner approves.
