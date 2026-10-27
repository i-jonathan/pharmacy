# Research a ticket

Load at the start of phase 2, and again whenever implementation raises a question research should answer.

The output of research is two lists that go straight into the plan: **findings** (each tagged with its source) and **gaps** (what the feature needs to be complete, each marked in-scope or follow-up).

## Sources and tags

Every finding carries one tag, and the tag goes into the plan with it:

| Tag | Meaning | Can the plan rely on it? |
|---|---|---|
| `code` | Checked in the repo, with `path:line` | Yes |
| `web` | Supported by a linked source | Yes |
| `memory` | From the owner's saved memories or earlier sessions | Yes for the owner's preferences and decisions; verify any file, function or fact it names in code first |
| `model` | From your own training knowledge | Only as a candidate or a recommendation, never as a fact the design depends on |

**Load-bearing claims need `code` or `web`.** Regulation, tax, money handling, security, and any library, API or version the design depends on must be confirmed in code or a linked source before the plan relies on it. Training knowledge is often out of date on exactly these. If you can't confirm one, put it under "Decisions for you" as an open question rather than stating it.

**When sources disagree, the stronger source wins:** code beats memory and model for how the system behaves; a current linked source beats model knowledge for the outside world. Note the correction in the plan.

## 1. Start from what you know

Before searching, write down from your own knowledge (tag `model`):
- what a complete version of this feature usually covers in retail and pharmacy software
- likely edge cases and failure modes
- candidate designs and their trade-offs
- what to verify, and where (which regulator, standard, library docs)

This is the checklist the code and web research confirm, correct and extend. It is not a finding until something backs it.

## 2. Memory

Read the owner's saved memories for this project (the `MEMORY.md` index and the files it points to) and anything recalled in context. Look for:
- preferences and working rules (how they want commits, reviews, UI, naming)
- earlier decisions about this area, and why they were made
- context the code can't show: how the shop actually works, what's been tried, what's planned

Memories reflect when they were written. Check any file, function or flag a memory names before relying on it. If the code and a memory disagree, trust the code and mention the stale memory to the owner.

## 3. Code

- Read the card, its `depends` and `merged` cards, and any card it mentions.
- For every claim on the card (a file, a function, "X is not checked", "Y is hard-coded"), find the code and confirm it. Record `path:line` for each. A claim you can't confirm is wrong until shown otherwise; say so in the plan and fix the card.
- Trace the request path the change touches: router → controller → service → repository → migration → Vue view/composable. Note existing patterns to copy (e.g. how purchase orders do versioning, audit entries and permissions).
- Look for the same concept elsewhere (other screens, other services) so the change is made everywhere it applies, not in one copy.
- Check `DONE.md` and git history for earlier work on the same area.

## 4. The web

Search when the card is a product feature, touches security or money, or involves regulation or a library choice. Skip it for purely internal refactors and say you skipped it. Use the step 1 list to aim the searches.

Look for, in this order:

1. **What complete looks like**: how established retail and pharmacy POS / inventory products handle this feature, which options and edge cases they cover, and what users complain about when it's missing.
2. **Rules that apply**: Nigerian pharmacy and retail rules where relevant (PCN, NAFDAC, FIRS/tax), data protection (NDPA), and security standards (OWASP ASVS / cheat sheets) for anything touching auth, sessions, money or personal data.
3. **How to build it**: library docs and well-regarded write-ups for the approach, checked against this stack (Go stdlib `net/http`, sqlx, Postgres, Vue 3, shadcn-vue).

Record every source as a link with one line on what it supports. If sources disagree or you can't find one for a claim, say so; never fill the gap with a confident guess.

## 5. Turn findings into gaps

For each gap: one line, its source tag, and **in scope** (needed for this card to be useful and small enough to fit) or **follow-up** (worth doing, separate card). Prefer a smaller complete feature over a large half-done one. Follow-ups are proposed in the plan and added to the board only after the owner approves.
