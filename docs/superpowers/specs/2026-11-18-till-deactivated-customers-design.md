# Deactivated customers at the till

Status: approved by the owner on 2026-11-18. Part of T-118 (PR #41), found while trying the till.

## Intent

A cashier should know a customer is deactivated as soon as they find them, not when the sale is
refused at save time, and should be able to reactivate them without leaving the till.

## Goals

- The till search finds deactivated customers and shows them as deactivated.
- A user with `customer:manage` can reactivate one from the till; the customer is then attached to
  the sale.
- A held sale resumed with a deactivated customer shows that state at once and cannot be
  completed until the customer is reactivated or removed.
- The `phone_taken` reply for a deactivated owner leads somewhere useful (the owner, with
  Reactivate), replacing the F-004 workaround (`blockedPhone`).

## Non-goals

- No new permission: Cashiers, Managers and Admins already hold `customer:manage`.
- No polling for a customer deactivated while a sale is open. The server still refuses a
  deactivated customer at save time (409 `customer_inactive`) as the backstop.
- No change to the Customers page.

## Constraints

- A deactivated customer can never be picked; they must be reactivated first.
- Reactivation sends the `version` read in the same search, and a 409 (changed elsewhere) is shown
  with its message and the row refreshed.
- Confirmations use the shared `confirm()` dialog (`composables/useConfirm.js`), not a new modal.
- Without `customer:manage` the row says "Deactivated, ask a manager" and offers no button.
- Reverses the documented rule "a deactivated customer is not found at the till"; the customers
  guide and technical doc change with it.

## Approaches

1. **Picker searches with `inactive=1` and marks deactivated rows (chosen).** One request, no
   backend change; `types.Customer` already carries `active` and `version`.
2. A second request for deactivated matches only when the first returns nothing. Two code paths,
   and a deactivated customer would be hidden whenever an active one also matches.
3. A new till-specific search endpoint. More surface for no gain.

## Design

- `CustomerPicker.vue`: search adds `inactive=1`; rows with `active === false` are disabled and
  labelled "Deactivated", with a **Reactivate** button for users with `customer:manage`. Enter
  (`pickFirst`) never picks a deactivated row. Reactivate -> `confirm()` -> `POST
  /customers/{id}/reactivate` with `{version}` -> read the customer back -> pick it. Uses the
  existing request token so a cancelled or overtaken action cannot attach late.
- `usePos.js`: the attached customer carries `active` (from the picked row and from
  `refreshCustomer`); `canComplete`-style gating is by the cart panel showing the state, with
  **Reactivate** and **Remove** on the attached customer.
- `phone_taken`: the owner is in the results, so `blockedPhone` and its template guards are
  removed.
- Docs: customers guide and technical doc; plan decision list updated.

## Left to the builder

Visual treatment of the deactivated row and the attached-customer banner (use existing tokens in
`design-system/primocrest-pharmacy/MASTER.md`); exact copy.
