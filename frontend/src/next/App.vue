<template>
  <div class="min-h-screen bg-background">
    <Sidebar
      :collapsed="sidebarCollapsed"
      :is-dark="isDark"
      @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
      @toggle-theme="toggleTheme"
    />

    <!-- Main Content -->
    <div
      class="transition-all duration-300"
      :class="sidebarCollapsed ? 'ml-16' : 'ml-60'"
    >
      <DashboardView />
    </div>

    <!-- Admin Panel (rendered outside main flow, shown on sidebar admin click) -->
    <PermissionGate permission="admin:access">
      <AdminPanel :open="adminOpen" @close="adminOpen = false" />
    </PermissionGate>
  </div>
</template>

<script setup>
import { ref, provide, onMounted } from "vue";
import { PermissionsKey, UserKey } from "./composables/usePermissions.js";
import Sidebar from "./components/Sidebar.vue";
import DashboardView from "./components/DashboardView.vue";
import PermissionGate from "./components/PermissionGate.vue";
import AdminPanel from "./components/AdminPanel.vue";

const sidebarCollapsed = ref(false);
const isDark = ref(false);
const adminOpen = ref(false);

const permissions = ref(window.__PERMISSIONS__ ?? {});
const user = ref(window.__USER__ ?? { id: 0 });

provide(PermissionsKey, permissions);
provide(UserKey, user);

function toggleTheme() {
  isDark.value = !isDark.value;
  document.documentElement.classList.toggle("dark", isDark.value);
  localStorage.setItem("theme", isDark.value ? "dark" : "light");
}

onMounted(() => {
  isDark.value = localStorage.getItem("theme") === "dark";
  document.documentElement.classList.toggle("dark", isDark.value);

  // Intercept clicks on sidebar admin link to open the slide-over
  document.addEventListener("click", (e) => {
    const link = e.target.closest('a[href="#"]');
    if (link && link.textContent?.trim() === "Admin") {
      e.preventDefault();
      adminOpen.value = true;
    }
  });
});
</script>
