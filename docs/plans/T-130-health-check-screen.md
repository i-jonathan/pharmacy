# T-130 Health checks screen and printed slip

Card: T-130 · Branch: feat/T-130-health-check-screen

## Goal

Staff record a blood pressure and/or glucose reading in a few taps and print it on the till's
80mm printer, replacing the handwritten note. For a customer the reading is saved (T-129's API);
for a walk-in it is printed and nothing reaches the server. Piece 2 of
`docs/superpowers/specs/2026-11-19-health-checks-design.md`.

## What I found

- Card claims verified (`code`):
  - `POST /health-checks/` needs a retry key, refuses 409 `health_consent_needed` unless
    `consent_given: true`, also `customer_inactive` and `already_saved`, and answers with the saved
    reading — `backend/service/health_checks.go` (T-129, merged as `d7f4260`).
  - `newRetryKey()` and `readError()` — `frontend/src/next/lib/api.js`.
  - Shared confirm dialog with `run` and an `alternate` action — `composables/useConfirm.js`.
- Printing: `lib/receipt.js` has `escapeHtml` and `printReceiptHtml(html)`, which prints any HTML
  document through a hidden frame and rejects when it can't. The slip reuses both (`code`).
- The till's `components/pos/CustomerPicker.vue` is a self-contained `v-model` picker: search by
  name or phone, deactivated customers shown and not pickable (Reactivate for `customer:manage`),
  add a new customer on the spot, and "Walk-in customer" when nothing is picked. It hides itself
  without `customer:view`. Reused as is (`code`).
- The signed-in user's name is `useUser().username` (`composables/usePermissions.js`,
  `components/TopNav.vue:79`) — for "Taken by" on a walk-in slip (`code`).
- Routes have no permission guard; each view gates itself (`views/StockAdjustments.vue:2`);
  the sidebar hides links with `hasPermission` (`components/Sidebar.vue:162`) (`code`).
- New Tailwind classes need `make tailwind` and the rebuilt `backend/template/static/css/tailwind.css`
  committed (`CLAUDE.md`) (`code`).
- Owner's rule: commits and push are asked once, after implementation and review, as one bundle
  (`memory`: confirm-before-commit-and-push).
- No JS test tooling in the repo (`CLAUDE.md`), so the screen is checked by driving it in a browser
  on the demo database.
- Web research skipped: the design was researched with T-129 (consent, NDPA), and this card is a
  screen on that API. Units (mmHg, bpm, mmol/L, mg/dL) are as in the design.
- Nothing on the card is wrong.

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| Screen: customer or walk-in, BP and/or glucose, Save and print | card, design | yes |
| Slip builder in `lib/healthSlip.js`, copy marking ready for T-131's reprint | design | yes |
| Consent asked at the counter on `health_consent_needed` | card, T-129 | yes |
| Customer declines consent: print without saving | `model` (unverified) — otherwise the person leaves with nothing | Decision 1 |
| Checks in the browser before sending (same ranges as the server) | design | yes |
| Print failure after a save: the reading stays saved, a manager reprints | design | yes |
| Sidebar link and route, hidden without `healthcheck:record` | design | yes |
| "What's new" entry with pictures | `CLAUDE.md` | Decision 2 |
| History, correct, reprint on the customer page | design | T-131 |

## Decisions for you

1. **Customer says no to keeping their readings.** Options: (a) the consent dialog's third button
   "Print only, don't save" prints their slip with their name and saves nothing, like a walk-in;
   (b) no such button: they are recorded as a walk-in by hand. Recommendation: **(a)**. It's the
   honest answer to "no" and costs one button.
2. **"What's new" announcement.** Options: (a) add it in this PR, with pictures of the screen and
   slip; (b) add it with T-131, once history and reprint exist and the feature is complete.
   Recommendation: **(b)**. **Owner chose (b)**; the line is added to T-131.

## Design

### Screen: `/health-checks`, "Health checks" in the Sales section of the sidebar

- Shown to users with `healthcheck:record`; the page itself says "You can't record health checks"
  otherwise.
- **Who**: the till's `CustomerPicker` (walk-in by default; search, add a new customer, deactivated
  ones marked). A user without `customer:view` sees only Walk-in. For a walk-in, an optional
  "Name on the slip" box.
- **Blood pressure** (on by default): systolic / diastolic, optional pulse.
  **Blood glucose** (off by default): value, unit (mmol/L default, mg/dL), when taken (Fasting /
  After a meal / Random). Each section has a checkbox; at least one must be ticked.
- **Note** (optional, kept with a customer's reading, not printed).
- Numbers are checked in the page with the server's ranges and rules (both BP numbers, top above
  bottom, one decimal place for glucose, unit and when-taken required) so a typo is caught before
  sending; the server's 400 message is the backstop. No message names a value as normal or high.
- **Save and print** (customer) / **Print** (walk-in). One retry key per form; a new one after a
  save or Clear. A failed send keeps the key, so pressing again can't save twice.

### Saving for a customer

1. `POST /health-checks/` with the readings and the key.
2. `201` → print the slip from the reply (customer name, time and "Taken by" from the server).
3. `409 health_consent_needed` → the shared confirm dialog: "May the pharmacy keep <name>'s health
   readings? Ask them first." Confirm "They agree: save and print" resends with
   `consent_given: true` (busy state and errors inside the dialog); alternate "Print only, don't
   save" (Decision 1); Cancel leaves the form as it was.
4. `409 customer_inactive` → the picker shows the customer as deactivated (Reactivate for those who
   may), the way the till does; Save is blocked while they are.
5. `409 already_saved` → this form was already saved (a reply lost on the way): the form is cleared
   and the message says to ask a manager to reprint it from the customer's page.
6. Printing fails after a save → "Saved, but the slip could not be printed. A manager can reprint it
   from the customer's page." The form clears either way, so it can't be saved twice.

### Walk-in

Nothing is sent. The slip is built from what was typed, the time on the till and the signed-in
user's name, then printed. A failed print keeps the form so it can be tried again.

### Slip: `frontend/src/next/lib/healthSlip.js`

`buildHealthSlipHtml({ name, takenAt, takenBy, systolic, diastolic, pulse, glucose, glucoseUnit,
glucoseContext, copy })`, printed with `printReceiptHtml`. Same 80mm layout as the receipt:

```
   Primocrest Pharmacy
   ---------------------
   Health check
   Name: Ada Obi            (or "Walk-in", or "Walk-in: <typed name>")
   24 Nov 2026, 10:42
   ---------------------
   Blood pressure
     120 / 80 mmHg
     Pulse 72 bpm
   Blood glucose
     5.6 mmol/L (fasting)
   ---------------------
   Taken by: tunde
   For your own records.
```

`copy: { at }` adds `*** COPY <date> ***` for T-131's reprint. Every value goes through
`escapeHtml`. No word interprets a number.

## Steps

1. `lib/healthSlip.js`: builder and `printHealthSlip`. Check: open the built HTML for a BP-only,
   glucose-only, both, walk-in-with-name and copy slip.
2. `views/HealthChecks.vue`: form, checks, walk-in print.
3. Customer save: consent dialog, refusals, print after save.
4. Route, sidebar link, `make tailwind`, `CLAUDE.md` line.
5. Run it on the demo database in a browser (below).

## Testing

No JS test tooling, so by hand on the demo database (`docs/whats-new/README.md` setup), driven in a
browser:
- As `tunde` (cashier): walk-in with and without a name; a customer without consent (the dialog:
  agree → saved and printed; print only → printed, nothing saved; cancel → nothing); a customer
  with consent; a deactivated customer; 1200/80 and a half-filled glucose caught in the page;
  a double click saves once; the slip's text for each case.
- As `amaka` (manager): the same screen works; the readings appear in the database and audit log.
- A user without `healthcheck:record`: no sidebar link, the page refuses.
- `go test ./...` still passes (no backend change expected).
- Not tested: a real 80mm printer (the slip is checked as HTML and in print preview).

## Deploy notes

`make assets` before building the image. No migration, no permission change (T-129 seeded them).

## Commits

Asked once, after implementation and review, with the push and PR (owner's standing rule).
Planned: (1) slip builder; (2) screen, route, sidebar, rebuilt `tailwind.css`, docs.

## Follow-up cards

None new. T-131's description gains the "What's new" announcement (Decision 2).

## Changes during build

- The customer picker sits outside the reading's `<form>`. Inside it, its buttons (which have no
  `type`) submitted the reading: Reactivate also re-sent the save and lost the reactivation, and
  its "add a customer" form would have nested. Found by the browser check; both cases are now in
  it.
- Browser check on the demo database (21 checks, all passing): walk-in with and without a name
  (nothing sent); 1200/80, 5.65 and a missing glucose context caught in the page; consent dialog
  cancel / print only / agree; consent on file; a double click saves once; deactivated customer
  blocks Save, Reactivate then saves; a customer added from the picker then saved; note stored;
  no link and a refusal without `healthcheck:record`.

## Review

`/code-review high` found eight; all fixed:
- Lost reply (the two important ones, and a third): once a save may have reached the server, the
  form locks and Save becomes "Try saving again", which resends exactly that request (same key,
  same consent answer), so the server answers with the first save or saves it once. A lost reply
  inside the consent dialog closes the dialog into the same state, so "Print only" is never offered
  after a save that may have landed.
- A failed print after a save offers "Print again" from the saved reading (the reprint on the
  customer's page is T-131, so the page no longer points there).
- The deactivated message is shown from the customer's state, so an edit can't hide it.
- A note is cleared when the customer changes.
- The till's picker takes `for-what` ("sale" by default, "reading" here) for its wording.
- The "When taken" radios share a `name`.
Browser check for these (12 more checks) plus the first 21 again: all pass.

Second review (`/code-review high`, five, all fixed): a 201 whose body is lost, and any 5xx (a
proxy may time out after the save committed), now count as "may have landed" like a dropped
connection; a reply clears the stale message; the retry carries the customer it was first sent
for; mmol/L glucose prints with one decimal place (5.0, not 5). Five more browser checks; all
pass, with the earlier ones.

## Sources

- Design: `docs/superpowers/specs/2026-11-19-health-checks-design.md` (screens and slip).
- T-129 plan: `docs/plans/T-129-health-checks-backend.md` (API, refusal codes, consent).
