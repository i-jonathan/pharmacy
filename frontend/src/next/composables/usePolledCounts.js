import { reactive } from "vue";
import { usePermissions } from "./usePermissions.js";

// polledCounts makes a composable for counts polled from `url` (every minute) for users with
// `permission`, such as the sidebar badges. One timer serves every component that asks for them.
export function polledCounts(url, permission, fields) {
  const counts = reactive(Object.fromEntries(fields.map((f) => [f, 0])));
  let timer = null;
  let users = 0;

  async function refresh() {
    try {
      const r = await fetch(url);
      if (!r.ok) return;
      const data = await r.json();
      for (const f of fields) counts[f] = data[f] || 0;
    } catch {
      // The badge is a convenience; keep the last numbers when the request fails.
    }
  }

  return function use() {
    const { hasPermission } = usePermissions();
    const allowed = () => hasPermission(permission);

    // Only a consumer that started polling may release it; otherwise a component that never
    // started (no permission) would lower the count for the ones that did.
    let started = false;

    function start() {
      if (!allowed() || started) return;
      started = true;
      users += 1;
      if (!timer) {
        refresh();
        timer = setInterval(refresh, 60000);
      }
    }

    function stop() {
      if (!started) return;
      started = false;
      users -= 1;
      if (users === 0 && timer) {
        clearInterval(timer);
        timer = null;
      }
    }

    // Call after anything that changes the counts, so badges don't linger until the next poll.
    function refreshNow() {
      if (allowed()) refresh();
    }

    return { counts, refresh: refreshNow, start, stop };
  };
}
