// readError reads a refused JSON reply: the server's message (or fallback) and its code, if any
// (for example "already_saved", "csrf_failed").
export async function readError(response, fallback) {
  const body = await response.json().catch(() => ({}));
  return { message: body.error || fallback, code: body.code || "" };
}

// newRetryKey makes a key that marks repeats of one save (randomUUID needs a secure context, so
// plain-HTTP LAN access falls back to time plus randomness).
export function newRetryKey() {
  return globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}
