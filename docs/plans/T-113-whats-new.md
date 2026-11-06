# T-113 What's new: release slideshow

Card: BOARD.md T-113 · Branch: feat/T-113-whats-new

## Goal

After an update, staff can see what changed without being told: a "What's new" item in the sidebar
opens a slideshow of real screenshots of the new features, and a small "New" dot points them to it
until they open it. No popup in front of the till.

## What I found

- `code` The sidebar is `components/Sidebar.vue`; its bottom block (Administration, Collapse, Dark
  mode) is where a "What's new" button fits. Items are gated with `usePermissions().hasPermission`
  (`Sidebar.vue:131`), and the collapsed sidebar shows icons only.
- `code` One shared dialog pattern exists, `ConfirmDialog.vue`, mounted once in `App.vue` (backdrop
  `z-[70]`, `role` dialog, Esc, focus on the panel). The slideshow follows its look but is its own
  component: it is not a yes/no question, so the confirm composable does not fit.
- `code` Static files under `backend/template/static/**` are embedded (`main.go:27`) and served at
  `/static/`. `static/dist/` is wiped by every Vite build, so the pictures go in a sibling folder:
  `static/img/whats-new/`.
- `code` The app has no version number at runtime. The slides carry their own release id
  (`v0.3.0`), which is what "seen" is compared with.
- `code` The server config is embedded (`config/.env`, gitignored); runtime env only reaches
  `DB_HOST` and `DB_PORT` (`config/config.go`), so a demo database must use the same user, name and
  password on another port. A throwaway Postgres container on port 5544 does that without touching
  the real database.
- `code` `create-admin` prompts for the password on a terminal, so the demo users are created by a
  seed program through the user service instead.
- `code` The repo has no JS tests or lint (CLAUDE.md), so the slide-filtering logic is checked by
  hand in Node and in the running app.
- `model` (unverified) Screenshots at 1280×800 as PNG are about 150–300 KB each, so five are around
  1–1.5 MB, embedded in the binary.

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| Sidebar "What's new" item + "New" dot (remembered in the browser) | owner | yes |
| Slideshow dialog: image, title, two lines, Next/Back, dots, Esc, arrow keys | owner | yes |
| Slides filtered by the person's permissions | owner | yes |
| Real screenshots from a demo database | owner | yes |
| Seed program for the demo database (re-runnable) | owner | yes |
| How to add the next release's slides, in CLAUDE.md | `code` | yes |
| Popup after sign-in, remembered per person on the server | owner | follow-up |
| Dark-mode screenshots | `model` | follow-up |
| Tour overlays that point at live screens | `model` | no |

## Decisions for you

1. **How the pictures are taken.** Decided: by hand, against the demo database, and the seed program plus a shot list (screen, user, window size) in the
   docs let anyone retake them. The alternative is a Playwright script that retakes them with one
   command, but it adds a dev dependency and a browser download to a repo that has no JS tooling.
2. **Light mode only** for the pictures. Recommended: yes; dark variants double the files.
3. **Slides in the first release (v0.3.0).** Recommended five, each shown only to people who can
   use it:
   1. Stock Adjustments (`stock:adjust`)
   2. Flagging a product for checking (`stock:flag`)
   3. Purchase orders: draft, submit, approve (`purchaseorder:draft` or `purchaseorder:approve`)
   4. Reprint receipt with the preview (everyone)
   5. Add user in Administration (`admin:access`)

   Anything the release notes mention that is not visual (audit log) stays out.

## Design

**No backend, no migration, no new permission.**

**Frontend**
- `lib/whatsNew.js`: the releases, newest first:
  `{ id: "v0.3.0", title, slides: [{ id, title, text, image, permissions?: [keys] }] }`.
  `visibleSlides(release, hasPermission)` returns the slides the person may see (a slide with
  `permissions` shows if they hold any of them; no list means everyone).
- `composables/useWhatsNew.js`: `open` state, `openWhatsNew()`, `unseen` (the newest release id
  against `localStorage["whatsNewSeen"]`, read and written in try/catch so a blocked store only
  means the dot stays). Opening marks the newest release seen.
- `components/WhatsNewDialog.vue`: mounted once in `App.vue`. `role="dialog"`, focus on open,
  Esc closes, left/right arrows and buttons move, dots jump, image preloads the next slide, alt
  text on every image, `prefers-reduced-motion` respected, works at phone width.
- `Sidebar.vue`: "What's new" button (sparkles icon) in the bottom block, with a dot while unseen,
  also when collapsed.
- Pictures: `backend/template/static/img/whats-new/v0.3.0-*.png`, referenced as `/static/img/...`.

**Demo data and running a second server**
- `PHARMACY_PORT` (default 8000) now sets the listen port (`config.Config.Port`, `main.go`; a failed listen now stops the server with the error instead of exiting silently), so a demo server
  can run next to a dev server. Nothing else about startup changes.
- `backend/cmd/demoseed/main.go` (a separate `main`, not part of the server binary): migrates and
  fills a database with made-up users (owner, manager, cashier), about 30 products, a day of sales
  including a partly returned one, a pending purchase order, an open stock flag and one adjustment.
  Refuses to run unless the database host is this machine (and no connection setting other than `sslmode` is given, since pgx lets `host=` override the address) and the database has no users yet, so it
  can never be pointed at real data by mistake. It also needs an explicit `-throwaway` flag, since a
  tunnel to a real database looks local. (A "name ends in demo" rule would not work: the
  embedded config fixes the database name.)
- The demo database runs as a throwaway container on port 5544 with the credentials the embedded
  config expects; stopped and removed afterwards.

**As built:** this session had no browser tool, so the pictures were taken with `playwright-core`
(installed outside the repo) driving an already-cached Chromium. The script is kept in
`docs/whats-new/capture.mjs` for retakes; it is not a repo dependency.

## Steps

1. Card and plan (this file).
2. Seed program; run it against the demo container. Test: the host/name guard refuses a
   non-demo database.
3. Start the app on the demo database; check each screen shows believable data.
4. Capture the five pictures (light mode, 1280×800), crop and check them for anything real.
5. `whatsNew.js`, composable, dialog, sidebar item. Test: slide filtering in Node (cashier,
   manager, admin) and the dialog in the running app at desktop and phone width.
6. CLAUDE.md note: how to add a release's slides and retake the pictures.
7. `make assets`; commit `tailwind.css` if it changed.

## Testing

- `go test ./...` (with `PHARMACY_TEST_DSN`) for the seed guard.
- Slide filtering run in Node for three roles.
- In the running app on the demo database: open from the sidebar, dot appears for a new browser
  and disappears after opening, Esc and arrows work, a cashier sees fewer slides than an admin,
  phone width, collapsed sidebar.
- Not tested: dark mode (the dialog uses the app's tokens, the pictures are light), and any
  browser but the one I drive.

## Deploy notes

No migrations or permissions. `make assets` before building; the pictures are tracked files, so
they ship in the image.

## Commits

Logical groups, proposed once after implementation and review (seed program; pictures; dialog,
sidebar and docs; plan and board).

## Follow-up cards

- Show What's new after sign-in, remembered per person on the server (P2, Notifications)
- Dark-mode versions of the pictures (P2, Notifications)

## Sources

None; this is internal UI. Existing patterns: `ConfirmDialog.vue`, `Sidebar.vue`.
