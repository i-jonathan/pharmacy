# T-131 Health readings history on the customer page

Card: T-131 · Branch: feat/T-131-health-readings-history

## Goal

Managers and admins can look back on a customer's blood pressure and glucose readings from the
Customers page, reprint any of them as a marked copy, and correct a wrong number with a reason.
Cashiers never see them. This finishes the health checks feature (piece 3 of
`docs/superpowers/specs/2026-11-19-health-checks-design.md`), so it also gets its feature docs and
its "What's new" announcement.

## What I found

- Card claims verified (`code`):
  - The API is complete from T-129: `GET /health-checks/customer/{id}?page=N` (20 per page, newest
    first, `{data, page, per_page, total_count}`), `POST /health-checks/{id}/reprint`,
    `PUT /health-checks/{id}`, gated by `healthcheck:view` / `healthcheck:correct` in
    `backend/adapter/http/router/health_checks.go:19-33` and again in
    `backend/service/health_checks.go`. No backend change is needed for the screen.
  - The correction replaces the whole reading; fields left out are cleared; it needs `version`
    (409 `changed`) and a `reason` of at most 200 characters (`service/health_checks.go:298-356`).
  - A correction that adds a blood pressure, pulse or glucose the reading lacked gets
    `customer_inactive` / `health_consent_needed` (`addsMeasurement`, `:362`). The correction body
    has no `consent_given`, so consent cannot be given from this form.
  - Every history load writes an audit entry `customer/health_history_viewed` (`:286`), and every
    reprint `health_check/reprinted` (`:487`).
  - Each reading carries `corrected` (`version > 1`) and `updated_at`
    (`repository/health_checks.go:17`), but not who corrected it or why (that is only in the
    audit log).
- The slip builder already takes `copy: { at }` and `slipFromSaved(reading, copy)` exists for the
  reprint (`frontend/src/next/lib/healthSlip.js:118`) (`code`).
- The reading checks (ranges, whole numbers, one decimal place for glucose) live inside
  `views/HealthChecks.vue` (`readings()`); the correction needs the same checks (`code`).
- `Customers.vue` is a list with a side panel holding the edit form; it asks before throwing
  typing away (`canLeavePanel`) (`code`). Purchase orders compose a preset reason and detail with
  `ReasonPicker` + `composeReason` (`views/PurchaseOrders.vue:1181`) (`code`).
- Health consent on the Customers form is its own card, T-141 (Up Next) (`code`: board).
- "What's new": newest entry is `v0.3.0`; tags go up to `v0.5.0`, and the customer features of
  v0.5.0 were never announced (T-125, Backlog). Only the newest release is shown, so whatever is
  not in the next entry is never announced (`code`: `lib/whatsNew.js`, `git tag`).
  `demoseed` makes customers but no readings, so the history picture needs a few seeded readings
  (`backend/internal/demoseed/demoseed.go:227`) (`code`).
- Owner's rule: commits and push are asked once, after implementation and review (`memory`).
- Web research skipped: a screen on an API already researched in T-129 (consent, NDPA).
- Nothing on the card is wrong.

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| Health readings on the customer panel, newest first, paged, hidden without `healthcheck:view` | card, design | yes |
| Load the history only when asked for (each load is audited) | `code` (audit per load) | yes |
| Reprint as a marked copy; print failure offers to try again | card, design | yes |
| Correct: every loaded field sent, version, reason; 409 `changed` reloads | card | yes |
| Correct that adds a measurement: `health_consent_needed` / `customer_inactive` explained | card | yes |
| Share the reading checks between the Health checks screen and the correction | `code` | yes |
| Feature docs `docs/features/health-checks/` (guide + technical), `feature-paths.json` | card, design | yes |
| "What's new" entry with pictures of the screen, slip and history; demo readings | card | yes (Decision 3) |
| Show who corrected a reading, when, why, and the old values | design ("trail") | Decision 2 |
| A reading saved for the wrong customer cannot be moved or removed | `model` | follow-up |
| Consent checkbox on the Customers form | T-141 | no (own card) |

## Decisions for you

1. **Where the history sits.** (a) A **Health readings** tab next to **Details** in the customer
   panel, loaded only when the tab is opened; (b) a section under the form, loaded with the panel.
   Recommendation: **(a)**. Every load is audited as "viewed health history", so it should happen
   only when someone means to look, not each time a manager opens a customer to fix a phone number.
   It also keeps the Save button of the details form where it is.
2. **Correction trail on screen.** (a) A "Corrected" mark with the date only; who, why and the old
   values stay in the audit log; (b) also list each correction under the reading (who, when, why,
   from → to), which needs a backend read of the audit log like Sales History's customer changes.
   Recommendation: **(a)** now, **(b)** as a follow-up card (P2). It keeps this card to the screen;
   the trail is already kept, only not shown.
3. **"What's new".** The next release is v0.6.0 (owner's call on the tag). Options: (a) health
   checks only (3 slides: recording screen, slip, history); (b) also the customer slides of T-125
   (customers list, customer at the till), since v0.5.0 was never announced and only the newest
   entry is ever shown. Recommendation: **(b)**: same capture run, two more slides, and T-125 is
   closed with it. Say the tag if not v0.6.0.

**Owner chose the recommendation on all three** (tab; mark only, trail as a follow-up; v0.6.0
with the customer slides).

## Design

### Customers page (`views/Customers.vue`)

- For a saved customer, the panel header gets two tabs, **Details** and **Health readings**; the
  second is shown only with `healthcheck:view`. A new customer has no tabs.
- Opening the tab loads page 1 (`GET /health-checks/customer/{id}`); switching customer resets it
  to Details; re-opening the tab for the same customer reuses what is loaded (a Refresh button
  reloads). Changing page loads that page.
- New component `components/customers/HealthReadings.vue` (keeps `Customers.vue` from growing by
  ~300 lines). Each reading: date and time, blood pressure `120 / 80 mmHg`, pulse, glucose
  `5.6 mmol/L (fasting)`, the note, "Taken by tunde", a "Corrected <date>" mark when
  `corrected`. No colour or word judges a number.
- Buttons per reading: **Reprint** (all with `healthcheck:view`), **Correct** (only with
  `healthcheck:correct`).
- Empty state: "No readings saved for this customer." Errors with Retry, as the list does.

### Reprint

`POST /health-checks/{id}/reprint` → print `slipFromSaved(reading, { at: new Date() })`
(`*** COPY <date> ***`). Button reads "Printing…" meanwhile. A failed print says so; pressing
Reprint again asks the server again (each print is recorded).

### Correct

- Opens an inline form in place of that reading, filled from what was loaded: blood pressure
  (top, bottom, pulse) and glucose (value, unit, when taken) each with its tick box as on the
  Health checks screen, the note, and a reason (`ReasonPicker` with: "Typed wrong", "Wrong unit",
  "Wrong time of glucose test", "Other"; composed as the purchase orders do; at most 200
  characters; detail required for Other).
- Checks before sending come from the shared `lib/healthReadings.js` (below). Nothing changed →
  "Nothing to correct" without sending (the server would also write nothing).
- Sends `PUT /health-checks/{id}` with every field (empty ones as null/""), `version`, `reason`.
  - 200 → the reading is replaced in the list with the reply; notice "Reading corrected."
  - 409 `changed` → "Someone corrected this reading first. Here is what is saved now; check it
    and correct again if needed." The list reloads and the form closes.
  - 409 `health_consent_needed` → "Adding a blood pressure or glucose this reading didn't have
    needs the customer's consent to keep their readings. Record it as a new health check instead,
    which asks for it." (The correction form can't ask for consent.)
  - 409 `customer_inactive` → "This customer is deactivated, so no new measurement can be added.
    You can still change the numbers already saved."
  - Anything else: the server's message.
- Leaving with an open, changed correction asks first (the panel's existing `canLeavePanel`, the
  tab switch, and the page's route leave).

### Shared checks: `lib/healthReadings.js`

`readingsFrom({ bpOn, systolic, diastolic, pulse, glucoseOn, glucose, glucoseUnit,
glucoseContext })` returns the body fields or throws an Error with the person's message: moved
out of `HealthChecks.vue` unchanged, which then imports it. Also `formatBloodPressure`,
`formatGlucose` for the list (the slip keeps its own formatting).

### What's new (Decision 3)

- `demoseed`: a few readings for one regular customer with health consent, spread over the last
  weeks (recorded through the service, then `taken_at` moved back in the throwaway database), one
  of them corrected.
- `capture.mjs`: slides for the Health checks screen (tunde, filled in), the slip preview, the
  history tab (amaka); and with (b) the Customers list and the till with a customer attached.
- `whatsNew.js`: a `v0.6.0` entry on top; health slides gated on `healthcheck:record` /
  `healthcheck:view`, customer slides on `customer:view`.

### Nothing else changes

No migration, no permission, no API change.

## Steps

1. `lib/healthReadings.js`; `HealthChecks.vue` uses it. Check: the Health checks screen still
   catches 1200/80, 120.5, 5.65, a missing context (browser).
2. `components/customers/HealthReadings.vue` list + paging + tabs in `Customers.vue`. Check: a
   manager sees the readings, a cashier sees no tab, one audit row per load.
3. Reprint. Check: copy slip text, audit row.
4. Correct, with every refusal. Check: each case below.
5. Feature docs (document-feature skill), `CLAUDE.md` line.
6. demoseed readings, `capture.mjs`, pictures, `whatsNew.js`; `make tailwind`.
7. `go test ./...` (with `PHARMACY_TEST_DSN` for demoseed) and the browser check.

## Testing

No JS test tooling, so the screen is driven in a browser (Playwright) on the demo database:
- amaka: tab appears; loads only when opened (one `health_history_viewed` audit row); paging past
  20; switching customer resets to Details; reprint prints a COPY slip and writes `reprinted`.
- Correct: change a number (audit has from/to); clear the pulse (sent as null, cleared); nothing
  changed (not sent); 409 `changed` (a second tab corrects first); add glucose to a BP-only
  reading for a customer without consent → message; for a deactivated customer → message; change
  numbers for a deactivated customer → saved; reason missing / Other without detail → caught.
- Leaving a changed correction asks first.
- tunde (cashier): no tab; `GET /health-checks/customer/1` answers 403.
- Health checks screen (T-130 checks on the shared validation): ranges, decimals, missing parts.
- `go test ./...`, with `PHARMACY_TEST_DSN` for `demoseed_test.go`.
- Not tested: a real 80mm printer.

## Deploy notes

`make assets` before building the image. No migration, no permission change.

## Commits

Asked once, after implementation and review (owner's standing rule). Planned: (1) shared reading
checks; (2) health readings tab, reprint, correction, rebuilt `tailwind.css`; (3) docs; (4) What's
new (demoseed, capture script, pictures, entry).

## Follow-up cards

- T-143 "Show the correction trail on a health reading" (P2, Customers & loyalty): who, when, why and
  the old values under a corrected reading, read from the audit log (Decision 2).
- T-144 "Move or void a health reading saved for the wrong customer" (P2, Customers & loyalty): today a
  correction can change the numbers but not the customer, and nothing removes a reading (T-142 is
  erasure on request, a different need).

## Sources

- Design: `docs/superpowers/specs/2026-11-19-health-checks-design.md`.
- T-129 / T-130 plans: `docs/plans/T-129-health-checks-backend.md`,
  `docs/plans/T-130-health-check-screen.md`.

## Changes during build

- Switching between Details and Health readings keeps an open correction (the tab is hidden, not
  closed), so the tab switch needs no question. Closing the panel, picking another customer,
  saving the details (which closes the panel) and leaving the page ask first; leaving the page
  also asks about unsaved details, which it did not before.
- `capture.mjs` now lists its slides per release (`node capture.mjs <folder> v0.6.0`); the v0.3.0
  steps are kept unchanged under their release.
- The pulse box in the correction form says "Optional" inside it, so the three boxes line up.
- Feature docs: `docs/features/health-checks/` (guide with three of the v0.6.0 pictures, and
  technical), the README row and `feature-paths.json`; the Customers docs gain the Health readings
  tab, and a stale answer there (who changed a sale's customer is "not on a screen yet") is fixed,
  since T-127 shipped it. Purpose as drafted from the design, approved by the owner.
- Browser check on the demo database: 38 checks, all passing (tab and lazy load with one audit row
  per load, paging, reprint copy and audit, correction with from/to audit, nothing changed, range,
  reason rules, 409 changed, consent and deactivated refusals and the allowed number changes,
  leave prompts, cashier has no tab and gets 403, the Health checks screen's checks on the shared
  rules).

## Review

`/code-review high` found seven; six fixed, one was the docs (written in this branch):
- Saving the details while a correction is open asks first (it closed the panel and lost the
  correction).
- Leaving, closing or picking another customer while a correction is saving says "Wait for the
  correction to finish saving." instead of doing nothing.
- A resend after a lost reply that gets 409 `changed` says the correction may already be saved,
  not that someone else corrected it.
- Retyping a number to its old value is no change (the form holds text, as the inputs do).
- The reason is counted in characters, as the server counts it.
- The tabs work from the keyboard (Left/Right; only the selected tab is in the Tab order).
Eight more browser checks for these; all 46 pass. `go test ./...` with `PHARMACY_TEST_DSN` passes.
`make check-go-format` and `check-go-lines` flag only files this branch does not touch.

Second review on the PR (`/code-review high`, seven; all addressed):
- A 502/503/504 or non-JSON 5xx on a correction counts as "may have saved", like a dropped
  connection, so a resend that gets 409 `changed` is not blamed on someone else; a 200 whose body
  is cut off reloads the list and says the reading was corrected; the reprint's reply is read
  safely.
- Deactivate/Reactivate and Save changes while a correction is saving say "Wait for the
  correction to finish saving." instead of closing the panel or doing nothing.
- (Reversed in the third review, below: closing a discarded correction at once lost it when the
  panel then stayed open.)
- The readings' messages use the customer's saved name, not a half-typed one on Details.
- The Health readings tab names its panel (`aria-controls`) only once the panel exists.
- demoseed says its audit entries keep the seed time (`audit_log` is append-only, so they can't be
  back-dated).
- `capture.mjs` says the slip step prints a walk-in's reading.
Eight more browser checks; all 54 pass.

Third review (`/code-review high`, six; all fixed):
- Agreeing to discard a correction only answers the question; the form closes when the panel
  really closes. A details save that then fails, or "Keep editing" on the details question, keeps
  the correction.
- Closing a correction whose save may have landed reloads the list, so the next correction isn't
  told someone else changed the reading.
- Any 5xx on a correction counts as "may have saved" (the app's own 500 can follow a commit whose
  acknowledgement was lost).
- Left/Right only move focus between the tabs; Enter or Space opens one, so an arrow key never
  records a health history view.
- The saved name is refreshed when the customer is reloaded after a 409.
- A demoseed comment points at `seedHealthChecks`.
Three checks changed to the new behaviour and three added; all 57 pass.

Fourth review (`/code-review high`, five; all fixed):
- Correct and Reprint are off while the list is loading, and a correction open on a reading that a
  new page no longer lists is closed (it had left Refresh disabled with no form shown).
- A saved correction drops any list reload still on its way, so a reload read before the save
  can't put the old version back (and with Correct off during a reload, the next correction can't
  start on a stale row).
- Reloading the page or closing the tab with unsaved details or a correction makes the browser ask
  (`beforeunload`), as Health checks and Stock Adjustments do; the Customers guide says so.
- Previous/Next while a correction saves says to wait instead of doing nothing.
- The Customers guide's "nobody can see that record yet (T-127)" and its "Planned next: T-127"
  are gone; T-127 shipped.
Four more browser checks; all 61 pass. The stale-reload race itself was not reproduced in the
browser: the checks cover the two guards that close it (buttons off while loading, reload dropped
on save).
