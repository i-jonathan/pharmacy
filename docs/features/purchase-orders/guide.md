# Purchase orders: staff guide

## What it's for

Purchase orders are how the shop plans what to reorder. An order is built from customer requests, items running low, or simply what the shop decides it needs. Once it's agreed, it's exported as a PDF and sent to the supplier, mostly on WhatsApp. When the delivery comes, it's received against the order, so you can see what has arrived and what is still outstanding.

## Who can use it

| Role | Can do |
|---|---|
| Cashier | Start an order, edit it while it's still a draft, and submit it for approval. Sees only their own orders. |
| Manager, Admin | Everything a cashier can, plus: edit any draft or open order, approve or reject, export the PDF to send, receive deliveries against an order, close an order, and move what's still outstanding to another supplier. Sees everyone's orders. |

Anyone without either permission sees "You do not have permission to use purchase orders. Ask an administrator for access."

## Where to find it

- **Sidebar → Inventory → Purchase Orders.** Managers see a small number next to it when orders are waiting for their approval.
- **Dashboard:** managers see a banner, "*N* purchase orders waiting for your approval", that opens the approval list.
- **Low Stock page → Create purchase order** starts a new order with the low stock list already open.

## How the shop uses it

Every week, sometimes several times a week, a cashier starts an order from what's needed, what's in stock, what's already on order, and any customer requests. The manager then checks it: adds or changes quantities, removes items the shop doesn't want, approves it, and sends it to the supplier. The shop works mainly with three suppliers. The number of items on an order varies every time.

## Tasks

### Start an order (cashier or manager)

1. Click **New purchase order**.
2. Add items in any of three ways:
   - **Search products you already stock**: type at least two letters and click the product. If it's already on another open order, you'll see "*N* on order" beside it.
   - **Choose from low stock**: shows everything at or below its reorder level, with "In stock *N* · reorder at *N*" and how much is already on order. Click a row (or tick it) to select it; the quantity starts at what would bring it back to its reorder level, after counting what's already on order. Adjust it, then click **Add *N* selected**. **Select all** and **Clear** work on the whole list.
   - **Not in your products yet?** Type the **New item name**, optionally a **Manufacturer** and a note, set the quantity and click **Add item**. If the name matches a product you already have, it's added as that product instead. Items added this way show a **New product** tag until they're registered when they arrive.
3. For any item, click **+ Add note** to add details such as brand, strength or pack size.
4. Change quantities with the − and + buttons, or click **Remove** to take an item off.
5. Optionally choose the **Supplier** (you can leave it until the delivery arrives) and add **Additional notes** such as delivery instructions or payment terms.

### Save and submit (cashier)

- **Save as draft** keeps it so you can come back to it. It appears under **My drafts**.
- **Save and submit for approval** sends it to a manager. Once submitted you can't change it; the card says "Submitted. A manager will review it; you cannot change it now."
- A draft you saved earlier can be submitted from its card with **Submit for approval**.

### Check and approve (manager)

1. Open the **Needs approval** tab. It lists drafts still being written and orders submitted to you.
2. Click **Edit** to check the order: change quantities, remove items you don't want, add anything missing.
3. Click **Save and approve**, or **Approve** on the card. You'll see "approved. It is now open: export the PDF to send it, then receive against it."
4. To turn an order down, click **Reject**, choose a reason (or **Other** with your own words) and click **Yes, reject**.

Orders a manager creates are approved straight away; they don't need a second approval.

### Send it to the supplier (manager)

1. On the **Waiting for stock** tab, click **PDF** on the order's card (or in **View**).
2. Send the downloaded file to the supplier on WhatsApp, or whichever way they prefer.

The PDF shows the order number (for example PO-00012), the supplier, the date, any notes and each item with its quantity and notes. Drafts are stamped DRAFT, so a printed copy can't be mistaken for an approved order.

### Receive the delivery (manager)

1. On the order's card, click **Receive**. This opens **Receive Items** with the order loaded: "Receiving against PO-*NNNNN*".
2. Quantities are filled in with what's still outstanding. Change them to what actually arrived, and remove anything that didn't come; it stays outstanding on the order.
3. Items marked "Not in your products yet" must be registered first: click **Create product**, or **Link existing product** and pick it in the search.
4. Finish the receipt as usual (supplier, expiry dates, cost). If the delivery is from a different supplier than the one on the order, you'll see a note saying the delivery will be recorded as coming from them.

### When the supplier can't deliver everything (manager)

- **Move outstanding** sends the undelivered quantities to a different supplier. Choose the new supplier and a reason, then click **Move outstanding**. The original order is closed with what it received, and a new order is created for the rest.
- **Close order** stops waiting for the remaining units ("Stop waiting for the remaining *N* units?"). An order with nothing received shows **Cancel order** instead. Both need a reason.

### Look at an order

Click **View** for the full list of items with Ordered, Received and Outstanding, filters for **All**, **Outstanding** and **Complete**, and a search box. **Activity** shows who did what and when, including reasons.

## What happens next

- **Approving** moves the order to **Waiting for stock**. Nothing changes in stock yet.
- **Receiving** adds the stock to your products straight away and updates the order's progress bar ("*N* of *N* units received"). The order shows as **partial** until everything arrives, then **received** and moves to **Completed**.
- Items ordered but not delivered show as "on order" when anyone builds the next order, so the same thing isn't ordered twice by mistake.
- **Closing** or **moving** keeps everything already received; only the outstanding part is dropped or moved.
- Every step (created, edited, submitted, approved, rejected, received, closed, moved) is recorded with who did it and when.

## Rules and limits

- An order needs at least one item, and each quantity must be a whole number of at least 1.
- The same product can't be on one order twice.
- Very large orders (any item at 100,000 or more, or more than 200 items) ask "This is a very large order" before saving, so an extra zero doesn't slip through.
- Once some of an item has been received, its quantity can't go below what arrived, and it can't be removed.
- Closing, cancelling, rejecting and moving all need a reason.
- Only approved orders can be received against. Drafts and submitted orders can't.
- A cashier can only change their own order, and only while it's a draft.
- Notes on an order are limited to 1,000 characters, and notes on an item to 500.

## When something goes wrong

| You see | What it means | What to do |
|---|---|---|
| "This order was changed by someone else (or received against) since you opened it. Your edits were not saved." | Someone saved or received on the same order while you had it open. | Click **Reload latest**, then make your changes again. |
| "*Item* is already on this order." | You tried to add a product twice. | Change the quantity on the existing line instead. |
| "*Item*: *N* already received, so the quantity cannot be lower." | You lowered a quantity below what has already arrived. | Set it to at least that amount, or close the order if the rest won't come. |
| "Could not load purchase orders." | The app couldn't reach the server. | Check the connection and click **Retry**. |
| "Discard your changes?" | You're leaving the editor with unsaved changes. | **Keep editing**, or **Discard changes** if you meant to leave. |

## Questions people ask

**I submitted the wrong quantities. Can I fix them?** Not after submitting. Ask a manager to edit the order before approving it, or to reject it so you can start again.

**Can I send the order to the supplier before it's approved?** A draft's PDF is stamped DRAFT. Wait for approval, then send the approved PDF.

**Why don't I see other people's orders?** Cashiers only see the orders they created. Managers see all of them.

## Related

- Receive Items: recording deliveries, including against an order
- Low Stock: what's at or below its reorder level
- Products: registering new items when they arrive
