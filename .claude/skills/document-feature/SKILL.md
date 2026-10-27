---
name: document-feature
description: Use when documenting a feature of the pharmacy app, writing or refreshing a staff guide or technical doc under docs/features/, explaining how a feature works for owners, cashiers or developers, or when a shipped change may have made a feature doc out of date.
---

# Document a feature

Internal skill for this repo. Every user-facing feature gets one folder under `docs/features/<slug>/` holding two documents about the same feature:

| Doc | Reader | Traces | Template |
|---|---|---|---|
| `guide.md` | Owners, managers, cashiers, storekeepers: no technical words | The person's journey: where to click, what they see, what happens next in the shop | `docs/features/templates/guide.md` |
| `technical.md` | Developers and AI agents | Each action from the Vue screen through the API, service, database and back, plus side effects | `docs/features/templates/technical.md` |

The templates are the source of truth for structure. Read the current template every time; don't work from memory. Keep its section order and headings; leave out a section that doesn't apply rather than leaving it empty.

## The rule that matters most

**The purpose comes from the owner.** Code shows what exists; only the owner knows why the feature exists, who uses it and how the shop actually works with it. For a new doc, if the request didn't include it, ask one question before reading any code:

> "Before I trace the code: in 2–3 sentences, what is `<feature>` for, who uses it day to day, and is there anything about how the shop works with it that the code won't show?"

Then wait. In update mode, the existing doc's purpose stands unless the owner says it changed.

## Workflow

1. **Get the purpose** (above).
2. **New or update?** `ls docs/features/`. A folder for this feature means update mode. For a new feature, propose the name and slug and confirm with the owner; one folder per feature as staff think of it ("Purchase orders", not "purchase_order_items").
3. **Which docs?** Both by default. The owner can ask for one.
4. **Trace.** For the guide, load `references/guide-tracing.md`; for the technical doc, load `references/technical-tracing.md`. Load them every time you trace, even if you remember them.
5. **Ask what code can't show** (one question each, skip if the owner skips): screenshots wanted?; typical volumes (sales a day, products, users); local practices (who approves, when stock is counted).
6. **Write** the docs from the templates. Diagrams: load `references/diagrams.md` before drawing any.
7. **Update `docs/features/README.md`**: one row per feature, in alphabetical order.
8. **Check** against the quality list below.
9. **Summarise** in five lines or fewer: what was traced, what was found wrong or missing, which board cards it links. Don't paste the docs.
10. **Commit only with the owner's yes.** Show the files (the feature folder and README only, staged by name) and the message `docs(features): add|refresh <feature> docs`. Never push from this skill.

## Update mode

- Read both existing docs first.
- Keep the owner-written parts word for word: purpose, "How the shop uses it", volumes, local practices, FAQs that came from the owner. Change them only if the owner says something changed.
- Re-trace and rewrite the code-derived parts. Report what changed: "Added: Reject step, 1 endpoint. Removed: 200-line limit. Unchanged: permissions."

## Rules every time

- **Every statement is checked in the code.** Screen labels, button text, limits, permissions, routes, statuses, table names. Other docs, cards and memories are leads, not facts; claims about what the system *can't* do go stale fastest. If a source was wrong, say so and offer to fix it.
- **Cite stable anchors** (`service/sales.go` `CreateSale`), never line numbers.
- **Known gaps link the board.** A limitation that has a card names its id (T-NNN); a limitation with no card is offered to the owner as a new card.
- **Money**: the guide shows naira (₦1,250.00); the technical doc says which unit each value is in at each layer (kobo in the database, naira floats on some API payloads).
- **No real data or secrets**: no customer names, phone numbers, real receipts, passwords, `.env` values. Use made-up examples.
- **Shipped work keeps docs current.** When a change touches a feature that already has docs, update them in the same branch. Creating docs for a feature that has none is the owner's call; offer, don't do.

## Quality checklist

- [ ] Purpose came from the owner (new docs).
- [ ] Every screen name, button label and message in the guide matches the `.vue` source exactly.
- [ ] Every role or permission named matches `internal/constant/permissions.go` and the route gating; ungated actions are stated as such in the technical doc.
- [ ] Every endpoint matches a route in `adapter/http/router/`; every table matches a migration.
- [ ] The guide has no technical words: run `grep -niwE 'endpoints?|api|database|kobo|json|backend|frontend|migrations?|sql|payload' docs/features/<slug>/guide.md` and reword any hit.
- [ ] Diagrams follow `references/diagrams.md` and pass its parser check.
- [ ] No `.go:NNN` or `.vue:NNN` citations: `grep -nE '\.(go|vue|js):[0-9]+' docs/features/<slug>/*.md`.
- [ ] Every known gap links a card or is offered as one.
- [ ] Empty sections removed; README row present.
