# Merge a draft into an open purchase order (design)

Status: approved by the owner in conversation on 2026-11-20 for the approach (option 2 below); this document is the short design, awaiting final approval before a plan is written.

## Intent and why

A purchase order goes out to a supplier. Before the delivery, staff notice more things they need and start a new draft. The supplier then delivers only part, and a manager uses "Move outstanding" to put the undelivered items on a new order for a different supplier. The shop now holds two orders it wants to send as one: the draft, and the open order that "Move outstanding" created (not yet sent to the supplier).

Merging puts the two together so a cashier can finish one order and send it through approval again.

## Goals

- One action that folds a draft's lines into an open order and hands back one **draft** for the cashier to keep editing.
- The merged order goes back through the normal approval flow (submit, then a manager approves). The merge never produces an approved order, so it cannot be used to skip approval.
- Both orders' history says what happened.

## Non-goals

- Merging two drafts, two open orders, or anything with stock already received.
- Pulling an approved order back to a draft as a general action (the existing "pull back" stays submitted-only). May be a later card.
- Choosing line by line what to merge, or splitting an order.
- A "What's new" slideshow entry (can be added at release time).

## Anti-goals

- No path where stock received against an order is lost or its lines are altered. The open order must have nothing received, which is what lets it return to a draft safely.
- No silent data loss: notes are kept, and nothing is changed or dropped without the user seeing it in a preview first. Where both orders list the same item, the higher quantity wins, so the preview shows which figure was kept.

## Approaches considered

1. **Merge, then add "pull back" for approved orders.** The merged order stays open; a separate manager action turns an open order with nothing received into a draft. More flexible, but opens a new state change on orders that may already be sent, and is two steps.
2. **The merge itself leaves a draft (chosen).** One step, the result needs approval again (the end state the owner described), and a cashier can do it because it can't bypass approval. Cost: only an open order with nothing received can be the target.

## Design

**Who and what can be merged**
- The target is an order with status `open` and no received quantity on any line (a "partial" order is refused).
- The source is an order with status `draft`.
- The caller needs `purchaseorder:approve` (a manager; route gated with `mayApprove`, service checks `canApprovePurchaseOrders`). Changed after review: the merge undoes the approval of an order that may already be with the supplier, which a drafter cannot do. The draft being merged can be anyone's. Everyone with either permission already sees all orders.
- The caller sends both ids and the version of each that they saw. A stale version is 409 `purchaseOrderChangedElsewhere`, as for every other order action.

**What happens (one transaction, `inPurchaseOrderTx`-style, both rows locked `FOR UPDATE`, in id order to avoid deadlocks)**
- The target keeps its id, supplier and `source_purchase_order_id`. Its lines receive the draft's lines:
  - same product (`product_id`) on both: one line, keeping the **higher** of the two quantities (not the sum, because both orders were usually written from the same need);
  - lines typed by name (no product): matched on name plus manufacturer, ignoring case, as the editor already does, with the same higher-quantity rule;
  - anything else: added as a new line.
  - Because nothing is added together, no quantity can pass the INTEGER ceiling.
- The target's status becomes `draft`, its version goes up, and its approval stamp (`approved_by_id`, `approved_at`) is cleared, since it is no longer approved. The earlier approval stays in the audit log.
- The draft is retired the way a rejected draft is: status `closed`, kept for the record, with no PDF. Its lines stay as they were.
- Notes: the target's note stays; the draft's note is added after it when it is different and not empty, within the existing note limits (a longer result is a 400 that says so). Line notes combine the same way.
- Supplier: the target's supplier is kept. If the draft names a different supplier, the page warns in the preview; the server does not refuse.
- Audit: a new action `merged` on both orders. On the target: the draft's id, the lines added, and the lines whose quantity went up (before and after). On the draft: the target's id. The history screen in `PurchaseOrders.vue` learns the new action.

**Screen**
- On an open order with nothing received, **Merge a draft in** (shown only when a draft exists). It lists the drafts, previews the combined lines (new, quantity raised from X to Y, kept as is, supplier warning), and confirms through the shared confirmation dialog, which says the order will go back to draft and need approval again.

**Docs**
- Update `docs/features/purchase-orders/guide.md` and `technical.md`.

## Calls made by me (change any of these)

- Same item on both orders keeps the higher quantity (the owner's decision, replacing my first idea of adding them).
- The retired draft becomes `closed` rather than being deleted or getting a new `merged` status, so no migration is needed.
- Clearing the approval stamp on the target (rather than leaving it).
- Joining differing notes instead of refusing the merge.
- Only managers may merge (first decided that cashiers could, since the result must be approved again; changed after review because the merge reverses an approval).
- One draft per merge; merging several means repeating it.

## Left to the builder

Endpoint shape (for example `POST /inventory/purchase-orders/{id}/merge` with the draft id and both versions), repository queries, the exact preview component, and tests (service tests with the mock repo plus a database integration test in `purchase_orders_integration_test.go`).
