# T-018 Lock down public user registration

Card: BOARD.md T-018 · Branch: fix/T-018-lock-down-registration

## Goal

Stop strangers from creating accounts, and give admins a proper way to add staff. Today anyone who can reach the server can create a working account, and that same unprotected endpoint is also the only way to create a user at all.

## What I found

- Verified on the card: `POST /user/register` has no auth (`code`: `router/user.go` `InitUserRouter`; `main.go` mounts `/user/` with no `AuthMiddleware`).
- New accounts get the Cashier role via the column default (`code`: migration 000023), and Cashier holds `purchaseorder:draft` (`code`: 000043). So a stranger gets a logged-in cashier account.
- CSRF doesn't stop it (`code`): `GET /csrf-token` is public (`main.go`), so a script can fetch a token and its cookie, then post the form. The card was right.
- Nothing in the app calls the endpoint any more (`code`: no references in `frontend/src` or `backend/template`). The old register page was removed with the legacy server-rendered UI. The Administration → Users screen can change roles and reset passwords but can't add a user (`code`: `components/admin/UsersManager.vue`, `router/admin.go`).
- No first-admin path (`code`): no migration, script or command creates an admin. A new install has no way in except this public endpoint plus manual SQL to set the role.
- Two bugs in the current create path (`code`: `service/user.go` `CreateUserAccount`):
  - a failed insert is logged and then reported as success
  - a database error during the duplicate check is reported as "username already exists"
- Password rules (`code`): `model.User.ValidatePassword` requires 8 or more characters, and the admin reset in `service/admin.go` uses the same 8.
- NIST SP 800-63B rev 4 (`web`) requires single-factor passwords to be at least 15 characters and to be checked against a blocklist. It allows a minimum of 8 only when the password is part of multi-factor login, and says not to impose composition rules. This app is single-factor, so 8 falls short.
- Patterns to reuse (`code`):
  - admin routes behind `admin:access` in `main.go`
  - `requestUserID` for the actor
  - `audit_log` written in the same transaction (`service/purchase_order_audit.go` `auditPurchaseOrder`)
  - `csrfFetch` in the admin components

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| Remove the public endpoint | `code` | yes |
| Admins can add a user (username, role, starting password) from Administration → Users | `code`; `model` (standard for staff systems) | yes |
| A way to create the first admin on a fresh install | `code`; `model` | yes (decision 2) |
| Fix the swallowed insert error and the misleading duplicate message | `code` | yes |
| Audit entry when a user is created | `code` (review checklist: user changes are audited) | yes |
| Common-password blocklist | `web` NIST SP 800-63B; list from SecLists NCSC top 100k | yes (owner's addition) |
| Rest of NIST: 15 minimum while login is single-factor, 64 maximum, no composition rules | `web` NIST SP 800-63B | follow-up (decision 3) |
| Deactivate users, users change their own password | `code` | already T-064 |
| Invitations by email or WhatsApp link | `model` | follow-up, not needed for one shop |

## Decisions for you

1. **What replaces the public endpoint?**
   - **A (recommended):** remove it, and add **Add user** to Administration → Users: username, role and starting password, available only to admins. It's what the shop needs day to day, and it reuses the admin permission that already guards that screen.
   - B: remove it only, and create users with a command or SQL. Smallest change, but you couldn't add staff from the app.
   - C: move the same endpoint under `/admin` with no screen. Secure, but no better for you than B.
2. **How does a fresh install get its first admin?** **Decided: a `create-admin` subcommand of the bundled executable** (`<executable> create-admin`, or `/main create-admin` inside the Docker image). It asks for a username and password and creates an Admin; nothing is exposed over the network. The shop runs a built executable, not source, so the command must live in the binary. It moves into the new-store setup later; noted on T-032.
3. **Password minimum.** **Decided: 10 characters, and not a common password** (owner's addition, after commit 1), for new users and admin resets. Existing passwords keep working; login doesn't re-check length. 15 isn't workable for the shop yet; the follow-up card tracks the rest of the NIST guidance.
4. **How do you add staff today?** **Answered: through `/user/register`.** That stops working with this change, and Administration → Users → Add user replaces it (called out in the deploy notes).

## Design

- **No migration.** Users keep the same table; the role is set explicitly on insert, no longer by the column default.
- **API:**
  - Remove `POST /user/register`, its controller method `CreateUserAccount` and `UserService.CreateUserAccount`.
  - Add `POST /admin/api/users` taking `{username, password, role_id}`. It's guarded by `admin:access` (as all of `/admin/` is) and CSRF.
  - Responses: `201 {id}` on success; `400` on invalid input; `409 "username already exists"`; `500` for anything else, with the real error logged.
- **Service:** `AdminService.CreateUser(ctx, username, password, roleID, actorID)` runs in one transaction:
  - trim and validate the username (existing rules), validate the password (≥ 10 characters, counted as characters, not bytes), and check the role exists
  - check for a duplicate, keeping real errors apart from duplicates
  - hash the password, insert with the role, and write an audit entry
- **Audit:** entity `user`, action `created`, actor is the admin. Details are the username and role, never the password. A shared helper writes it, so T-035 can reuse it.
- **Passwords:** one shared rule (`model.User.ValidatePassword`: minimum 10 characters, not on the common-password list, and not containing the username) used by create, `create-admin` and the admin reset, which today has its own 8-character check. The reset form's hint and `minlength` change to 10.
- **Common-password list:** the NCSC top 100,000 passwords (SecLists, MIT licence), cut to the 9,126 distinct entries of 10 or more characters (shorter ones already fail the length rule), lowercased, and embedded in the binary (`backend/model/common_passwords.txt`, about 106 KB, with the licence notice). The check is case-insensitive. The rejection says why, as NIST asks: "this password is too common; choose another". The username check applies when creating a user; the admin reset doesn't load the username.
- **CLI:** in `main.go`, `create-admin` runs after config and DB setup and before templates and the server start. It reads the username from a flag or prompt and the password from a prompt, or from `PHARMACY_ADMIN_PASSWORD` for scripted installs. It uses the same service method with the Admin role and no actor, then exits.
- **`create-admin` after the first admin** (owner's addition, after the PR opened): when an Admin already exists, the command refuses with "an admin already exists (NAME); create-admin is for a new install". The message and `-h` don't mention the override. `--recover` / `-r` (hidden from usage) creates another admin anyway, for when no admin can sign in. The audit entry's details record `"recovery": true`, so a break-glass account shows in the audit log. The flag isn't settable over the web API: the request field is `json:"-"`.
- **UI:** **Add user** button and form in `UsersManager.vue`, matching its existing style: username, role (the existing roles list), password and confirm password. It uses `csrfFetch`, shows errors inline, and refreshes the list on success.
- **Permissions:** no new keys. Creating a user needs `admin:access`, same as the rest of the Users screen. T-099 will split it later.

## Steps

1. **Repository:** `CreateUserTx` (insert with role, return id) and a role-exists check. Test: integration test under `PHARMACY_TEST_DSN`.
2. **Service:** `AdminService.CreateUser` with validation, duplicate handling, hashing and the audit entry, using the shared audit helper. Tests first: invalid username or password, unknown role, duplicate, DB error not reported as duplicate, audit written, password never in the audit details.
3. **Routes:** add `POST /admin/api/users` and controller `CreateUser`; remove `/user/register`, the controller method and `UserService.CreateUserAccount`, and update the interface and mocks. Tests: controller test for status codes; register returns 404.
4. **CLI:** `create-admin` in `main.go`, built into the executable. Test: build the binary, run `create-admin` against a scratch database, and log in with the new admin.
5. **UI:** Add user form in `UsersManager.vue`, then `make assets`. Test: manual check in the running app (add a cashier, log in as them, try a duplicate).
6. **Docs and board:** `create-admin` in `README.md` and `AGENTS.md` (first-run setup); note on T-032 that first-admin creation moves into the store setup. No feature docs exist for Administration yet, so nothing to refresh.

## Testing

- **Automated:** `cd backend && go test ./...`, plus `PHARMACY_TEST_DSN=... go test ./...` for the repository and an end-to-end create.
- **Manual:**
  - `POST /user/register` returns 404
  - the Add user flow works, and a non-admin can't reach `/admin/api/users` (403)
  - `create-admin` works on an empty database
- **Not tested:** no browser end-to-end tests (no JS tooling yet, T-076).

## Deploy notes

- No migrations, and existing users and sessions are unaffected.
- **`/user/register` stops working.** Add staff from Administration → Users → Add user from now on.
- New and reset passwords need at least 10 characters; tell staff before they next set one.
- `make assets` before building the image.

## Commits

Logical groups, each approved by the owner:

1. `7805d23` Replace public registration with admin-only user creation
2. `c57695b` Refuse common passwords and ones containing the username (owner's addition)
3. `58cacf0` Add the `create-admin` command
4. `c1eda82` Add the Add user form
5. `c748b22` Fix review findings: the reset checks the username and keeps the password as typed; strength ignores surrounding spaces; "no audit actor" is explicit; 401 without a signed-in admin; the role foreign key replaces the role query
6. Docs, plan and board (with the PR)
7. `create-admin` refuses once an admin exists, unless the hidden `--recover`/`-r` flag is given (recorded in the audit details)

## Follow-up cards

Added as T-104 and T-105. The invite-by-link idea is folded into T-064.

- **Password policy toward NIST guidance:** at least 64 maximum, no composition rules, and 15 minimum when it's workable or once there's a second factor (T-075). P1, Security hardening.
- **Case-insensitive unique usernames** (from the review): "Ada" and "ada" can both exist. P2, Security hardening.

## Sources

- https://github.com/danielmiessler/SecLists/blob/master/Passwords/Common-Credentials/100k-most-used-passwords-NCSC.txt — the common-password list (NCSC top 100k), MIT licence: https://github.com/danielmiessler/SecLists/blob/master/LICENSE

- https://pages.nist.gov/800-63-4/sp800-63b.html — single-factor passwords at least 15 characters, 8 only with multi-factor, allow 64 or more, use a blocklist, no composition rules.
- https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html — the same length guidance, and generic messages on registration to avoid revealing which accounts exist (not needed here once creation is admin-only).
