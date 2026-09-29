export function csrfFetch(input, init = {}) {
  const method = (init.method || "GET").toUpperCase();
  const headers = new Headers(init.headers || {});

  if (!["GET", "HEAD", "OPTIONS", "TRACE"].includes(method)) {
    const token = window.__CSRF_TOKEN__ || "";
    if (token) headers.set("X-CSRF-Token", token);
  }

  return fetch(input, { ...init, headers, credentials: init.credentials || "same-origin" });
}
