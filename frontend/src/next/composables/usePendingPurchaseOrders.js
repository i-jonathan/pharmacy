import { polledCounts } from "./usePolledCounts.js";

// Managers get a badge and a dashboard banner when purchase orders wait for approval.
export const usePendingPurchaseOrders = polledCounts(
  "/inventory/purchase-orders/pending-count",
  "purchaseorder:approve",
  ["pending", "drafts"],
);
