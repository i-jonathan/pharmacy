import { polledCounts } from "./usePolledCounts.js";

// Managers get a sidebar badge for products flagged for checking.
export const useStockFlags = polledCounts("/inventory/stock-flags/count", "stock:adjust", ["open"]);
