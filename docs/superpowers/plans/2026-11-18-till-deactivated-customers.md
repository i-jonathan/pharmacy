# Deactivated customers at the till Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The till search shows deactivated customers as deactivated, a user with `customer:manage` can reactivate one from the till (which attaches them), and a resumed held sale with a deactivated customer says so and cannot be completed.

**Architecture:** Frontend only. `CustomerPicker.vue` searches with `inactive=1`, marks deactivated rows, and reactivates through the existing `POST /customers/{id}/reactivate` inside the shared `confirm()` dialog. `usePos.js` keeps `active` on the attached customer; `PosView.vue` blocks Complete while it is `false`. The server's save-time refusal (409 `customer_inactive`) stays as the backstop.

**Tech Stack:** Vue 3 SPA (`frontend/src/next`), Go docs. No backend change, no new permission.

**Spec:** `docs/superpowers/specs/2026-11-18-till-deactivated-customers-design.md`

## Global Constraints

- A deactivated customer is never picked, attached or sent as `customer_id`; reactivate first.
- The Reactivate button shows only with `customer:manage`; without it the row says "Deactivated, ask a manager".
- Confirmations use `confirm({ title, message, detail, tone, confirmLabel, busyLabel, run })` from `composables/useConfirm.js`; no new modal, no `window.confirm`.
- Errors come from `readError(response, fallback)` in `lib/api.js`; mutating calls use `csrfFetch` from `lib/csrf.js`.
- No JS test tooling exists: frontend tasks are verified with `cd frontend && npx vite build` and the manual script in Task 3. Do not add a test framework.
- No new Tailwind class unless `make tailwind` is run and `backend/template/static/css/tailwind.css` is kept in the change.
- Do not commit in tasks. The owner confirms one commit and push at the end (Task 3, last step).

## Review Focus

- Search matches an active and a deactivated customer for one phone/name: the active one is listed first, and the picker's six-row cap never hides every active match behind deactivated ones.
- Enter with exactly one result that is deactivated does nothing (no attach, no reactivate).
- Reactivate when someone else already reactivated or edited the customer: the dialog shows the server's message, nothing attaches twice.
- Reactivate confirmed, then the cashier closes the picker or completes the sale before it returns: the customer is reactivated but not attached to whatever is on screen.
- A user with `customer:view` but not `customer:manage` sees "Deactivated, ask a manager" and no button, and cannot reactivate by any picker path.
- Resumed held sale with a deactivated customer: Complete, Complete and Print and F5 are refused with the deactivated message, not "Enter the full payment".

---

### Task 1: Picker lists, marks and reactivates deactivated customers

**Files:**
- Modify: `frontend/src/next/components/pos/CustomerPicker.vue`

**Interfaces:**
- Consumes: `GET /customers/?q=&inactive=1` rows `{id, name, phone, active, version}`; `GET /customers/{id}`; `POST /customers/{id}/reactivate` body `{version}`; `confirm` from `useConfirm()`.
- Produces: the `update:modelValue` payload is now `{id, name, phone, active}` (was `{id, name, phone}`); `active` is `true` for anything picked or reactivated here.

- [ ] **Step 1: Search request.** In `search()` request `/customers/?q=<text>&inactive=1`. After the response, store `results` ordered with `active !== false` rows first, keeping the server's order within each group (stable partition).

- [ ] **Step 2: Rows.** In the results list, a row with `c.active === false` is a non-clickable `div` (not the pick `button`), muted, with a "Deactivated" badge. With `canManage` (`hasPermission("customer:manage")`) it also has a small **Reactivate** button (`@click="reactivate(c)"`); without it, the text "Deactivated, ask a manager". Active rows are unchanged.

- [ ] **Step 3: `pickFirst` and `noMatch`.** `pickFirst` picks only when the single result has `active !== false`. `noMatch` is unchanged (a deactivated match is a match: no "Add as new" button is offered for it).

- [ ] **Step 4: `async function reactivate(c)`** in the picker. Takes `const mine = createCount` first. Calls `confirm({ title: "Reactivate customer?", message: "They will be added to this sale.", detail: c.name, tone: "warning", confirmLabel: "Reactivate and add", busyLabel: "Saving…", run })`. `run`: `GET /customers/${c.id}` for the current `version` (non-ok -> throw the `readError` message); `csrfFetch POST /customers/${c.id}/reactivate` with `{version}` (non-ok -> throw the `readError` message, which the dialog shows inline; network failure -> throw `new Error("Could not reach the server. Check the connection and try again.")`); read the customer back with `GET /customers/${c.id}`; then, only if `mine === createCount`, call `pick({ ...saved, active: true })`. A stale `mine` reactivates without attaching.

- [ ] **Step 5: Attached deactivated customer.** In the "customer on this sale" branch, when `modelValue.active === false` show the "Deactivated" badge and, with `canManage`, a **Reactivate** button calling `reactivate(modelValue)`; its success emits `update:modelValue` with the reactivated customer. The existing Change and remove (x) buttons stay.

- [ ] **Step 6: Remove the F-004 workaround.** Delete `blockedPhone`, `phoneBlocked`, their uses in the template and in the `watch(query)`, and the `blockedPhone.value = ""` in `reset()`. The `phone_taken` branch keeps `mode.value = "search"; query.value = phone` and the error text, so the owner now appears in the results with their state. Update its comment accordingly.

- [ ] **Step 7: Include `active: true`** in `pick()`'s emitted object (`{ id, name, phone, active: c.active !== false }`).

- [ ] **Step 8: Verify.** Run `cd frontend && npx vite build`. Expected: build succeeds with no new warnings.

### Task 2: Held sale and Complete respect a deactivated customer

**Files:**
- Modify: `frontend/src/next/composables/usePos.js`, `frontend/src/next/components/PosView.vue`

**Interfaces:**
- Consumes: Task 1's `active` on the customer object.
- Produces: `canComplete` is false while `pos.customer.value?.active === false`.

- [ ] **Step 1: `usePos.js`.** `refreshCustomer(id)` sets `active: latest.active` on the refreshed object (so a resumed held sale learns it). `restoreHeld` leaves `active` undefined until then.

- [ ] **Step 2: `PosView.vue`.** Add to `canComplete`: `&& pos.customer.value?.active !== false`. In the `complete` shortcut handler, when the customer is deactivated (and the cart is non-empty and not processing) set `actionError` to "This customer is deactivated. Reactivate them or remove them from the sale." and `mobilePane.value = "cart"`, instead of the "Enter the full payment" message.

- [ ] **Step 3: Verify.** `cd frontend && npx vite build`; expected success.

### Task 3: Docs, manual check, handoff

**Files:**
- Modify: `docs/features/customers/guide.md` (lines about the till, deactivated customers at the till, and the two error rows), `docs/features/customers/technical.md`, `docs/plans/T-118-attach-customer-to-sale.md` (note the reversed rule in its decisions), `CLAUDE.md` (the Customers gotcha: the till search now lists deactivated customers and can reactivate them), `.claude/reviews/T-118.md` (a "round 2 scope" note listing these changes).

- [ ] **Step 1: Docs.** Change the rule "A deactivated customer is not found at the till" to: found, labelled Deactivated, cannot be picked until reactivated; a user with `customer:manage` reactivates from the till and the customer joins the sale. Update the `phone_taken` row to say the owner shows in the search with Reactivate, and the "customer has been deactivated" row to say a resumed held sale shows it before saving. Keep `docs/features/feature-paths.json` accurate if a path is added.

- [ ] **Step 2: Build.** `make assets`; then restart the demo server on 8123 (stop the background `go run .`, start `cd backend && DB_HOST=localhost DB_PORT=5544 PHARMACY_PORT=8123 go run .`).

- [ ] **Step 3: Manual script** on `http://127.0.0.1:8123` (password `Demo-Pharmacy-2026`), each with expected result:
  1. As `tunde`, deactivate nobody; as `amaka` deactivate one customer, keep another active that shares part of the name. At the till (`tunde`) search that name: the active one is listed first; the deactivated one is greyed with "Deactivated" and a Reactivate button; clicking its row does nothing.
  2. Search a phone that matches only the deactivated customer and press Enter: nothing attaches.
  3. Click Reactivate, confirm: customer is attached, no "Deactivated" label, sale completes. On the Customers page they show as active.
  4. Add a new customer with the deactivated customer's phone: the owner appears in the results with Reactivate (no "No customer found").
  5. Hold a sale with a customer, deactivate them as `amaka`, resume as `tunde`: the cart shows Deactivated with Reactivate and Remove; Complete, Complete and Print and F5 are refused with the deactivated message; Reactivate fixes it.
  6. With a throttled network, confirm Reactivate then immediately press X: the customer is reactivated but not attached.
  7. Sign in as a role with `customer:view` only (remove `customer:manage` from a test role or use the Customers page permission list) and check the row reads "Deactivated, ask a manager" with no button. If no such user is practical, state that this was not exercised.

- [ ] **Step 4: Report and confirm commit.** Run `go vet ./...` and `PHARMACY_TEST_DSN=... go test ./...` in `backend` (expected unchanged: pass). Tell the owner what was checked and what was not, then ask for one yes to commit everything and push, per the owner's rule (one confirmation after implementation and review).
