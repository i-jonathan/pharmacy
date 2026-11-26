# Deployment

How to run the server for a shop, beyond the local setup in the [README](../README.md#quick-start).
Each item is marked **Required** (the app breaks without it) or **Advised**.

The server listens on plain HTTP on port 8000 (`0.0.0.0:8000`; see `compose.yml`). It has no TLS of its
own, so HTTPS comes from a reverse proxy in front of it. Its settings come from `backend/config/.env`,
which is embedded in the binary at build time, so a settings change means a rebuild.

## Before building

- **Required:** `make assets` before `go build` or `docker compose build server`. The frontend is
  embedded in the binary, and `backend/template/static/dist/` is not tracked in git, so a build without
  it has no screens.
- **Required:** run the migrations (`make migrate-up DB_URL=...`) before starting a new version.

## Behind a reverse proxy (nginx)

```nginx
server {
    listen 443 ssl;
    server_name shop.example.com;
    # ssl_certificate / ssl_certificate_key ...

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;                       # Required
        proxy_set_header X-Forwarded-Proto $scheme;        # Required with HTTPS
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # Live stock-taking updates (WebSocket)
    location /ws {
        proxy_pass http://127.0.0.1:8000;
        proxy_http_version 1.1;                            # Required before nginx 1.29.7
        proxy_set_header Upgrade $http_upgrade;            # Required
        proxy_set_header Connection "upgrade";             # Required
        proxy_set_header Host $host;                       # Required
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 1h;                             # Advised
    }
}
```

What each line is for:

- **`Host $host` (Required).** By default nginx sends the `proxy_pass` address as `Host`
  (`127.0.0.1:8000` here), not the address the browser used. The app compares the browser's `Origin`
  with `Host` in two places:
  - every save (CSRF check, `gorilla/csrf`); without it, every save is refused with 403
    `csrf_failed`.
  - the live stock-taking socket (`/ws`, gorilla/websocket's same-origin check); without it, the
    handshake gets 403 and the count screen says "Live updates aren't connecting".
- **`X-Forwarded-Proto $scheme` (Required when the proxy serves HTTPS).** The app sees plain HTTP from
  the proxy. It reads this header (`backend/adapter/http/middleware/csrf.go`) to learn the browser used
  HTTPS. Without it the CSRF check compares the browser's `https://` origin with `http://` and refuses
  every save.
- **`Upgrade` / `Connection` (Required on `/ws`).** WebSocket needs the HTTP/1.1 upgrade, which nginx
  doesn't pass on by itself. `proxy_http_version 1.1` is nginx's default only from 1.29.7.
- **`proxy_read_timeout` (Advised on `/ws`).** nginx closes a proxied WebSocket that has sent nothing
  for 60 s by default, and the server sends no keep-alive pings. A quiet count would then reconnect
  every minute. Nothing is lost (the page re-reads the count after each reconnect), but it's needless
  churn. Make it longer than a quiet spell is likely to last.
- **`X-Forwarded-For` (Advised).** The app doesn't read it today. It is there for nginx-style logs and
  for future rate limiting (T-024).

Other proxies (Caddy, Traefik, a cloud load balancer) need the same things: keep the browser's
`Host`, send `X-Forwarded-Proto`, allow WebSocket upgrades on `/ws`, and don't cut idle sockets after
a minute.

## Keep the app port private

- **Advised:** behind a proxy, publish the app port only to the machine itself, so nobody reaches
  the app around the proxy (and its HTTPS). In `compose.yml`, `ports: - 127.0.0.1:8000:8000` instead of
  `8000:8000`. Likewise the database's `5432:5432` is only needed for running migrations and tools from
  the host; `127.0.0.1:5432:5432` keeps it off the network.

## Known gaps

- Session and CSRF cookies are not marked `Secure`, so the browser would also send them over plain
  HTTP. Tracked in T-024 (login rate limiting and cookie hardening). Until then, redirect HTTP to HTTPS
  at the proxy and don't serve the app on plain HTTP from the same name.
- No health-check endpoint or graceful shutdown yet (T-027), and no automated backups (T-026).

## Sources

- [nginx: WebSocket proxying](https://nginx.org/en/docs/http/websocket.html): `Upgrade`/`Connection`
  headers, and the 60-second idle close and `proxy_read_timeout`.
- [nginx: proxy module](https://nginx.org/en/docs/http/ngx_http_proxy_module.html): default
  `proxy_set_header Host $proxy_host`, `proxy_read_timeout` 60s, `proxy_http_version` 1.1 from 1.29.7.
