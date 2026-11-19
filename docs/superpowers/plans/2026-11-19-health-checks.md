# Health checks: plan

Design: `docs/superpowers/specs/2026-11-19-health-checks-design.md`. Three pieces, each its own board card
and pull request, built in this order.

## 1. Saving readings (backend)

- A new `health_check` table (migration 51) and three permissions: `healthcheck:record` (Cashier, Manager,
  Admin), `healthcheck:view` and `healthcheck:correct` (Manager, Admin).
- Recording a reading for a customer, reading a customer's history, correcting a reading and reprinting
  one. Routes live under `/health-checks/`.
- Same safeguards as stock adjustments: a retry key so a double click saves once, an audit entry in the same
  transaction, a version check on corrections. Opening someone's history is audited too.
- Sensible ranges reject typos such as 1200/80. A deactivated customer cannot be recorded.

**Tests:** the database refuses bad rows; the service refuses bad readings, a deactivated customer, and a
stale correction; a cashier calling history, reprint or correct directly gets 403.

## 2. Recording screen and slip

- A "Health checks" screen: pick a customer or Walk-in, enter BP and/or glucose, Save and print.
- The slip is built in the browser and printed the way receipts are. A walk-in never reaches the server.
- Nothing on the slip says "normal" or "high".

**Check:** build, then try it on a demo database as a cashier: customer and walk-in slips, and no history
visible.

## 3. History on the customer page

- A "Health readings" section on the customer, only for people with `healthcheck:view`: newest first, with
  Reprint (marked as a copy) and, for managers, Correct.
- Feature docs written once it works.

**Check:** try it as a manager on the demo database, including a stale correction from a second tab.

## Later (not in this plan)

Non-stock service items for the fee, linking a check to its sale, last reading on the slip, other tests, and a
consent prompt if you decide readings need one.
