# Review a ticket

Load at the start of phase 5. There is no second reviewer, so this review is the gate.

## Run

1. `/code-review high` on the branch diff (the `code-review` skill). Fix every confirmed finding; list any you leave and why.
2. Backend: `make check-go-format`, `make check-go-lines`, `cd backend && go test ./...`, and with `PHARMACY_TEST_DSN` set when the change touches SQL or migrations.
3. Migrations: apply, roll back and re-apply against a scratch database.
4. Frontend: `make assets` (Tailwind and Vite), then run the app and use the changed screens (the `run` skill). Commit `backend/template/static/css/tailwind.css` if it changed.
5. **REQUIRED SUB-SKILL:** superpowers:verification-before-completion before saying anything passes.

## Check by hand

- Plan vs diff: every plan step done or explained; nothing in the diff that the plan didn't cover.
- Money is integer kobo end to end; no float naira multiplied into kobo.
- Every new route and UI action has a permission check on the server and matching gating in the UI.
- Every change to stock, money, users or permissions writes an audit entry in the same transaction.
- Mutating requests carry CSRF tokens; retries are idempotent where a duplicate would hurt (sales, receipts, orders).
- Concurrent edits are handled (versions or row locks) where two users can touch the same record.
- Migrations have working down files and are safe on existing data.
- Errors reach the user in plain words; nothing silently swallowed.
- Existing feature docs under `docs/features/` that this change makes wrong are refreshed (or listed for the ship step).
- Each claim on the card was checked; the card's checklist matches what was built.

Record the results (commands run and their outcomes, manual checks done) for the PR's Testing section.
