import { createRouter, createWebHashHistory } from "vue-router";

const routes = [
  {
    path: "/",
    name: "dashboard",
    component: () => import("./components/DashboardView.vue"),
    meta: { title: "Dashboard", subtitle: "Overview of your pharmacy operations" },
  },
  {
    path: "/low-stock",
    name: "low-stock",
    component: () => import("./components/LowStockPage.vue"),
    meta: { title: "Low Stock Items", parent: "Dashboard" },
  },
  {
    path: "/expiring",
    name: "expiring",
    component: () => import("./components/ExpiringPage.vue"),
    meta: { title: "Expiring Items", parent: "Dashboard" },
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
