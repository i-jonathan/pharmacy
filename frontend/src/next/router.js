import { createRouter, createWebHashHistory } from "vue-router";

const routes = [
  {
    path: "/",
    name: "dashboard",
    component: () => import("./components/DashboardView.vue"),
  },
  {
    path: "/low-stock",
    name: "low-stock",
    component: () => import("./components/LowStockPage.vue"),
  },
  {
    path: "/expiring",
    name: "expiring",
    component: () => import("./components/ExpiringPage.vue"),
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
