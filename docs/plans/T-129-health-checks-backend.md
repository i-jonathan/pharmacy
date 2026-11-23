# T-129 Save blood pressure and glucose readings (health checks backend)

Card: T-129 · Branch: feat/T-129-health-checks-backend

## Goal

The server can save a customer's blood pressure and glucose readings, show a manager their
history, correct a wrong number with a trail, and reprint one. This is piece 1 of 3 of the design in
`docs/superpowers/specs/2026-11-19-health-checks-design.md`; the screen (T-130) and the customer-page
history (T-131) build on it. No UI in this card.

## What I found

- Card's design and plan docs exist and match the card — `docs/superpowers/specs/2026-11-19-health-checks-design.md`, `docs/superpowers/plans/2026-11-19-health-checks.md` (`code`)
- Latest migration is 50, so this is 51 — `backend/db/migrations/000050_sale_customer_old_sales.up.sql` (`code`)
- Permission keys are `lower(resource:action)` from the `permissions` table, so resource `HealthCheck` with actions `Record`/`View`/`Correct` gives `healthcheck:record|view|correct` — `backend/adapter/http/controller/user.go:74` (`code`)
- Seeding pattern for permissions with a working down file — `000047_customers.up.sql`, `000050_*.sql` (`code`)
- Retry key + request fingerprint pattern (unique key, `request_hash`, same user + same hash → first save, else 409 `already_saved`) — `backend/service/stock_adjustments.go:58-110`, `repository/stock_adjustments.go:60-90` (`code`)
- Version-checked edits with row lock and no-op on no change — `service/customers.go` `lockCustomer`, `UpdateCustomer` (`code`)
- Deactivated customer check with `FOR SHARE` so deactivation can't race the save — `repository/customers.go` `GetCustomerRefTx` (`code`)
- Shared helpers `runInTx`, `appendAudit`, `badRequest`, `conflict` — `service/helpers.go` (`code`)
- Router pattern with `RequirePermissions` and router tests that check 403s — `adapter/http/router/customers.go`, `customers_test.go` (`code`); mounting in `backend/main.go:90-92` (`code`)
- Integration test fixture with real migrations — `service/customers_integration_test.go`, `internal/testdb` (`code`)
- Health status is "sensitive personal data" under the Nigeria Data Protection Act 2023 s.30; processing needs one of the listed grounds, explicit consent being one, medical care another. I could not confirm the exact wording of the grounds (`web`, see Sources). Not load-bearing for this card unless you decide readings need consent (Decisions, 1).
- Nothing on the card is wrong or stale.

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| Table, checks, ranges, index, migration 51 with down | design | yes |
| Three permissions seeded to the roles in the design | design | yes |
| Record (retry key, deactivated/unknown customer, audit) | design | yes |
| History, paged, newest first, audited per load | design | yes |
| Correct with version check, before/after in the audit | design | yes |
| Reprint, audited | design | yes |
| Recorder needs `customer:view` too | design | yes |
| A reason on corrections | `model` (unverified) — a medical value changed with no why is a weak trail | Decision 2 |
| Consent for storing readings | `web` NDPA s.30 | Decision 1 |
| Screens and slip | design | follow-up: T-130, T-131 (already on board) |
| Feature docs | design | T-131 (written once the feature works end to end) |

## Decisions for you

1. **Consent before saving readings.** Owner chose: **add a health consent flag now** (see
   "Health consent" under Design).
2. **Reason on a correction.** Owner chose: **a required reason** (up to 200 characters), stored in
   the audit entry.

## Design

### Migration 51 `health_checks` (up and down)

```sql
CREATE TABLE health_check (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customer(id),
    taken_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    taken_by_id INTEGER NOT NULL REFERENCES users(id),
    systolic INTEGER, diastolic INTEGER, pulse INTEGER,
    glucose NUMERIC(4,1), glucose_unit TEXT, glucose_context TEXT,
    note TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1,
    idempotency_key TEXT, request_hash VARCHAR(64),
    created_at, updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- checks:
    --   systolic and diastolic both or neither; systolic > diastolic
    --   pulse only with BP (pulse comes off the same cuff)
    --   ranges: systolic 50-300, diastolic 30-200, pulse 20-250
    --   glucose, unit, context all or none; unit in (mmol/L, mg/dL);
    --     context in (fasting, after_meal, random); range by unit
    --   at least one of BP or glucose
    --   note <= 500 characters
);
CREATE UNIQUE INDEX health_check_idempotency_key_unique ON health_check (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
CREATE INDEX health_check_customer_idx ON health_check (customer_id, taken_at DESC, id DESC);
-- seed HealthCheck/Record to Cashier, Manager, Admin; HealthCheck/View and /Correct to Manager, Admin
```

Down: delete the role permissions and permissions, drop the table. The database checks repeat the
service's, so a bad row can't get in by another path. Pulse alone is refused (my call: a pulse
reading comes with a BP reading; say if you want pulse on its own).

Glucose is `NUMERIC(4,1)`: one decimal place. The service refuses more decimals rather than rounding
a medical value. Units are stored as entered, never converted.

### Health consent (owner's decision, added after first review)

Migration 51 also adds to `customer`: `health_consent BOOLEAN NOT NULL DEFAULT FALSE` and
`health_consent_updated_at TIMESTAMPTZ`. Separate from `contact_consent`, which is about being
contacted.

- Saving a reading for a customer without health consent is refused, 409
  `code: "health_consent_needed"`, unless the request carries `consent_given: true`: the recorder
  asked the customer at the counter. Then the same transaction sets the customer's consent, bumps
  their `version`, and writes `customer` / `health_consent_changed` `{from: false, to: true,
  via: "health_check"}`. This is the "prompt on first save" T-130 will show.
- The customer edit (`PUT /customers/{id}`, `customer:manage`) gains an optional
  `health_consent`. It is a pointer: absent means "leave as is", so today's Customers page, which
  doesn't send it, can't wipe a consent by saving a phone change. A change is audited like
  `contact_consent` (field name in `changed`, from/to) and stamps `health_consent_updated_at`.
  `GET /customers/...` returns `health_consent` and `health_consent_updated_at`.
- Withdrawing consent stops new saves. Readings already saved stay (withdrawal doesn't undo past
  lawful processing); whether they should be deleted on request is a follow-up card.
- Walk-ins are untouched: nothing of theirs is stored.
- A retry of a reading already saved is answered with it, even if consent was withdrawn since
  (key looked up first, as for deactivated customers).

### API (`/health-checks/`, mounted in `main.go` like `/customers/`)

| Route | Needs | Does |
|---|---|---|
| `POST /health-checks/` | `healthcheck:record` + `customer:view` | Record. Body `customer_id`, `systolic`, `diastolic`, `pulse`, `glucose`, `glucose_unit`, `glucose_context`, `note`, `idempotency_key` (required). 201 with the saved reading. |
| `GET /health-checks/customer/{id}?page=` | `healthcheck:view` | History, newest first, 20 per page (Sales History's size), with total. Audited. |
| `PUT /health-checks/{id}` | `healthcheck:correct` + `healthcheck:view` | Correct. Body: the readings, note, `version`, `reason` (required, ≤200). 200 with the reading. |
| `POST /health-checks/{id}/reprint` | `healthcheck:view` | Audited; 200 with the reading. |

The reading returned everywhere: `id, customer_id, customer_name, taken_at, taken_by` (username),
the readings, `note, version, corrected` (version > 1), `updated_at`. That's what the slip needs.

Refusals: unknown customer 400; deactivated customer 409 `customer_inactive` (new saves only; a
retry of a saved one is answered with it, key looked up first, as sales do); bad readings 400 with a
plain message that never echoes the value; stale correction 409 `changed` ("this reading was
changed by someone else; reload and try again"); unknown reading 404; reused key 409
`already_saved`. Routes are gated in the router and the service checks again with `HasPermission`.

### Audit (same transaction, `appendAudit`)

- `health_check` / `recorded` — `{customer_id}` (no values; the row holds them)
- `health_check` / `corrected` — `{changed: {field: {from, to}}}` for the changed fields only, plus
  the reason
- `health_check` / `reprinted` — `{}`
- `customer` / `health_history_viewed` — `{page}`

A correction that changes nothing writes nothing (no version bump, no audit). Health values never go
into error messages, logs or URLs.

## Steps

1. Migration 51 up/down. Test: integration test that the database refuses bad rows (each check) and
   the permissions are seeded to the right roles.
2. `internal/types/health_check.go`, permission constants, repository interface + `repository/health_checks.go`
   (insert with `ON CONFLICT DO NOTHING`, by-key lookup, lock for update, update, history page,
   get one).
3. Service validation (`normalizeHealthCheck`, fingerprint). Test: unit tests for ranges, pairs,
   decimals, units, contexts, at-least-one, note length.
4. Service record / history / correct / reprint. Tests: integration — record and retry, reused key
   by another user or with another body, deactivated and unknown customer, no consent refused,
   `consent_given` sets it with its audit, history order/paging and its audit, correction with
   before/after audit and reason, missing reason, stale version, no-op correction, reprint audit;
   unit — each method refuses without its permission.
5. Customer health consent on the customer edit (types, repository, `UpdateCustomer`,
   `changedCustomerFields`). Tests: integration — absent field leaves consent alone, a change is
   audited and stamped.
6. Controller, router, `main.go`. Test: router test that a cashier (record + customer:view) gets 403
   on history, correct and reprint, and each route reaches the service with the right permissions.
7. `CLAUDE.md`: migration 51 and a short Gotchas entry.

## Testing

- `cd backend && go test ./...` and with `PHARMACY_TEST_DSN` against local Postgres.
- `make check-go-format`, `make check-go-lines`.
- Migration up, down, up on local Postgres.
- Manual: curl the four routes on the running server as a cashier and a manager.
- Not tested here: screens and slip (T-130, T-131).

## Deploy notes

- Run migration 51. Users must sign in again to pick up the new permissions (permissions are read
  at login).
- No frontend change, so no `make assets` needed for this card.

## Commits

Style: **logical groups** (owner's choice). Planned points, each shown to the owner first:

1. Migration 51, types, repository (steps 1–2)
2. Service, including customer consent (steps 3–5)
3. Controller, router, `main.go`, `CLAUDE.md`, this plan (steps 6–7)

## Follow-up cards

- "Health consent on the Customers page form": a checkbox in the customer edit form and a mark in
  the list, P1, Customers & loyalty (or fold into T-131; T-130 covers the prompt on first save).
- "Delete a customer's health readings on request (NDPA)": P2, Customers & loyalty.
- T-130 body gains: prompt for consent on first save, send `consent_given`, handle
  `health_consent_needed`.

## Sources

- [FPF: Nigeria's New Data Protection Act, Explained](https://fpf.org/blog/nigerias-new-data-protection-act-explained/) — sensitive data and the narrower role of consent
- [Mondaq: Data dignity, digital health](https://www.mondaq.com/nigeria/data-protection/1781892/data-dignity-digital-health-balancing-technological-innovation-and-patient-privacy) — health status is sensitive personal data; National Health Act 2014 also applies
- [NDPC GAID 2025](https://ndpc.gov.ng/wp-content/uploads/2025/07/NDP-ACT-GAID-2025-MARCH-20TH.pdf) — the implementing directive, authoritative for the details I could not confirm
