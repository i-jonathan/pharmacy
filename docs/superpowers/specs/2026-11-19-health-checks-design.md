# Health checks: blood pressure and blood glucose readings

Status: draft for the owner's approval · 2026-11-19

## Intent and why

Customers come to the pharmacy to have their blood pressure and/or blood glucose tested. Today the
staff write the result on paper with a biro: the readings, the date, sometimes the time, as the
customer's own record. The pharmacy keeps nothing, so nobody can look back at how a regular's
numbers have moved.

This feature replaces the handwritten note with a printed slip, and, for people who are customers,
keeps the readings as a history that managers can look back on.

## Goals

- A cashier records a reading in a few taps and prints a slip on the till's 80mm printer.
- For a person who is a customer (a record on the Customers page), the reading is saved and joins
  their history.
- Managers and Admins can see a customer's history and reprint any past reading.
- A wrong number is corrected on the record with a trail, never overwritten silently.

## Non-goals (for this version)

- The fee. Tests are paid, but the money stays out of this feature (see Approaches, "Fee").
- Any interpretation: no "normal", "high" or colour flags, on the slip or on screen. That is the
  pharmacist's call, not the till's.
- Other tests (weight, temperature, cholesterol, malaria strips).
- Trend charts, and "last time was 130/85" on the slip. The history list comes first.
- Sending a slip by SMS, WhatsApp or email; proper ESC/POS printer support (T-048).
- Offline use (T-047 covers offline generally).

## Anti-goals (a failure even if it technically works)

- A cashier seeing any customer's saved readings, or reprinting an old one.
- Health data kept for people who did not become customers: a walk-in's numbers are printed and
  then gone, from the server too.
- Health values in logs, error messages or URLs.

## Decisions made in conversation (the owner's)

- A person is either a **customer** (record on the Customers page, a name is enough) or a
  **walk-in**. A reading for a customer is saved. A walk-in reading is printed under "Walk-in"
  (with a name typed on the slip if the person wants one) and nothing is stored.
- Cashiers can record and print, but cannot see history. History and reprints are for managers and
  admins; the data is sensitive medical information.
- The readings are a "history" a manager can look back on, not only a reprint store.
- The fee stays out of the system for now. The shop currently sells tests as ordinary products with
  made-up stock; it may switch to non-stock service items later.

## Approaches considered

**Storage.** (A) One typed row per health check: the customer, time, who took it, BP (systolic,
diastolic, pulse) and glucose (value, unit, context). Validated, simple history, but a new test type
needs a small database change. (B) A generic measurement row (kind, value, unit), one per number:
flexible, but nothing validates the numbers and a BP pair spans two rows. (C) Free text in the
customer's notes: cannot be validated or permissioned separately. **Chosen: A.**

**Where it is entered.** (A) A standalone Health checks screen, which does not touch the till or
the sale flow. (B) Inside the till's cart, which only makes sense once the test is billed there.
**Chosen: A.**

**Fee.** (A) Keep it out for now; cashiers charge as they do today. (B) Add non-stock service items
to the till, then link a health check to its sale. (C) The health check screen takes payment
itself: rejected, because it puts money outside sales and payments. **Chosen: A now, B as its own
later card**; the link is one optional `sale_id`, so nothing here needs redoing.

## Design

### Data

New table `health_check` (migration 51, with a working down file):

- `id`, `customer_id` NOT NULL references `customer` (a walk-in is never stored).
- `taken_at` timestamptz, set by the server to now (not editable, so a record cannot be backdated).
- `taken_by_id` references `users`.
- BP: `systolic`, `diastolic`, `pulse` (all integers, all nullable). Systolic and diastolic come
  together or not at all; systolic must exceed diastolic.
- Glucose: `glucose` (numeric, nullable), `glucose_unit` (`mmol/L` or `mg/dL`), `glucose_context`
  (`fasting`, `after_meal`, `random`). Value, unit and context come together or not at all.
- A check that at least one of BP or glucose is present.
- `note` text, short, optional.
- `version` integer (changes on every correction), `created_at`, `updated_at`.
- `idempotency_key` and `request_hash`, with the unique key and the same-request rule that stock
  adjustments use (a resend of the same save answers with the first; a different request under a
  used key is 409 `already_saved`).
- An index on `(customer_id, taken_at DESC)`.

Value ranges (rejected with a plain message, not silently accepted): systolic 50-300, diastolic
30-200, pulse 20-250; glucose 1.0-50.0 mmol/L or 20-900 mg/dL. These catch typos such as 1200/80;
they are not clinical limits and the numbers are never judged. Units are stored as entered and never
converted.

### Permissions (migration 51 seeds them)

| Key | Meaning | Default roles |
|---|---|---|
| `healthcheck:record` | Record a reading and print its slip (customer or walk-in) | Cashier, Manager, Admin |
| `healthcheck:view` | See a customer's history; reprint a saved reading | Manager, Admin |
| `healthcheck:correct` | Correct a saved reading (also needs `healthcheck:view`) | Manager, Admin |

Recording for a customer also needs `customer:view` to find them, as the till does. The routes are
gated in the router and the service checks with `HasPermission`; the UI hides what the user cannot
use, and the server decides.

### API

The routes live under their own prefix, `/health-checks/`, mounted like `/customers/` and `/sales/`.

- `POST /health-checks/`: record a reading for a customer (`healthcheck:record` and
  `customer:view`). Body: `customer_id`, the readings, note, `idempotency_key`. Answers with what
  was saved, so the recorder can print the slip right away even without `healthcheck:view`.
  Refuses a deactivated customer (409 `customer_inactive`) and an unknown one (400).
- `GET /health-checks/customer/{id}`: history, newest first, paged (`healthcheck:view`). Writes one
  audit entry per history load.
- `PUT /health-checks/{id}`: correct a reading (`healthcheck:correct` and `healthcheck:view`).
  Sends the `version` it loaded; 409 if it moved; a change of nothing writes nothing.
- `POST /health-checks/{id}/reprint`: records the reprint (audited) and answers with the reading,
  for a slip marked as a copy (`healthcheck:view`).
- A walk-in has no server call at all: the page builds the slip itself from what was typed and the
  signed-in user's name, and prints it.

### Audit

Written in the same transaction as the change (`appendAudit`): entity `health_check`, actions
`recorded`, `corrected`, `reprinted`; entity `customer`, action `health_history_viewed`. A
correction's entry holds the before and after values of the fields that changed, because a medical
value must be traceable (unlike customer edits, which list field names only). Viewing the history
leaves a trace because the data is sensitive.

### Screens

- **Health checks** in the sidebar (needs `healthcheck:record`): choose Customer (the till's
  customer search, add one on the spot) or Walk-in (optional name); tick BP and/or glucose and
  enter the numbers; Save and print. A saved check prints at once; if printing fails the reading
  is saved and a message says so, and a manager can reprint it. Customers who were deactivated
  cannot be chosen.
- **Customer page: Health readings** section, shown only with `healthcheck:view`: the list (date
  and time, BP and pulse, glucose with unit and context, who took it, corrected marker), a Reprint
  button, and Correct for those with `healthcheck:correct`.
- **The slip** (80mm, same print path as receipts, `lib/healthSlip.js`, reusing `escapeHtml`):

```
   Primocrest Pharmacy
   ---------------------
   Health check
   Name: Ada Obi
   19 Nov 2026, 10:42
   ---------------------
   Blood pressure
     120 / 80 mmHg
     Pulse 72 bpm
   Blood glucose
     5.6 mmol/L (fasting)
   ---------------------
   Taken by: Tunde
   For your own records.
```

A reprint is marked `*** COPY <date> ***`. No word on the slip interprets a number.

## Constraints

- Follows the repo's patterns: services own transactions (`runInTx`), retry keys from
  `newRetryKey()`, errors read with `readError`, the shared confirm dialog, JSON `{error, code}`.
- A frontend change is not visible until `make build-frontend`; new Tailwind classes need
  `make tailwind`.
- Never edit migration 50 or earlier.
- Docs: a feature doc under `docs/features/health-checks/` (guide and technical) once built, with
  the entry in `feature-paths.json`.

## Left to the builder (my calls, mark any you disagree with)

- Exact table and column names, the migration number, and the error wording.
- Paging size for history (follow Sales History).
- Whether the Health checks screen is its own route or a dialog on the Customers page (own route,
  my call, so a cashier without `customer:manage` can use it).
- Suggested split into cards: (1) migration, service, permissions, audit; (2) recording screen and
  slip; (3) history, correction and reprint on the customer page.

## Open for the owner

- **Consent.** `contact_consent` on a customer is about being contacted, so this design does not
  check it. If saved health readings need their own consent (data protection rules, NDPA), say so
  and it becomes a prompt on first save. I have not researched the law here.
- **Follow-up cards** once approved: non-stock service items; link a health check to a sale; last
  reading and trend on the slip; other test types.
