<template>
  <div class="min-h-screen bg-background">
    <Sidebar
      :collapsed="sidebarCollapsed"
      :is-dark="isDark"
      :mobile-open="mobileNavOpen"
      @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
      @toggle-theme="toggleTheme"
      @open-admin="openAdminPanel"
      @close-mobile="mobileNavOpen = false"
    />

    <!-- Main Content -->
    <div
      class="min-h-screen min-w-0 transition-[margin] duration-200"
      :class="sidebarCollapsed ? 'lg:ml-16' : 'lg:ml-60'"
    >
      <TopNav @toggle-navigation="mobileNavOpen = !mobileNavOpen" />
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
const mobileNavOpen = ref(false);
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
  try {
    localStorage.setItem("theme", isDark.value ? "dark" : "light");
  } catch {
    // Theme preference is optional when storage is unavailable.
  }
}

function openAdminPanel(payload = {}) {
  adminModule.value = payload?.module ?? null;
  adminOpen.value = true;
}

onMounted(() => {
  let savedTheme = null;
  try {
    savedTheme = localStorage.getItem("theme");
  } catch {
    // Fall back to the default light theme.
  }
  isDark.value = savedTheme === "dark";
  document.documentElement.classList.toggle("dark", isDark.value);
});
</script>
