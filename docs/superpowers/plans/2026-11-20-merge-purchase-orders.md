# Merge a draft into an open purchase order: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** One action folds a draft purchase order's lines into an open order that has nothing received; the open order comes back as a draft needing approval again, and the old draft is closed.

**Architecture:** A new service method `MergePurchaseOrder` in a new file `backend/service/purchase_order_merge.go`, built on a pure planning function (`planPurchaseOrderMerge`) so the line rules are unit tested without a database. It runs in one transaction with both rows locked, reusing existing repository calls plus one new query that reverts an order to draft. One new route, one button and inline panel on the Purchase Orders page. No migration.

**Tech Stack:** Go (net/http mux, sqlx, raw SQL), Postgres, Vue 3 SPA (`frontend/src/next/views/PurchaseOrders.vue`).

**Spec:** `docs/superpowers/specs/2026-11-20-merge-purchase-orders-design.md`

## Global Constraints

- Same item on both orders keeps the **higher** quantity, never the sum.
- The target must be status `open` with every line's `received_quantity` 0; the source must be status `draft`. A `partial` order is refused.
- Permission: `purchaseorder:draft` or `purchaseorder:approve` (`canWorkOnPurchaseOrders`); route gated with `mayDraft`.
- Both versions are sent by the caller; a stale one is `purchaseOrderChangedElsewhere()` (409).
- The target keeps its id, supplier and `source_purchase_order_id`; it becomes `draft`, approval stamp cleared. The draft is closed with `ClosePurchaseOrderTx`.
- Audit action `merged` on both orders, written in the same transaction (`s.auditPurchaseOrder`).
- Limits stay: order note `maxPurchaseOrderNote` (1000), line note `maxPurchaseOrderItemNote` (500); a longer combined note is 400.
- Errors are JSON `{error, code?}`; use `badRequest(...)`, `httperror.NewHttpError(http.StatusConflict, ...)`, `httperror.Forbidden(...)`.
- Go lines wrap at 100 columns (`make check-go-lines`), `make check-go-format`. Never edit an existing migration (none needed here).
- Frontend: a yes/no goes through `confirm({...})` from `composables/useConfirm.js`, no new modal. Any new Tailwind class needs `make tailwind` and the result committed; the Go server needs `make build-frontend` and a restart to show frontend changes.
- No commits during the tasks. The owner approves one commit and push at the end, after review (see memory: confirm before commit and push).

## Review Focus

1. Draft quantity lower than or equal to the open order's: the open order's figure stays and the line is not reported as raised.
2. Name-only lines match on name plus manufacturer ignoring case, and a name-only draft line never matches a product line.
3. The combined order note over 1000 characters (or a line note over 500): 400, nothing written.
4. Double click or two people merging the same draft: only one merge commits, the other gets 409.
5. Target not mergeable (partial, received stock, draft, pending, closed), draft not a draft, or the same order given twice: refused with the right status and no writes.

## File Structure

- Create `backend/service/purchase_order_merge.go`: `planPurchaseOrderMerge`, `MergePurchaseOrder`.
- Create `backend/service/purchase_order_merge_test.go`: unit tests with the existing fake repo.
- Modify `backend/service/purchase_orders_integration_test.go`: database tests.
- Modify `backend/service/interfaces.go`, `backend/repository/repository.go`, `backend/repository/purchase_orders.go`, `backend/repository/query.go`, `backend/internal/constant/constant.go`, `backend/internal/types/inventory.go`.
- Modify `backend/adapter/http/controller/inventory.go`, `backend/adapter/http/router/inventory.go`.
- Modify `frontend/src/next/views/PurchaseOrders.vue`.
- Modify `docs/features/purchase-orders/guide.md`, `docs/features/purchase-orders/technical.md`, `CLAUDE.md` (the purchase-order bullet).

---

### Task 1: The merge rules (pure function)

**Files:**
- Create: `backend/service/purchase_order_merge.go`
- Test: `backend/service/purchase_order_merge_test.go`

**Interfaces:**
- Produces (used by Task 2):
  ```go
  type purchaseOrderMergedLine struct {
      item           types.PurchaseOrderItem // the target line as it will be saved (ID set)
      quantityBefore int
  }
  type purchaseOrderMergePlan struct {
      note    string                          // combined order note
      updates []purchaseOrderMergedLine       // matched target lines that change (higher quantity and/or combined note)
      inserts []types.PurchaseOrderItem       // draft lines with no match, ID 0, ReceivedQuantity 0
  }
  func planPurchaseOrderMerge(target, draft types.PurchaseOrder) (purchaseOrderMergePlan, error)
  ```

- [ ] **Step 1: Write the failing tests** in `purchase_order_merge_test.go`, table style like `TestPlanPurchaseOrderEdit`. Names and assertions:
  - `TestPlanPurchaseOrderMerge`:
    - "same product keeps the higher quantity": target Alpha(product 1) 10, draft Alpha 15 → one update with `item.Quantity == 15`, `quantityBefore == 10`, no inserts.
    - "a lower or equal draft quantity changes nothing": draft 4 and draft 10 against 10 → no updates, no inserts.
    - "lines without a product match on name and manufacturer ignoring case": target `{Name:"Amoxil", Manufacturer:"GSK"}` 5, draft `{Name:"AMOXIL", Manufacturer:"gsk"}` 8 → update to 8. Same name, different manufacturer → insert.
    - "a name-only draft line does not match a product line": target product 1 "Alpha", draft no product "Alpha" → insert.
    - "unmatched draft lines are added": draft Beta → inserts has Beta with `ID == 0`, same quantity, same note.
    - "notes are joined": order notes "A" and "B" → `plan.note == "A\nB"`; equal or empty draft note → target's note unchanged; empty target note → draft's note. Matched line notes join with `"; "` and make an update even when the quantity is not higher (quantity unchanged).
    - "too long notes are refused": a joined order note over `maxPurchaseOrderNote` runes and a joined line note over `maxPurchaseOrderItemNote` runes each return a 400 (`requireHTTPStatus`), the line one naming the item.

- [ ] **Step 2: Run** `cd backend && go test ./service/ -run TestPlanPurchaseOrderMerge -v`. Expected: build failure (function undefined).

- [ ] **Step 3: Implement `planPurchaseOrderMerge`.** Match key: `"p"+productID` when both lines have a product id, else lowercase `name + "\x00" + manufacturer` only when both have none. Note joining helper: empty or equal (case-sensitive after trim) second note → first; empty first → second; otherwise first + separator + second (`"\n"` for the order note, `"; "` for line notes). Rune counts via `utf8.RuneCountInString`.

- [ ] **Step 4: Run** the same command. Expected: PASS.

---

### Task 2: The service action, repository query, audit constant

**Files:**
- Modify: `backend/internal/constant/constant.go` (add `AuditMerged = "merged"` next to `AuditMoved`)
- Modify: `backend/internal/types/inventory.go` (add the request type)
- Modify: `backend/repository/query.go`, `backend/repository/purchase_orders.go`, `backend/repository/repository.go`
- Modify: `backend/service/interfaces.go`
- Modify: `backend/service/purchase_order_merge.go`
- Test: `backend/service/purchase_order_merge_test.go`, `backend/service/purchase_orders_integration_test.go`

**Interfaces:**
- Consumes: `planPurchaseOrderMerge` (Task 1); existing `GetPurchaseOrderForUpdateTx`, `UpdatePurchaseOrderHeaderTx`, `UpdatePurchaseOrderItemTx`, `InsertPurchaseOrderItemTx`, `ClosePurchaseOrderTx`, `s.inTx`, `s.auditPurchaseOrder`, `purchaseOrderChangedElsewhere`, `canWorkOnPurchaseOrders`, `purchaseOrderItemError`, `auditItemSummaries`.
- Produces (used by Task 3):
  ```go
  // types
  type MergePurchaseOrderRequest struct {
      Version      int `json:"version"`        // of the open order in the URL
      DraftID      int `json:"draft_id"`
      DraftVersion int `json:"draft_version"`
  }
  // repository.PharmacyRepository
  RevertPurchaseOrderToDraftTx(ctx context.Context, tx *sqlx.Tx, id int) error
  // service.InventoryService
  MergePurchaseOrder(ctx context.Context, id int, req types.MergePurchaseOrderRequest, userID int) error
  ```

- [ ] **Step 1: Write the failing unit tests** in `purchase_order_merge_test.go` with `fakePurchaseOrderRepo`. Extend the fake: an `orders map[int]types.PurchaseOrder` that `GetPurchaseOrderForUpdateTx` consults before `order`, plus recorded `updatedLines`, `reverted []int`, `closedIDs []int`. Test `TestMergePurchaseOrder`:
  - "a manager merges a draft into an open order": `approverContext()`, target 1 open (version 3, supplier "Emzor", Alpha 10), draft 2 draft (version 2, Alpha 15, Beta 4) → no error; `reverted == [1]`, draft 2 closed, one line updated to 15, one line inserted (Beta), two `merged` audit entries (one per order id), one commit, no rollback.
  - "a drafter can merge": `drafterContext()` succeeds.
  - "no purchase order permission is refused": `bareContext()` → 403, nothing written.
  - "refused states": each → 409 and `commits == 0`: target `partial`; target `open` with a line `ReceivedQuantity: 1`; target `draft`, `pending_approval`, `closed`; source `open`, `pending_approval`, `closed`.
  - "stale versions": target version mismatch and draft version mismatch each → 409.
  - "the same order twice" → 400. Unknown draft id → 404.
  - "a different supplier on the draft is allowed": target supplier stays (no supplier change written).
  - "a combined note over the limit writes nothing": 400, `commits == 0`.

- [ ] **Step 2: Write the failing database tests** in `purchase_orders_integration_test.go` using `newFixture`, `f.order`, `f.product`; a drafter fixture is `asUser(constant.DraftPurchaseOrderPermissionKey)`. Tests:
  - `TestMergeDraftIntoOpenOrder`: manager creates open order A (Alpha 10, Beta 5, supplier "Emzor"); a drafter creates draft B (Alpha 15, Gamma 3). After `MergePurchaseOrder(ctx, A, {A.Version, B, B.Version})`: A is `draft`, `approved_by_id`/`approved_at` null, supplier "Emzor", lines Alpha 15, Beta 5, Gamma 3; B is `closed`; audit rows with action `merged` exist for both; A's version increased.
  - `TestMergeRefusesAnOrderWithReceivedStock`: receive against A first (`f.receive`) then merge → 409; both orders unchanged.
  - `TestConcurrentMergesOfOneDraftCommitOnce`: two goroutines merge draft B into two different open orders; exactly one succeeds, the other is 409; B ends up closed once and only the winner's target changed (pattern: `TestConcurrentTransfersCreateOneOrder`).
  - `TestMergedOrderGoesThroughApprovalAgain`: after the merge, submit then approve A as a manager → status `open`.

- [ ] **Step 3: Run** `cd backend && go test ./service/ -run 'TestMergePurchaseOrder' -v` and the integration tests (`PHARMACY_TEST_DSN=... go test ./service/ -run 'TestMerge|TestConcurrentMerges' -v`). Expected: build failure.

- [ ] **Step 4: Implement.**
  - `revertPurchaseOrderToDraftQuery` in `query.go`: set `status = 'draft'`, `approved_by_id = NULL`, `approved_at = NULL`, `version = version + 1`, `updated_at = NOW()` where `id = $1`; repo method `RevertPurchaseOrderToDraftTx`; add to the repository interface.
  - `MergePurchaseOrder` in `purchase_order_merge.go`: check `canWorkOnPurchaseOrders` (403 "you cannot merge purchase orders"); reject `req.DraftID == id` (400) and `req.DraftID <= 0` (400); run inside `s.inTx(ctx, "failed to merge purchase orders", ...)`; lock both rows with `GetPurchaseOrderForUpdateTx` **in ascending id order**, mapping `sql.ErrNoRows` to `purchaseOrderNotFound`; check versions; check statuses and "nothing received" (409 messages: "only an open order with nothing received can have a draft merged into it", "only a draft can be merged into another order"); build the plan; write in this order: header (target's own supplier and the plan's note via `UpdatePurchaseOrderHeaderTx`), line updates, inserts, `RevertPurchaseOrderToDraftTx(target)`, `ClosePurchaseOrderTx(draft)`, then the two audit entries.
  - Audit details: target gets `{"merged_purchase_order_id": DraftID, "status_before": "open", "added": auditItemSummaries(inserts), "raised": [{"name", "quantity": change{From,To}}]}`; the draft gets `{"merged_into_purchase_order_id": id}`.
  - Add the method to `service/interfaces.go`.

- [ ] **Step 5: Run** the unit tests, the integration tests, then `go vet ./... && go test ./service/ -count=1`. Expected: PASS; the integration tests skip without `PHARMACY_TEST_DSN`, so run them once with it set.

---

### Task 3: HTTP route

**Files:**
- Modify: `backend/adapter/http/controller/inventory.go`
- Modify: `backend/adapter/http/router/inventory.go`

**Interfaces:**
- Consumes: `service.InventoryService.MergePurchaseOrder`, `types.MergePurchaseOrderRequest` (Task 2).
- Produces: `POST /inventory/purchase-orders/{id}/merge`, body `{version, draft_id, draft_version}`, reply 200 `{"id": <open order id>}`; errors as the other order actions.

- [ ] **Step 1: Implement** `func (c *inventoryController) MergePurchaseOrder(w http.ResponseWriter, r *http.Request)`: parse `{id}` (400 "invalid purchase order id"), decode the body (400 "invalid json"), call the service with `requestUserID(r)`, `respondServiceError(w, err, "could not merge purchase orders")`, reply `helper.JSONResponse(w, http.StatusOK, map[string]int{"id": id})`. Register the route in the router next to `withdraw`, wrapped in `mayDraft`.

- [ ] **Step 2: Verify** `cd backend && go build ./... && go vet ./...`. Then start the server (`make run` against the local Postgres) and, after Task 4, check through the page; a direct check without a session is not meaningful because of auth and CSRF.

---

### Task 4: The page

**Files:**
- Modify: `frontend/src/next/views/PurchaseOrders.vue`

**Interfaces:**
- Consumes: `POST ${API}/inventory/purchase-orders/{id}/merge`; `GET ${API}/inventory/purchase-orders?view=drafts` (the existing list endpoint, as the Drafts view uses it; read how `loadOrders` builds the query before using it).

- [ ] **Step 1: Add `canMergeInto(o)`** next to `canWithdraw`: `o.status === "open"`, every line's `received_quantity === 0`, and `canDraft.value || canApprove.value`.

- [ ] **Step 2: Add the button and inline panel**, following the `moveTarget` panel's structure: a **Merge a draft in** button (icon `Merge` from lucide, or `GitMerge` if that is what the installed lucide version offers) in the actions row after "Pull back to draft"; a `mergeTarget` ref (like `moveTarget`, mutually exclusive with `closeTarget` and `moveTarget`). Opening the panel loads the drafts (excluding the order itself) and shows them in a select; if there are none it says "There are no drafts to merge." The panel shows a preview of the combined result for the chosen draft and a supplier warning when the draft's `supplier_id` differs from the order's and the draft names a supplier.

- [ ] **Step 3: Preview function** `mergePreview(target, draft)` returning rows `{name, kind: "new" | "raised" | "same", from, to}`, with exactly the server's match rule (product id when both have one, else lowercased name + manufacturer when neither has one) and the higher-quantity rule. The server stays authoritative; add a comment saying the two must match.

- [ ] **Step 4: Confirm and send.** The panel's confirm button calls `confirm({ title: "Merge into PO-#?", message: "It goes back to a draft and needs approval again. The draft you picked is closed.", confirmLabel: "Merge", busyLabel: "Merging...", cancelLabel: "Cancel", run })` where `run` posts `{version: o.version, draft_id, draft_version}` with `csrfFetch` and throws `new Error(await apiError(r, "Could not merge the orders."))` on failure (on 409 also close the panel and reload, as `moveOutstanding` does). On success show a notice ("PO-# merged. It is a draft again: edit it, then submit it for approval.") and `loadOrders()`.

- [ ] **Step 5: History text.** In the audit-line function (next to `e.action === "moved"`), add `merged`: for the open order "Draft PO-N merged in (X lines added, Y quantities raised)"; for the draft "Merged into PO-M and closed" (distinguish by `d.merged_into_purchase_order_id` versus `d.merged_purchase_order_id`).

- [ ] **Step 6: Build and look.** `make assets` (Tailwind and Vite), restart `make run`, then in the browser: create an open order as a manager, create a draft as a cashier, merge, and check the preview, the result, the history of both orders, the empty state, a different-supplier draft, and a stale click (merge the same draft twice from two tabs). Report what was seen, including anything not exercised.

---

### Task 5: Docs and board

**Files:**
- Modify: `docs/features/purchase-orders/guide.md`, `docs/features/purchase-orders/technical.md`, `CLAUDE.md`

- [ ] **Step 1:** Add a **Merge a draft into an open order** task to the staff guide (who, when it applies, what happens to quantities, that approval is needed again, what you see in history) and the endpoint, rules, audit shape and error cases to the technical doc. Use the `document-feature` skill's conventions if it applies. Add a short sentence about merge to the purchase-orders bullet in `CLAUDE.md` Gotchas.
- [ ] **Step 2:** Run `make check-go-format check-go-lines`, `cd backend && go test ./... -count=1` (with `PHARMACY_TEST_DSN` set), then review with the `review-ticket` skill before asking for the single commit-and-push approval.
- [ ] **Step 3:** Update the board card: `waypoint move T-138 "In Progress" --project Primocrest` when starting, `waypoint done T-138 --project Primocrest` with `prs=#N` once the pull request exists.
