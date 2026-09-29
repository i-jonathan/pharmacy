<template>
  <div v-if="mobileOpen" class="fixed inset-0 z-40 bg-neutral-950/40 lg:hidden" aria-hidden="true" @click="$emit('close-mobile')"></div>
  <aside
    class="fixed inset-y-0 left-0 z-50 flex w-[min(86vw,18rem)] flex-col border-r border-border bg-background transition-transform duration-200 lg:z-40 lg:translate-x-0"
    :class="[mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0', collapsed ? 'lg:w-16' : 'lg:w-60']"
  >
    <!-- Logo -->
    <div class="flex items-center h-14 px-4 border-b border-border">
      <PillBottle :stroke-width="1.5" :size="22" class="text-foreground shrink-0" />
      <span v-if="!collapsed || mobileOpen" class="ml-3 truncate text-base font-semibold tracking-tight lg:block">
        Primocrest <span class="font-normal text-muted-foreground">Pharmacy</span>
      </span>
    </div>

    <!-- Nav -->
    <nav class="flex-1 px-2 py-3 space-y-4 overflow-y-auto">
      <!-- Dashboard -->
      <router-link
        to="/"
        :class="linkClasses('/')"
        :title="collapsed ? 'Dashboard' : ''"
        @click="$emit('close-mobile')"
      >
        <LayoutDashboard :stroke-width="1.5" :size="18" class="shrink-0" />
        <span v-if="!collapsed || mobileOpen">Dashboard</span>
      </router-link>

      <!-- Sales Section -->
      <div>
        <button
          v-if="!collapsed || mobileOpen"
          class="flex items-center justify-between w-full px-3 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
          @click="toggleSection('sales')"
        >
          <span>Sales</span>
          <ChevronDown :stroke-width="1.5" :size="14" :class="sectionOpen.sales ? 'rotate-0' : '-rotate-90'" class="transition-transform" />
        </button>
        <div v-if="!collapsed || mobileOpen" class="w-full h-px bg-border mb-1" />
        <div v-show="collapsed || sectionOpen.sales || isInSection(salesPaths)" class="space-y-0.5">
          <router-link to="/pos" :class="linkClasses('/pos')" @click="$emit('close-mobile')"><ShoppingCart :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed || mobileOpen">Point of Sale</span></router-link>
          <router-link to="/sales-history" :class="linkClasses('/sales-history')" @click="$emit('close-mobile')"><History :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed || mobileOpen">Sales History</span></router-link>
          <router-link to="/held-sales" :class="linkClasses('/held-sales')" @click="$emit('close-mobile')"><PauseCircle :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed || mobileOpen">Held Sales</span></router-link>
        </div>
      </div>

      <!-- Inventory Section -->
      <div>
        <button
          v-if="!collapsed || mobileOpen"
          class="flex items-center justify-between w-full px-3 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
          @click="toggleSection('inventory')"
        >
          <span>Inventory</span>
          <ChevronDown :stroke-width="1.5" :size="14" :class="sectionOpen.inventory ? 'rotate-0' : '-rotate-90'" class="transition-transform" />
        </button>
        <div v-if="!collapsed || mobileOpen" class="w-full h-px bg-border mb-1" />
        <div v-show="collapsed || sectionOpen.inventory || isInSection(inventoryPaths)" class="space-y-0.5">
          <router-link to="/products" :class="linkClasses('/products')" @click="$emit('close-mobile')"><Package :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed || mobileOpen">Products</span></router-link>
          <router-link to="/receive-items" :class="linkClasses('/receive-items')" @click="$emit('close-mobile')"><Truck :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed || mobileOpen">Receive Items</span></router-link>
          <router-link to="/received-items-history" :class="linkClasses('/received-items-history')" @click="$emit('close-mobile')"><ClipboardList :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed || mobileOpen">Received History</span></router-link>
          <router-link to="/held-receive-items" :class="linkClasses('/held-receive-items')" @click="$emit('close-mobile')"><PauseCircle :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed || mobileOpen">Held Receipts</span></router-link>
          <router-link to="/stock-taking" :class="linkClasses('/stock-taking')" @click="$emit('close-mobile')"><ClipboardCheck :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed || mobileOpen">Stock Taking</span></router-link>
        </div>
      </div>
    </nav>

    <!-- Bottom Actions -->
    <div class="border-t border-border px-2 py-3 space-y-1">
      <PermissionGate permission="admin:access">
        <button
          class="flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          @click="$emit('close-mobile'); $emit('open-admin', {})"
        >
          <Shield :stroke-width="1.5" :size="18" class="shrink-0" />
          <span v-if="!collapsed || mobileOpen">Administration</span>
        </button>
      </PermissionGate>

      <button
        class="hidden w-full items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring lg:flex"
        @click="$emit('toggle-collapse')"
      >
        <component :is="collapsed ? PanelRightOpen : PanelLeftClose" :stroke-width="1.5" :size="18" class="shrink-0" />
        <span v-if="!collapsed || mobileOpen">Collapse</span>
      </button>

      <button
        class="flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        @click="$emit('toggle-theme')"
      >
        <component :is="isDark ? Sun : Moon" :stroke-width="1.5" :size="18" class="shrink-0" />
        <span v-if="!collapsed || mobileOpen">{{ isDark ? 'Light mode' : 'Dark mode' }}</span>
      </button>

    </div>
  </aside>
</template>

<script setup>
import { ref, watch } from "vue";
import { useRoute } from "vue-router";
import {
  LayoutDashboard, ShoppingCart, History, PauseCircle,
  Package, Truck, ClipboardCheck, ClipboardList, Shield,
  PillBottle, PanelLeftClose, PanelRightOpen, Moon, Sun,
  ChevronDown,
} from "lucide-vue-next";
import PermissionGate from "./PermissionGate.vue";

defineProps({
  collapsed: { type: Boolean, default: false },
  isDark: { type: Boolean, default: false },
  mobileOpen: { type: Boolean, default: false },
});

defineEmits(["toggle-collapse", "toggle-theme", "open-admin", "close-mobile"]);

const route = useRoute();

const salesPaths = ["/pos", "/sales-history", "/held-sales"];
const inventoryPaths = ["/products", "/receive-items", "/received-items-history", "/held-receive-items", "/stock-taking"];

function activeSectionKeys() {
  const keys = [];
  if (salesPaths.includes(route.path)) keys.push("sales");
  if (inventoryPaths.includes(route.path)) keys.push("inventory");
  return keys;
}

const sectionOpen = ref(
  Object.fromEntries(activeSectionKeys().map((k) => [k, true]))
);

watch(
  () => route.path,
  () => {
    for (const key of activeSectionKeys()) {
      sectionOpen.value[key] = true;
    }
  }
);

function isInSection(paths) {
  return paths.includes(route.path) || paths.some(p => route.path.startsWith(p + "/"));
}

function toggleSection(key) {
  sectionOpen.value[key] = !sectionOpen.value[key];
}

function linkClasses(targetPath) {
  const isActive = route.path === targetPath;
  return [
    "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
    isActive
      ? "bg-foreground/5 text-foreground"
      : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
  ];
}
</script>
