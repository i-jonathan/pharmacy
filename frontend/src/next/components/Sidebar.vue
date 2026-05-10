<template>
  <aside
    class="fixed inset-y-0 left-0 z-40 flex flex-col border-r border-border bg-background transition-all duration-300"
    :class="collapsed ? 'w-16' : 'w-60'"
  >
    <!-- Logo -->
    <div class="flex items-center h-14 px-4 border-b">
      <Pill :size="22" class="text-primary shrink-0" />
      <span
        v-if="!collapsed"
        class="ml-3 font-bold text-sm tracking-tight truncate"
      >
        Primocrest
      </span>
    </div>

    <!-- Nav Links -->
    <nav class="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
      <a
        v-for="link in visibleLinks"
        :key="link.label"
        :href="link.href"
        :class="[
          'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors',
          link.active
            ? 'bg-primary/10 text-primary'
            : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
        ]"
        :title="collapsed ? link.label : ''"
        @click="link.onClick ? link.onClick($event) : null"
      >
        <component :is="link.icon" :size="18" class="shrink-0" />
        <span v-if="!collapsed">{{ link.label }}</span>
      </a>
    </nav>

    <!-- Bottom Actions -->
    <div class="border-t px-2 py-3 space-y-1">
      <button
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors w-full"
        @click="$emit('toggle-collapse')"
      >
        <component :is="collapsed ? PanelRightOpen : PanelLeftClose" :size="18" class="shrink-0" />
        <span v-if="!collapsed">Collapse</span>
      </button>

      <button
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors w-full"
        @click="$emit('toggle-theme')"
      >
        <component :is="isDark ? Sun : Moon" :size="18" class="shrink-0" />
        <span v-if="!collapsed">{{ isDark ? 'Light mode' : 'Dark mode' }}</span>
      </button>

      <a
        href="/app/dashboard"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        :title="collapsed ? 'Switch to old UI' : ''"
      >
        <ArrowLeftRight :size="18" class="shrink-0" />
        <span v-if="!collapsed">Old UI</span>
      </a>
    </div>
  </aside>
</template>

<script setup>
import { computed } from "vue";
import {
  LayoutDashboard,
  Package,
  ShoppingCart,
  ClipboardCheck,
  Shield,
  Pill,
  PanelLeftClose,
  PanelRightOpen,
  Moon,
  Sun,
  ArrowLeftRight,
} from "lucide-vue-next";
import { usePermissions } from "../composables/usePermissions.js";

const props = defineProps({
  collapsed: { type: Boolean, default: false },
  isDark: { type: Boolean, default: false },
});

defineEmits(["toggle-collapse", "toggle-theme", "open-admin"]);

const { hasPermission } = usePermissions();

const allLinks = [
  { label: "Dashboard", href: "/app/dashboard?ui=v2", icon: LayoutDashboard, permission: null, active: true },
  { label: "Inventory", href: "/inventory/items", icon: Package, permission: null, active: false },
  { label: "Sales", href: "/sales/receipt", icon: ShoppingCart, permission: null, active: false },
  { label: "Stock Taking", href: "/stock-taking/", icon: ClipboardCheck, permission: null, active: false },
  { label: "Admin", href: "#", icon: Shield, permission: "admin:access", active: false, onClick: (e) => { e.preventDefault(); /* handled by parent */ } },
];

const visibleLinks = computed(() =>
  allLinks.filter((l) => !l.permission || hasPermission(l.permission))
);
</script>
