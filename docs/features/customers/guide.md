# Customers: staff guide

## What it's for

Keep a list of the pharmacy's customers: their name, phone number, email, any notes, and whether they agreed to be contacted. Today the list is only kept and managed on its own page; the till does not use it yet.

## Who can use it

| Role | Can do |
|---|---|
| Cashier | See the list, add customers, change their details, deactivate and reactivate them |
| Manager | The same |
| Admin | The same |

A user whose role has neither permission does not see **Customers** in the menu. An administrator can change who has access in Administration → Permissions.

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

1. Choose the customer's name in the list.
2. Change what you need and choose **Save changes**.

**You'll see:** "Saved <name>." If nothing was changed, nothing is saved.

When you change **Agreed to be contacted by the pharmacy**, the panel shows the date it last changed. Untick it if a customer asks not to be contacted.

### Deactivate a customer

Use this for someone who should no longer be on the list, for example a duplicate entry or a customer who asked to be left out.

1. Choose the customer's name, then **Deactivate**.
2. Confirm in the box that opens.

**You'll see:** "<name> deactivated." They disappear from the list unless **Show deactivated** is ticked, where they carry a **Deactivated** label. To bring them back, choose them, then **Reactivate**.

## What happens next

- Every add, change, deactivation and reactivation is recorded with who did it and when. Whether a customer agreed to be contacted is recorded each time it changes.
- Nothing else in the shop changes yet: sales, receipts and the till do not read this list. The name typed in the till is still only printed on the receipt.

## Rules and limits

- A name is required, up to 120 characters. Notes can be up to 1,000 characters.
- Each phone number belongs to one customer. This includes deactivated customers, so a number is never handed to someone else by mistake. A second person who shares a number can be added without a phone, with a note.
- A Nigerian phone number has 10 digits after the first 0 (11 with the 0): `0803 123 4567`. A number from another country must start with `+` and the country code.
- Customers are never deleted, only deactivated.
- You can see the list with permission to see customers; adding and changing needs permission to manage customers. Someone who can only see customers sees the details but no **Add customer**, **Save changes** or **Deactivate** buttons.

## When something goes wrong

| You see | What it means | What to do |
|---|---|---|
| Enter the customer's name. | The name box is empty. | Type a name. |
| Enter a Nigerian number with 11 digits starting 0 (e.g. 0803 123 4567), or a number starting with + and the country code | The phone number has the wrong number of digits or a stray character. | Check the number with the customer and type it again. |
| This phone number is already saved for <name> | Another customer already has this number. | Search for that customer and change them instead. |
| This phone number is already saved for <name> (deactivated; reactivate them instead) | The number belongs to a deactivated customer. | Search with **Show deactivated** ticked, open them and choose **Reactivate**. |
| Enter a valid email address, or leave it empty | The email is not written like an address. | Correct it or clear the box. |
| This customer was changed by someone else; reload and try again. Your changes are still in the form. Save again to replace theirs, or close this panel to keep theirs. | Another person saved a change first. | Your typing is kept. Look at the list to see what they changed, then either save again to replace their change or close the panel to keep theirs. |
| You are not allowed to change customers | Your role may see customers but not change them. | Ask a manager or an administrator. |
| Could not reach the server. Check the connection, then check the list before trying again. | The connection dropped while saving. | Look in the list to see whether the customer was saved before you add them again. |

## Questions people ask

**Can I delete a customer?** No. Deactivate them. Their record stays so that future sales can still refer to them.

**Does the customer get messages from the system?** No. The tick box only records what the customer said, for the shop's own use.

**Why can't I see the customer's name at the till?** The till does not use this list yet. It is planned in T-118.

## Related

- Planned next: attach a customer to a sale (T-118), purchase history and totals (T-119), import and merge duplicates (T-120), erase or anonymize on request (T-121).
