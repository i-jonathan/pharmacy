import { reactive } from "vue";
import { usePermissions } from "./usePermissions.js";

// Managers get a badge and a dashboard banner when purchase orders wait for approval. One timer
// serves every component that asks for the counts.
const counts = reactive({ pending: 0, drafts: 0 });
let timer = null;
let users = 0;

async function refresh() {
  try {
    const r = await fetch("/inventory/purchase-orders/pending-count");
    if (!r.ok) return;
    const data = await r.json();
    counts.pending = data.pending || 0;
    counts.drafts = data.drafts || 0;
  } catch {
    // The badge is a convenience; keep the last numbers when the request fails.
  }
}

export function usePendingPurchaseOrders() {
  const { hasPermission } = usePermissions();
  const canApprove = () => hasPermission("purchaseorder:approve");

  function start() {
    if (!canApprove()) return;
    users += 1;
    if (!timer) {
      refresh();
      timer = setInterval(refresh, 60000);
    }
  }

  function stop() {
    users = Math.max(0, users - 1);
    if (users === 0 && timer) {
      clearInterval(timer);
      timer = null;
    }
  }

  // Call after anything that changes what waits for approval, so the badge and banner do not
  // linger until the next poll.
  function refreshNow() {
    if (canApprove()) refresh();
  }

  return { counts, refresh: refreshNow, start, stop };
}
