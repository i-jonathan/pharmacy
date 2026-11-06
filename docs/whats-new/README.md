# What's new

The sidebar's **What's new** item opens a slideshow of the newest release's features, with a "New"
dot until the person opens it. It is remembered per browser (`localStorage["whatsNewSeen"]`); there
is no popup at sign-in.

| Piece | Where |
|---|---|
| The releases and their slides | `frontend/src/next/lib/whatsNew.js` |
| Open/seen state | `frontend/src/next/composables/useWhatsNew.js` |
| The dialog | `frontend/src/next/components/WhatsNewDialog.vue` |
| Menu item and dot | `frontend/src/next/components/Sidebar.vue` |
| Pictures | `backend/template/static/img/whats-new/<release>-<slug>.png` |
| Demo data for the pictures | `backend/internal/demoseed`, `backend/cmd/demoseed` |
| Retake script | `docs/whats-new/capture.mjs` |

## Announce a release

1. Add an entry **on top** of `releases` in `whatsNew.js`: an `id` (the release tag, for example
   `v0.4.0`), a `title`, and one slide per feature people can see: `title`, a sentence or two of
   `text`, `image`, `alt`, and `permissions` (keys; the slide shows to anyone who holds at least
   one, and to everyone when the list is missing).
2. Take the pictures (below) and put them in `backend/template/static/img/whats-new/`.
3. `make assets`, then rebuild the image. The pictures are embedded in the binary.

Older entries can stay; only the newest one is shown. Changing the newest entry's `id` shows the
"New" dot again to everyone.

## Take the pictures

Pictures are real screens of the app running on made-up data, light mode, 1100×700, one per slide.

1. A throwaway Postgres on a spare port, with the user, password and database name that the
   embedded `backend/config/.env` expects:

   ```sh
   set -a; . backend/config/.env; set +a
   docker run -d --name pharmacy-demo -p 5544:5432 \
     -e POSTGRES_USER -e POSTGRES_PASSWORD -e POSTGRES_DB postgres:latest
   DSN="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@localhost:5544/$POSTGRES_DB?sslmode=disable"
   make migrate-up DB_URL="$DSN"
   (cd backend && go run ./cmd/demoseed -throwaway -dsn "$DSN")
   ```

   `demoseed` needs `-throwaway` (a tunnel to a real database also looks local), and creates three users (`owner` Admin, `amaka` Manager, `tunde` Cashier, all with the
   password it prints) and a believable shop. It refuses any database that is not on this machine
   or that already has users, so it cannot touch real data.

2. The app on the demo database, on its own port so a dev server on 8000 can keep running:

   ```sh
   make assets
   cd backend && DB_HOST=localhost DB_PORT=5544 PHARMACY_PORT=8123 go run .
   ```

   Open it at `http://127.0.0.1:8123`, or in a private window. Cookies are shared between ports on
   the same host, so signing in to the demo on `localhost:8123` would sign you out of (or mix up) a
   dev server on `localhost:8000`.

3. The pictures, from a folder outside the repo (the repo has no JS tooling):

   ```sh
   mkdir /tmp/shots && cd /tmp/shots && npm init -y && npm i playwright-core
   CHROMIUM=/path/to/chromium node <repo>/docs/whats-new/capture.mjs <repo>/backend/template/static/img/whats-new v0.4.0
   ```

   Any Chromium or Chrome works; Playwright's own (`npx playwright install chromium`) is fine.

4. Look at every picture before committing: no real names, no odd numbers. Then stop the app and
   `docker rm -f pharmacy-demo`.

`capture.mjs` signs in as the right user for each slide and drives the screen into the state shown.
When a screen changes, update the step for its slide there.

## The v0.3.0 pictures

| Slide | User | Screen and state |
|---|---|---|
| `stock-adjustments` | `amaka` | Stock Adjustments → Adjust on the open flag; quantity 9, a note typed |
| `flag-product` | `tunde` | Products → Diclofenac Gel → Flag for check, filled in, not sent |
| `purchase-orders` | `amaka` | Purchase Orders → Needs approval |
| `reprint-receipt` | `owner` | Sales History → the oldest sale (it has a return) → Reprint receipt |
| `add-user` | `owner` | Administration → Users → Add user, username and role filled in |
