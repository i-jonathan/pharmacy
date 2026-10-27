- [x] On default, focus on search bar in sales receipt page
- [x] Add price display to search result
- [x] Update search query to search from beginning of word and not anywhere
- [x] Clearly indicate that search result is scrollable
- [x] Allow scrolling by keyboard in search result
- [x] Improve autofocusing into text inputs in sales receipts page
- [x] Add Print receipt functionality
- [x] Add Daily total to Sales History page
- [x] Add Hold Sale Feature
- [x] Enable delete held sale & add links to held sale where necessary
- [x] Add Warning to Sales Receipt page when cart isn't empty and an attempt to navigate away happens
- [ ] Consider adding modal on Sales receipt after save when change > 0
- [x] Accept Return Feature
- [x] Add Inventory List page
- [ ] Stock taking
## Purchase orders / audit follow-ups
- [ ] Audit log for everything: extend `audit_log` beyond purchase orders to sales and returns, stock movements and adjustments, stock taking, product price and cost edits, receiving, and user/role/permission changes. Add a shared `auditEntry` helper in the service layer, write each entry inside the same transaction as the change, and add an admin screen to browse and filter it
- [ ] Capture batch/lot numbers on receiving (`product_batch.batch_no` exists but is never written) and allow several batches (different expiry dates) for one purchase order line
- [ ] Costs on purchase orders: expected cost per line, order total, and compare with the received cost
- [ ] Purchase order history: search and filter by supplier, status and date (the Completed tab is capped at 100)
- [ ] Show on-order quantities on the Products list and the Low Stock page
- [ ] Supplier management screen: rename, merge duplicates, contact details
- [ ] Decide whether receiving against a purchase order should need its own permission (today anyone who can receive can)
- [ ] Browser/end-to-end tests for the purchase order screens and the receive-against-order flow
- [ ] Idempotency key for creating a purchase order (the Save button is disabled while saving, but a retry after a lost response can still create a duplicate order)
- [ ] Share one "lock the order, check version and status" helper between edit, close, move and receive (the four copies can drift apart)
- [ ] Pagination for the Completed purchase orders list (capped at the newest 100) and decide whether the read endpoints (list, get, history) need a permission
- [ ] Purchase order approval: let a drafter withdraw a submitted order back to draft, let a manager send it back for changes (instead of only rejecting), and notify the drafter when their order is approved or rejected (today they see it on their list)
- [ ] Real notifications for approvers (stored inbox, push or email) instead of the sidebar badge and dashboard banner, which poll once a minute

- [ ] Changing or clearing the supplier on an order that already has receipts is allowed silently (the receipts stay with the old supplier). Block it and point to Move outstanding
- [ ] Show a rejected draft as "Rejected" with its reason on the order itself, not only "Closed" with the reason in the activity log
- [ ] Migration 43 matches roles by name (Cashier, Manager, Admin); warn, or grant by a configured role, when a deployment's roles are named differently

## Make configurable (settings)
Values that are hard-coded today and should become settings, edited by an admin in a settings screen and stored in the database (read once per session and served to the SPA, so the frontend and backend share one value).

- [ ] Large purchase order confirmation: the editor asks "Save it anyway?" when a line is 100,000 units or more, or an order has more than 200 items (`CONFIRM_QUANTITY` and `CONFIRM_LINES` at the top of `frontend/src/next/views/PurchaseOrders.vue`). There is no hard line limit; the only ceiling is the 32-bit quantity column
- [ ] Near-expiry alert on the Receive Items page: an expiry date under 183 days (about 6 months) away turns the field amber and the user must tick "I confirm the near-expiry dates" before receiving (`expiryClass`, `expiryTitle` and `hasNearExpiry` in `frontend/src/next/views/ReceiveItems.vue`). Dates in the past are always flagged red
- [ ] Expiring-soon window on the dashboard and expiring list: 90 days (`INTERVAL '90 days'` in `backend/repository/dashboard.go`). Probably should share a setting with the Receive Items alert, or be set separately
- [ ] Company name: "Primocrest Pharmacy" is repeated in `Sidebar.vue`, `LoginPage.vue`, `NotFoundPage.vue`, the receipt printed by `composables/usePos.js` and the `<title>` of `login-v2.html`, `next-dashboard.html` and `error.html`
- [ ] Logo or icon, plus receipt header details (address, phone), to replace the text-only brand mark
- [ ] Colour scheme: brand colours live in the Tailwind theme tokens and `design-system/primocrest-pharmacy/MASTER.md`; let a pharmacy pick a primary colour (status colours stay fixed so green, amber and red keep their meaning)
- [ ] Other numbers worth a setting once there is a screen for them: the approval badge poll interval (60 seconds, `usePendingPurchaseOrders.js`) and the Completed purchase orders cap (100, `purchaseOrderHistoryMax`)
