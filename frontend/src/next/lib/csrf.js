export function csrfFetch(input, init = {}) {
  const method = (init.method || "GET").toUpperCase();
  const headers = new Headers(init.headers || {});
  const isUnsafeMethod = !["GET", "HEAD", "OPTIONS", "TRACE"].includes(method);

  if (isUnsafeMethod) {
    const token = window.__CSRF_TOKEN__ || "";
    if (token) headers.set("X-CSRF-Token", token);
    if (!headers.has("Accept")) headers.set("Accept", "application/json");
  }

  const request = () => fetch(input, {
    ...init,
    headers,
    credentials: init.credentials || "same-origin",
  });

  const refreshToken = async () => {
    const tokenResponse = await fetch("/csrf-token", {
      headers: { Accept: "application/json" },
      credentials: "same-origin",
      cache: "no-store",
    });
    if (!tokenResponse.ok) return "";

    const refreshed = await tokenResponse.json().catch(() => null);
    if (!refreshed?.token) return "";

    headers.set("X-CSRF-Token", refreshed.token);
    window.__CSRF_TOKEN__ = refreshed.token;
    return refreshed.token;
  };

  return (async () => {
    if (isUnsafeMethod) {
      try {
        await refreshToken();
      } catch {
        // Keep the page token as a fallback if token refresh is unavailable.
      }
    }

    const response = await request();
    if (response.status !== 403 || !["POST", "PUT", "PATCH", "DELETE"].includes(method)) {
      return response;
    }

    const failure = await response.clone().json().catch(() => null);
    if (failure?.code !== "csrf_failed") return response;

    try {
      if (!await refreshToken()) return response;
    } catch {
      return response;
    }
    return request();
  })();
}
