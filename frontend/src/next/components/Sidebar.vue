<template>
  <aside
    class="fixed inset-y-0 left-0 z-40 flex flex-col border-r border-border bg-background transition-all duration-300"
    :class="collapsed ? 'w-16' : 'w-60'"
  >
    <!-- Logo -->
    <div class="flex items-center h-14 px-4 border-b border-border">
      <Pill :stroke-width="1.5" :size="22" class="text-foreground shrink-0" />
      <span v-if="!collapsed" class="ml-3 font-bold text-lg tracking-tight truncate">
        Primocrest
      </span>
    </div>

    <!-- Nav -->
    <nav class="flex-1 px-2 py-3 space-y-4 overflow-y-auto">
      <!-- Dashboard -->
      <a
        href="/app/dashboard?ui=v2"
        :class="linkClasses(true)"
        :title="collapsed ? 'Dashboard' : ''"
      >
        <LayoutDashboard :stroke-width="1.5" :size="18" class="shrink-0" />
        <span v-if="!collapsed">Dashboard</span>
      </a>

      <!-- Sales Section -->
      <div>
        <button
          v-if="!collapsed"
          class="flex items-center justify-between w-full px-3 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
          @click="toggleSection('sales')"
        >
          <span>Sales</span>
          <ChevronDown :stroke-width="1.5" :size="14" :class="sectionOpen.sales ? 'rotate-0' : '-rotate-90'" class="transition-transform" />
        </button>
        <div v-if="!collapsed" class="w-full h-px bg-border mb-1" />
        <div v-show="collapsed || sectionOpen.sales" class="space-y-0.5">
          <a href="/sales/receipt" :class="linkClasses(false)"><ShoppingCart :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed">Point of Sale</span></a>
          <a href="/sales/history" :class="linkClasses(false)"><History :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed">Sales History</span></a>
          <a href="/sales/held" :class="linkClasses(false)"><PauseCircle :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed">Held Sales</span></a>
        </div>
      </div>

      <!-- Inventory Section -->
      <div>
        <button
          v-if="!collapsed"
          class="flex items-center justify-between w-full px-3 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
          @click="toggleSection('inventory')"
        >
          <span>Inventory</span>
          <ChevronDown :stroke-width="1.5" :size="14" :class="sectionOpen.inventory ? 'rotate-0' : '-rotate-90'" class="transition-transform" />
        </button>
        <div v-if="!collapsed" class="w-full h-px bg-border mb-1" />
        <div v-show="collapsed || sectionOpen.inventory" class="space-y-0.5">
          <a href="/inventory/items" :class="linkClasses(false)"><Package :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed">Products</span></a>
          <a href="/inventory/receive-items" :class="linkClasses(false)"><Truck :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed">Receive Items</span></a>
          <a href="/stock-taking/" :class="linkClasses(false)"><ClipboardCheck :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed">Stock Taking</span></a>
          <PermissionGate permission="admin:access">
            <button class="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors" @click="$emit('open-admin', { module: 'categories' })">
              <Tags :stroke-width="1.5" :size="18" class="shrink-0" /><span v-if="!collapsed">Categories</span>
            </button>
          </PermissionGate>
        </div>
      </div>
    </nav>

    <!-- Bottom Actions -->
    <div class="border-t border-border px-2 py-3 space-y-1">
      <PermissionGate permission="admin:access">
        <button
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors w-full"
          @click="$emit('open-admin', {})"
        >
          <Shield :stroke-width="1.5" :size="18" class="shrink-0" />
          <span v-if="!collapsed">Administration</span>
        </button>
      </PermissionGate>

      <button
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors w-full"
        @click="$emit('toggle-collapse')"
      >
        <component :is="collapsed ? PanelRightOpen : PanelLeftClose" :stroke-width="1.5" :size="18" class="shrink-0" />
        <span v-if="!collapsed">Collapse</span>
      </button>

      <button
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors w-full"
        @click="$emit('toggle-theme')"
      >
        <component :is="isDark ? Sun : Moon" :stroke-width="1.5" :size="18" class="shrink-0" />
        <span v-if="!collapsed">{{ isDark ? 'Light mode' : 'Dark mode' }}</span>
      </button>

      <a
        href="/app/dashboard"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        :title="collapsed ? 'Switch to old UI' : ''"
      >
        <ArrowLeftRight :stroke-width="1.5" :size="18" class="shrink-0" />
        <span v-if="!collapsed">Old UI</span>
      </a>
    </div>
  </aside>
</template>

<script setup>
import { reactive } from "vue";
import {
  LayoutDashboard, ShoppingCart, History, PauseCircle,
  Package, Truck, ClipboardCheck, Tags, Shield,
  Pill, PanelLeftClose, PanelRightOpen, Moon, Sun, ArrowLeftRight,
  ChevronDown,
} from "lucide-vue-next";
import PermissionGate from "./PermissionGate.vue";

defineProps({
  collapsed: { type: Boolean, default: false },
  isDark: { type: Boolean, default: false },
});

defineEmits(["toggle-collapse", "toggle-theme", "open-admin"]);

const sectionOpen = reactive({
  sales: true,
  inventory: true,
});

function toggleSection(key) {
  sectionOpen[key] = !sectionOpen[key];
}

function linkClasses(active) {
  return [
    "flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors",
    active
      ? "bg-foreground/5 text-foreground"
      : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
  ];
}
</script>
