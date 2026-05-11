<template>
  <div class="min-h-screen bg-background">
    <Sidebar
      :collapsed="sidebarCollapsed"
      :is-dark="isDark"
      @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
      @toggle-theme="toggleTheme"
      @open-admin="openAdminPanel"
    />

    <!-- Main Content -->
    <div
      class="transition-all duration-300"
      :class="sidebarCollapsed ? 'ml-16' : 'ml-60'"
    >
      <TopNav />
      <RouterView />
    </div>

    <!-- Admin Panel -->
    <PermissionGate permission="admin:access">
      <AdminPanel
        :open="adminOpen"
        :initial-module="adminModule"
        @close="adminOpen = false"
      />
    </PermissionGate>
  </div>
</template>

<script setup>
import { ref, provide, onMounted } from "vue";
import { PermissionsKey, UserKey } from "./composables/usePermissions.js";
import Sidebar from "./components/Sidebar.vue";
import TopNav from "./components/TopNav.vue";
import PermissionGate from "./components/PermissionGate.vue";
import AdminPanel from "./components/AdminPanel.vue";

const sidebarCollapsed = ref(false);
const isDark = ref(false);
const adminOpen = ref(false);
const adminModule = ref(null);

const permissions = ref(window.__PERMISSIONS__ ?? {});
const user = ref(window.__USER__ ?? { id: 0 });

provide(PermissionsKey, permissions);
provide(UserKey, user);

function toggleTheme() {
  isDark.value = !isDark.value;
  document.documentElement.classList.toggle("dark", isDark.value);
  localStorage.setItem("theme", isDark.value ? "dark" : "light");
}

function openAdminPanel(payload = {}) {
  adminModule.value = payload?.module ?? null;
  adminOpen.value = true;
}

onMounted(() => {
  isDark.value = localStorage.getItem("theme") === "dark";
  document.documentElement.classList.toggle("dark", isDark.value);
});
</script>
