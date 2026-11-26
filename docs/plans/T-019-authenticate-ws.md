# T-019 Authenticate the /ws WebSocket

Card: T-019 · Branch: fix/T-019-authenticate-ws

## Goal
Right now anyone who can reach the server can open `/ws?stockTakingId=N` and watch a stock count live:
product ids, counts and when the count finishes. After this change only a signed-in user, on a page
served by this server, can open the connection, and it closes when their session ends.

## What I found
- Verified on the card: `/ws` skips the middleware stack. `main.go:132-145` sends `/ws` to its own
  mux before `mainHandler`. `code`
- Verified on the card: `HandleWebSocket` never reads the session. It parses `stockTakingId` and
  upgrades (`adapter/websocket/handler.go:46-83`). `code`
- Why `/ws` skips the stack: `Logging`'s `wrappedWriter` (`middleware/logging.go:9`) has no `Hijack`,
  and the upgrade needs one. `AuthMiddleware` passes the original `ResponseWriter` through
  (`middleware/auth.go:18-44`), so it can wrap `/ws` without breaking the upgrade. `code`
- The upgrade request is a GET with no `text/html` Accept, so `AuthMiddleware` answers a signed-out
  call with 401 JSON before the upgrade and does not redirect it or save it as `next`
  (`middleware/auth.go:14-16, 29-35`). `code`
- **Origin check is weaker than it looks** (`handler.go:15-41`). It allows any origin whenever the
  server's Host contains `localhost`/`127.0.0.1`, which is the normal setup for a shop running on
  its own machine. It also uses substring matching. Once cookies authenticate `/ws`, that opens
  Cross-Site WebSocket Hijacking: any website the cashier visits could open the socket using their
  cookie. `code` + `web` (OWASP)
- gorilla/websocket's default `CheckOrigin` (used when the field is nil) allows a missing Origin and
  otherwise requires the Origin's host to equal `r.Host`, compared case-insensitively
  (`gorilla/websocket server.go:89-99`). That is the allowlist we need. `code`
- The SPA connects to `${location.host}/ws` (`StockTakingDashboard.vue:558-588`) and retries every
  3 s after any close. Vite's dev proxy has no `/ws` entry (`frontend/vite.config.js`), so live
  updates already fail under `npm run dev`. That is why the localhost exceptions exist, and they
  never actually helped. `code`
- Who may watch: every stock-taking read route needs only a signed-in user. There is no "view"
  permission (`router/stock_taking.go:17-27`, `main.go:111`), so `/ws` needs the same. `code`
- The server only sends messages. `ReadPump` throws away whatever the client sends
  (`hub.go:131-150`), so there are no per-message actions to authorise. `code`
- The session is a cookie that expires at the next 03:00 (`config/session.go:10-25`). Logout clears
  the cookie (`controller/user.go:99`). Logout leaves the page, which closes the socket anyway, but a
  socket opened before 03:00 would stay open after the session expires. `code`
- Wrong or stale on the card: "and tenant". There are no tenants yet (T-029), so this card has
  nothing to check there. T-029 will need to scope hub broadcasts by tenant. I'll add that note to
  T-029 and correct this card.

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| Refuse the upgrade without a session (401 before upgrade) | `code`, `web` | yes |
| Strict same-origin check (drop the localhost/substring rules) | `code`, `web` | yes |
| Close the socket when the session expires (03:00 cutoff) | `web` | yes |
| `stockTakingId` must be a positive number | `code` | yes (one line) |
| Vite dev proxy for `/ws` so live updates work in dev under the strict check | `code` | yes (dev only) |
| Re-check that the user is still active or allowed, on each request and on open sockets | `web` | follow-up: already T-023 |
| Close a user's sockets on logout from another tab | `web` | follow-up: low value while logout leaves the page; revisit with T-023 or T-024 |
| Client stops retrying after a refusal (today it retries every 3 s, forever) | `code`, code review | yes (moved in by owner after review) |

## Decisions for you
None blocking. One judgement call, recommended as written: **allow a missing Origin.** Browsers
always send Origin on a WebSocket handshake, so a request without one is not a browser, and
cross-site hijacking needs a browser. Such a client still needs a valid session cookie. This is also
gorilla's default.

## Design
- `main.go`: `wsRouter.Handle("/ws", middleware.AuthMiddleware(http.HandlerFunc(wsHub.HandleWebSocket)))`.
  `/ws` still skips Logging/CSRF (a GET; CSRF does not apply), so the CLAUDE.md gotcha changes to:
  "bypasses logging and CSRF middleware, but requires a session".
- `handler.go`: remove the custom `CheckOrigin` so gorilla's same-origin default applies, and refuse
  `stockTakingId <= 0`.
- Session expiry: move the 03:00 calculation in `config/session.go` into
  `config.SessionExpiry(now time.Time) time.Time`, which `NewSessionStore` keeps using. After the
  upgrade the handler arms `time.AfterFunc(until expiry, conn.Close)`, stopped when the read pump
  ends. The client's reconnect then gets 401.
- `vite.config.js`: add `/ws` with `ws: true` and rewrite `Origin` to the target in `proxyReqWs`.
  This is dev only.
- No migration, permission, audit entry or UI change.

## Steps
1. Test (`adapter/websocket/handler_test.go`): through `AuthMiddleware`, a handshake without a
   session gets 401 and no upgrade. With a session and a same-origin Origin it upgrades and receives
   a broadcast. A foreign Origin (also with Host `localhost:8000`) gets 403. `stockTakingId=0` gets
   400. Then implement the `main.go` wiring and `handler.go` changes.
2. Test (`config/session_test.go`): `SessionExpiry` before and after 03:00. Test (handler): a
   connection is closed at the deadline (inject a short expiry through a package-level func var).
   Then implement.
3. Vite `/ws` proxy, then update the CLAUDE.md `/ws` gotcha and add the reverse-proxy `Host` note to
   `AGENTS.md`.
4. `StockTakingDashboard.vue`: after the socket closes, back off between retries (3 s doubling to
   60 s, reset on open). When a reconnect fails, ask the stock-taking API whether the session is still
   valid. On 401, stop retrying and show "Live updates stopped, sign in again". No JS test tooling
   exists, so this is checked by hand.

## Testing
- `go test ./adapter/... ./config/...` using `httptest.Server` and the gorilla client dialer, with no
  database.
- Manual: `make dev`, open a stock taking in two browsers and check that live updates arrive. Signed
  out, `websocat ws://localhost:8000/ws?stockTakingId=1` gets 401. A different Origin gets 403.
- Not tested: real 03:00 rollover (covered by the injected clock).

## Deploy notes
If the server ever sits behind a reverse proxy that rewrites `Host`, the same-origin check will refuse
the socket. The proxy must pass `Host` through (nginx `proxy_set_header Host $host`). No proxy config
exists in the repo today. `make assets` is not needed (only the dev proxy config changes in the
frontend). Nobody has to sign in again.

## Changes after approval
- After the first code review the owner moved the client retry fix (step 4) into this ticket.

## Commits
Style: logical groups (owner's choice). Commit points:
1. Session required + same-origin check + id check on `/ws` (step 1), with the plan file.
2. Close the socket at session expiry (step 2).
3. Vite `/ws` dev proxy, CLAUDE.md and AGENTS.md notes (step 3), and the page's retry back-off
   (step 4).

## Follow-up cards
- Check whether stock-taking REST and live updates send `snapshot_quantity` to users the screen
  hides it from; if they do, strip it on the server (P1, Inventory control).
- Note on T-024: `config/session.go` sets only `Options.MaxAge`, so securecookie accepts a session
  cookie for 30 days; call `s.MaxAge(...)` so the server enforces the 03:00 cutoff too.
- Note on T-029: hub broadcasts must be scoped by tenant.

## Sources
- https://cheatsheetseries.owasp.org/cheatsheets/WebSocket_Security_Cheat_Sheet.html: browsers send
  cookies on the handshake (CSWSH); validate Origin with an explicit allowlist and no substring
  matching; don't "just return true" in Gorilla's CheckOrigin; close connections when sessions expire.
