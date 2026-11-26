# Customers: staff guide

## What it's for

Keep a list of the pharmacy's customers: their name, phone number, email, any notes, and whether they agreed to be contacted. The list is managed on its own page, and the cashier picks a customer from it at the till so the sale, Sales History and the receipt know who bought.

## Who can use it

| Role | Can do |
|---|---|
| Cashier | See the list, add customers, change their details, deactivate and reactivate them |
| Manager | The same |
| Admin | The same |

Fixing the customer on a sale that is already saved: a cashier can do it for a sale made in the last 24 hours. A manager or an admin can do it for a sale of any age.

A user whose role has neither permission does not see **Customers** in the menu and has no customer box at the till: their sales are walk-in sales. Picking a customer at the till needs permission to see customers; adding a new one there needs permission to manage them. An administrator can change who has access in Administration → Permissions.

## Where to find it

Open **Sales** in the menu on the left and choose **Customers**.

## Tasks

### Add a customer

1. Choose **Add customer**.
2. Type the **Name**. This is the only thing you must fill in.
3. Type the **Phone** if the customer gives one, in any usual way: `0803 123 4567`, `803 123 4567` or `+234 803 123 4567` all mean the same number.
4. Fill in **Email** and **Notes** if useful.
5. Tick **Agreed to be contacted by the pharmacy** only if the customer said they are happy to be contacted.
6. Choose **Add customer** at the bottom of the panel.

**You'll see:** the new customer in the list, and "Added <name>." at the top of the page.

### Find a customer

Type in **Search by name or phone**. Letters search names. Numbers search phone numbers: a number written from its start (`0803`, `+234 803`) finds numbers that begin that way, and digits from the middle of a number (`803`) find any number containing them. Tick **Show deactivated** to include customers who were deactivated.

### Change a customer's details

1. Choose the customer in the list (anywhere on their row).
2. Change what you need and choose **Save changes**.

**You'll see:** "Saved <name>." If nothing was changed, nothing is saved.

When you change **Agreed to be contacted by the pharmacy**, the panel shows the date it last changed. Untick it if a customer asks not to be contacted.

If you have typed changes and then open another customer, choose **Add customer** or close the panel, the page asks **Discard your changes?** so typing is not lost by a stray click. Leaving the page asks the same.

### See a customer's health readings

Managers and admins see a **Health readings** tab beside **Details** in the panel, with the customer's saved blood pressure and glucose readings, and can reprint or correct them there. See the [Health checks guide](../health-checks/guide.md).

### Put a customer on a sale

1. At the till, in **Current sale**, choose **Add customer** (the box says **Walk-in customer** until you do).
2. Type the customer's phone number or name in **Search by phone or name** (two characters or more). Part of a phone number works: `0803 555` finds numbers that start that way.
3. Choose the customer in the list. If only one customer matches (for example a full phone number), press Enter to take them.
4. Finish the sale as usual.

**You'll see:** the customer's name and phone in place of **Walk-in customer**. **Change** picks someone else; the **x** takes the customer off the sale. After the sale the box goes back to **Walk-in customer** for the next one.

If the customer is not on the list, choose **Add "<what you typed>" as a new customer**, fill in the name (and phone if they give one) and choose **Save and add to sale**. They are saved on the Customers page and put on the sale in one step. If that phone number already belongs to someone, the box says so and lists that customer so you can pick them.

A customer who was deactivated still appears in the search, greyed out and marked **Deactivated**, below the active matches. You cannot put them on the sale as they are. If you have permission to manage customers, choose **Reactivate** beside them and confirm: they are reactivated and put on the sale in one step. Otherwise the row says **Ask a manager**. Pressing Enter does not take a deactivated customer, even when they are the only match.

A held sale keeps its customer. If that customer was deactivated while the sale was held, the sale shows them marked **Deactivated** as soon as you resume it, and **Complete** stays off until you choose **Reactivate** or remove them with the **x**.

A held sale keeps its customer. A sale with no customer is a walk-in: nothing is printed for the customer.

### See who bought

Choose a sale in **Sales History**: **Customer** shows beside the cashier and the date. **Reprint receipt** prints the copy with **Customer: <name>**.

### Change or remove the customer on a saved sale

Use this when the wrong customer was put on a sale and the sale is already saved.

1. Open **Sales History** and choose the sale.
2. In the sale details, beside **Customer:** (or **Walk-in customer**), choose **Change**.
3. In **Customer on this sale**, choose **Change** and search for the right customer, or choose **x** to take the customer off so the sale becomes a walk-in sale. On a walk-in sale, choose **Add customer**.
4. Under **Reason**, choose **Picked by mistake**, **Customer asked** or **Other**. You can add a few words of detail. For **Other** the words are required.
5. Choose **Save**.

**You'll see:** "Customer changed to <name>" (or "Customer removed from the sale") at the bottom of the page, and the new name in the sale details. **Reprint receipt** now prints the new customer.

**Save** stays off until you have chosen a different customer and a reason. **Change** only shows if you may see and manage customers. For a sale more than 24 hours old it shows only for a manager or an admin. A deactivated customer cannot be put on a sale, but one who is already on the sale can be taken off. A manager can reactivate a deactivated customer from the search, the same way as at the till. Reactivating happens at once, even if you then close the box without saving; only the sale waits for **Save**.

### See who changed a sale's customer

Every time the customer on a sale is changed or removed, the sale keeps a record of it.

1. Open **Sales History** and choose the sale.
2. If its customer was ever changed, a **Customer Changes** list shows under the items, newest first.

**You'll see:** for each change, the old customer, an arrow and the new customer (**Walk-in** when there was none), then who made the change and when, the reason, and the words they added if any. A sale whose customer was never changed has no such list.

Only people who may see customers see this list. Customer names in it are the names as they are today: if a customer has been renamed since, the new name shows. A change you have just saved appears in the list at once.

### Deactivate a customer

Use this for someone who should no longer be on the list, for example a duplicate entry or a customer who asked to be left out.

1. Choose the customer's name, then **Deactivate**.
2. Confirm in the box that opens.

**You'll see:** "<name> deactivated." They disappear from the list unless **Show deactivated** is ticked, where they carry a **Deactivated** label. To bring them back, choose them, then **Reactivate**.

## What happens next

- Every add, change, deactivation and reactivation is recorded with who did it and when. Whether a customer agreed to be contacted is recorded each time it changes.
- A sale with a customer shows that customer in Sales History, and the receipt (and every reprint) prints **Customer: <name>**. A walk-in sale prints no customer line.
- Only people who may see customers see the customer on Sales History and reprints.
- Every change of the customer on a saved sale is recorded with who did it, when, the reason, the receipt number, and which customer it was and became. Nobody can see that record on a screen yet (T-127).

## Rules and limits

- A name is required, up to 120 characters. Notes can be up to 1,000 characters.
- Each phone number belongs to one customer. This includes deactivated customers, so a number is never handed to someone else by mistake. A second person who shares a number can be added without a phone, with a note.
- A Nigerian phone number has 10 digits after the first 0 (11 with the 0): `0803 123 4567`. A number from another country must start with `+` and the country code.
- Customers are never deleted, only deactivated. A deactivated customer is shown at the till marked **Deactivated** and cannot be put on a sale until they are reactivated.
- A cashier can change the customer only on a sale made in the last 24 hours, counted from the moment of the sale, not from midnight: a sale made at 11:59 pm can still be fixed after midnight. A sale more than 24 hours old needs a manager or an admin. If a change was saved but the reply was lost, trying again is answered as saved, even if the sale has passed 24 hours since.
- Changing the customer does not change the sale's items, payments, stock or totals.
- Someone who cannot see customers prints a reprint without the customer line. If they resume a held sale that has a customer, the till says it will be recorded as a walk-in; holding it again keeps the customer for someone who can see customers.
- You can see the list with permission to see customers; adding and changing needs permission to manage customers. Someone who can only see customers sees the details but no **Add customer**, **Save changes** or **Deactivate** buttons.

## When something goes wrong

| You see | What it means | What to do |
|---|---|---|
| Enter the customer's name. | The name box is empty. | Type a name. |
| Enter a Nigerian number with 11 digits starting 0 (e.g. 0803 123 4567), or a number starting with + and the country code | The phone number has the wrong number of digits or a stray character. | Check the number with the customer and type it again. |
| This phone number is already saved for <name> | Another customer already has this number. | Search for that customer and change them instead. |
| This phone number is already saved for <name> (deactivated; reactivate them instead) | The number belongs to a deactivated customer. | At the till, they now show in the list marked **Deactivated**: choose **Reactivate**. On the Customers page, search with **Show deactivated** ticked, open them and choose **Reactivate**. |
| The name, email and notes can't contain control characters | Something invisible was pasted into a box. | Retype the text instead of pasting it. |
| Enter a valid email address, or leave it empty | The email is not written like an address. | Correct it or clear the box. |
| This customer was changed by someone else; reload and try again. Your changes are still in the form. Save again to replace theirs, or close this panel to keep theirs. | Another person saved a change first. | Your typing is kept. Look at the list to see what they changed, then either save again to replace their change or close the panel to keep theirs. |
| This customer has been deactivated; remove them from the sale, or ask a manager to reactivate them | The customer was deactivated after you picked them, while the sale was open. The sale was not saved. | Choose **Reactivate** beside the customer in the sale (or remove them with the **x**), then complete the sale. |
| This sale was already saved with a different customer; check Sales History before selling again | The sale was saved a moment ago (the connection dropped before you saw it) and you have since changed the customer on this screen. | Open Sales History, find the sale and check it. If the customer is wrong, change it there. Do not sell it again. The till empties the sale for you, because it was already saved. |
| You are not allowed to attach customers to a sale | Your role cannot see customers. | Sell as a walk-in, or ask an administrator. |
| You are not allowed to change customers | Your role may see customers but not change them. | Ask a manager or an administrator. |
| You are not allowed to change a sale's customer / You are not allowed to see or change a sale's customer | Your role cannot manage customers, or cannot see them. | Ask a manager or an administrator. |
| Only a manager can change the customer on a sale more than 24 hours old | The sale was made more than 24 hours ago and you are a cashier. | Ask a manager to change it. |
| Someone else changed this sale's customer; check who is on it now before changing it again | Another person changed the customer after you opened the sale. Nothing was saved. | Sales History reloads and shows the customer as it is now. Check it, and change it again only if it is still wrong. |
| Choose a reason for the change / Say why the customer is changed | No reason was chosen, or **Other** has no words. | Choose a reason, and write a few words for **Other**. |
| This customer has been deactivated; choose another, or ask a manager to reactivate them | You picked a customer who was deactivated. | Pick someone else, or ask a manager to reactivate them. |
| Customer was not found | The customer or the sale no longer exists. | Search again. |
| Could not reach the server. Check the connection, then check the list before trying again. | The connection dropped while saving. | Look in the list to see whether the customer was saved before you add them again. |

## Questions people ask

**Can I delete a customer?** No. Deactivate them. Their record stays so that future sales can still refer to them.

**Does the customer get messages from the system?** No. The tick box only records what the customer said, for the shop's own use.

**Does the customer's name always print on the receipt?** Only when a customer is on the sale. Their phone number is never printed.

**I picked the wrong customer and already completed the sale.** Open the sale in Sales History and choose **Change** beside the customer. See "Change or remove the customer on a saved sale". A cashier can do this within 24 hours of the sale; an older sale needs a manager.

**Can I see who changed a customer on a sale?** Yes, if you can see customers: open the sale in Sales History. See "See who changed a sale's customer".

## Related

- [Health checks](../health-checks/guide.md): blood pressure and glucose readings kept for customers who agree.
- Planned next: purchase history and totals (T-119), import and merge duplicates (T-120), erase or anonymize on request (T-121), see who changed a sale's customer (T-127).
