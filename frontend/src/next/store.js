import { reactive } from "vue";

export const shared = reactive({
  lowStockItems: [],
  expiringItems: [],
});
