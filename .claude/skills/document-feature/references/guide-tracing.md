# Trace a feature for the staff guide

Load before writing or refreshing any `guide.md`.

The guide follows a person through the shop, not the code. Trace it by reading the screens, then confirm each rule the screen implies in the backend, because the backend is what actually stops or allows things.

## Who and where

- **Who can use it**: the roles that have the permissions the screen and routes check (`internal/constant/permissions.go`, `PermissionGate.vue`, `usePermissions.js`, route gating in `adapter/http/router/`). Say it in role words: "Managers and admins can approve; cashiers can draft." If something is open to everyone who can log in, say so.
- **Where to find it**: the sidebar entry (`components/Sidebar.vue`), the page title (`router.js` `meta.title`), and any other way in (dashboard cards, badges, links from other pages).

## The journeys

List the tasks a person does with the feature ("Make a sale", "Hold a sale and finish it later", "Return items"). For each task, trace step by step through the `.vue` view and its components:

- what they click or type, using the **exact** label from the source
- what the screen shows back: messages, colours, badges, statuses, confirmation dialogs (`useConfirm` calls), the printed or exported output
- shortcuts that exist in the code (keys handled in the component)

## What happens next

The effects someone in the shop would notice, confirmed in the service code:

- stock goes up or down, and where it shows (Products, Low Stock, Expiring)
- totals and reports that change (Sales History total, dashboard)
- things created for later (held sale, draft order, pending approval badge)
- what other people see (an approver's badge, a stock count updating live)

## Rules and limits

Every validation, as a plain sentence a cashier can act on: "You can't receive more than was ordered", "Expiry dates under six months away need you to tick the confirmation box". Find them in the screen first, then confirm the backend enforces the same rule; if only the screen checks it, the guide still states the rule, and the technical doc notes the gap.

## When something goes wrong

For each error message the screens can show: what it means and what to do. Include the common cases: lost connection mid-save (is it safe to retry? idempotency keys say), someone else changed the same record, missing permission.

## Language

Second person, short sentences, the shop's own words (receipt, till, stock, supplier). Money as ₦ with two decimals. No code, file names or technical terms; those belong in the technical doc. Screenshots only if the owner asked; take them from the running app with made-up data.
