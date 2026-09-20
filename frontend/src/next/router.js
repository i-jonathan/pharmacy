import { createRouter, createWebHistory } from "vue-router";

const routes = [
  // Redirect old /dashboard path to root (v2)
  {
    path: "/dashboard",
    redirect: "/",
  },
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
  {
    path: "/pos",
    name: "pos",
    component: () => import("./components/PosView.vue"),
    meta: { title: "Point of Sale", parent: "Dashboard" },
  },
  {
    path: "/sales-history",
    name: "sales-history",
    component: () => import("./views/SalesHistory.vue"),
    meta: { title: "Sales History", parent: "Sales" },
  },
  {
    path: "/held-sales",
    name: "held-sales",
    component: () => import("./views/HeldSales.vue"),
    meta: { title: "Held Sales", parent: "Sales" },
  },
  {
    path: "/products",
    name: "products",
    component: () => import("./views/ProductList.vue"),
    meta: { title: "Products", parent: "Inventory" },
  },
  {
    path: "/receive-items",
    name: "receive-items",
    component: () => import("./views/ReceiveItems.vue"),
    meta: { title: "Receive Items", parent: "Inventory" },
  },
  {
    path: "/held-receive-items",
    name: "held-receive-items",
    component: () => import("./views/HeldReceiveItems.vue"),
    meta: { title: "Held Receive Items", parent: "Inventory" },
  },
  {
    path: "/received-items-history",
    name: "received-items-history",
    component: () => import("./views/ReceivedItemsHistory.vue"),
    meta: { title: "Received Items History", parent: "Inventory" },
  },
  {
    path: "/stock-taking",
    name: "stock-taking",
    component: () => import("./views/StockTakingDashboard.vue"),
    meta: { title: "Stock Taking", parent: "Inventory" },
  },
  {
    path: "/categories",
    name: "categories",
    component: () => import("./views/CategoriesPage.vue"),
    meta: { title: "Categories", parent: "Inventory" },
  },
];

const router = createRouter({
  history: createWebHistory("/app/"),
  routes,
});

export default router;