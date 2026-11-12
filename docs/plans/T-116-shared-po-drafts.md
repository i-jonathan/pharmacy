# T-116 Shared purchase order drafts

Card: T-116 · Branch: feat/T-116-shared-po-drafts

## Goal

Two cashiers no longer write the same purchase order twice. Anyone who can draft purchase orders
can see every order, edit any draft, submit any draft, and pull a submitted order back to a draft
to fix a mistake without waiting for a manager. Managers keep approving.

## What I found

- `code` Editing: `checkPurchaseOrderEditable` lets a drafter change only their **own** order, and
  only while it is a draft; a submitted order gives "submitted; a manager will review it"
  (`service/purchase_orders.go:78`). The Edit button follows the same rule (`canEdit`,
  `PurchaseOrders.vue:831`).
- `code` Submitting: `SubmitPurchaseOrder` refuses anyone but the author, "only the author can
  submit a draft" (`service/purchase_order_approval.go:46`); the Submit button needs `isMine`
  (`PurchaseOrders.vue:836`). If someone else fixes a draft, the author must come back to submit it.
- `code` There is no way to take a submitted order back: no route, no service method.
- `code` Visibility: `visiblePurchaseOrder` hides drafts and submitted orders from everyone but their
  author and approvers (`service/purchase_orders.go:483`), and `ListPurchaseOrders` filters **every**
  tab by `created_by_id` for non-approvers (`:466`, `query.go:579`). So a cashier also doesn't see
  other people's *approved* orders in "Waiting for stock", although the comment on
  `visiblePurchaseOrder` says approved orders are visible to anyone. The list and the single-order
  read disagree; this is part of why duplicates happen.
- `code` Concurrent edits are already safe: every edit sends the version it loaded and gets a 409 if
  the order changed (`editPurchaseOrder`, `service/purchase_orders.go:381`), and every change writes
  an audit entry naming the editor (`auditPurchaseOrder`), shown in the order's history.
- `code` Purchase orders hold quantities and no prices (`PurchaseOrderItem`), so showing them to
  all drafters exposes no cost data.
- `code` The cashier role already has `purchaseorder:draft` (migration 43); no new permission or
  migration is needed. Statuses `draft` and `pending_approval` already exist, and
  `SetPurchaseOrderStatusTx` changes between them.
- `code` Documents that state the old rule and must change with it: `docs/features/purchase-orders/
  guide.md` (lines 11, 93) and `technical.md` (lines 45, 46, 72, 73, 101, 102), the
  purchase-orders paths in `docs/features/feature-paths.json` (a hook blocks the PR otherwise), and
  the purchase-order paragraph in `CLAUDE.md`.
- `code` The order card does not show who wrote a draft (only `isMine` is used), which matters
  once drafts are shared.

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| Any drafter can edit any draft | owner | yes |
| Any drafter can submit any draft | `code` (author-only today) | yes |
| Drafters see every order in every tab, not only their own | `code` | yes |
| Pull a submitted order back to a draft (drafters and managers) | owner (recommended) | yes |
| Cards say who wrote a draft and who last changed it | `code` | yes |
| Docs, tests and `CLAUDE.md` match the new rules | `code` | yes |
| Notice in the editor when a product is already on another open or draft order | `model` (unverified) | follow-up |
| Notify the author when someone else edits their draft | `model` | no |

## Decisions for you

1. **Anyone who can draft may submit any draft.** Recommended: yes. Otherwise a draft that someone
   else fixed waits for its author. The audit log records who submitted.
2. **Pull back to draft.** Recommended: any drafter or manager may move a *submitted* order back
   to a draft; it is audited and the order can then be edited and submitted again. Without it a
   cashier who spots a mistake after submitting has to ask a manager to reject it.
3. **Approved orders visible to every drafter.** Recommended: yes, so "Waiting for stock" and
   "Completed" show everyone's orders, as the single-order read already allows. Otherwise a
   duplicate of an *approved* order is still possible.

## Design

**No migration, no new permission.**

**Backend** (`service/purchase_orders.go`, `purchase_order_approval.go`, router, controller)
- `checkPurchaseOrderEditable`: an approver can change anything unfinished (unchanged); a drafter
  can change **any** order while it is a draft. Submitted orders stay manager-only to edit.
- `SubmitPurchaseOrder`: drop the author check; a drafter or approver can submit any draft.
- New `WithdrawPurchaseOrder(ctx, id, version, userID)` and `POST /inventory/purchase-orders/{id}/withdraw`
  (same gate as submit: draft or approve). Only a `pending_approval` order can be withdrawn; it
  goes back to `draft` with an `AuditWithdrawn` entry, using `inPurchaseOrderTx` for the lock and
  version check.
- `ListPurchaseOrders`: the `created_by_id` filter applies only to people with neither
  `purchaseorder:draft` nor `purchaseorder:approve`. `visiblePurchaseOrder`: drafters see drafts and
  submitted orders too.
- Rejected drafts, approval, close, move and receiving are untouched.

**Frontend** (`PurchaseOrders.vue`)
- `canEdit`: a drafter on any draft; `canSubmit`: any drafter or approver on a draft; new
  `canWithdraw`: a drafter or approver on a submitted order, asked through the shared confirm
  dialog.
- Draft and submitted cards show "Started by <name>" and, when someone else changed it, "last
  changed by <name>" from the existing history.
- A stale edit already shows the "changed elsewhere" message; check it reads well when it is
  another cashier.

**Docs**: update the guide, the technical doc, `CLAUDE.md`.

## Steps

1. Backend rules: editable, submit, list and visibility. Tests first (`TestPurchaseOrderApprovalRules`
   and the integration tests): a cashier edits and submits another cashier's draft; cannot edit a
   submitted order; sees all orders in every tab; a user with neither permission still sees only
   their own.
2. `WithdrawPurchaseOrder`, route, controller, audit constant. Tests: only a pending order can be
   withdrawn; stale version is refused; the audit entry names the user.
3. Frontend: the three button rules, the "Started by" line, the withdraw confirm.
4. Docs and `CLAUDE.md`.
5. `make assets`; use the page as a cashier and a manager on the demo database.

**As built (differences from the plan):**
- The cards show "Started by <name>" ("you" for your own) on drafts and submitted orders. The
  "last changed by" line was dropped: it needs the history of every order, and the order's
  View → Activity already shows who changed what.
- `GetPurchaseOrderHistory` now lets any drafter read the activity of an order they can see (it was
  author-only); a user with neither permission still reads only their own.
- The drafter tab was labelled "My drafts"; it is "Drafts" now that they are shared.
- The "submitted" message for a cashier now says how to change it ("pull it back to a draft first").
- Owner's answers to the three decisions: yes to all (any drafter submits; pull back to draft for
  drafters and managers; approved orders visible to every drafter).

## Testing

- `go test ./...` with `PHARMACY_TEST_DSN`; the new rules have unit and integration tests.
- Two signed-in people editing one draft at the same time (the second gets the 409 message).
- In the running app on the demo database (`docs/whats-new/README.md` has the steps): as `tunde`,
  edit and submit a draft written by `amaka`, pull a submitted order back, and see every tab.
- Not tested: more than two people editing at once, and a real shop's roles beyond the three demo
  roles.

## Deploy notes

No migrations or permissions. `make assets` before building. Behaviour changes for cashiers on
deploy: they will suddenly see every order.

## Commits

Logical groups, proposed once after implementation and review: backend rules and withdraw, then
the page, then docs, plan and board.

## Follow-up cards

- Notice in the editor when a product is already on another open or draft order (P2, Purchasing & receiving): T-117

## Sources

None; internal rules. Code read: `service/purchase_orders.go`, `purchase_order_approval.go`,
`repository/purchase_orders.go`, `PurchaseOrders.vue`, `docs/features/purchase-orders/`.
