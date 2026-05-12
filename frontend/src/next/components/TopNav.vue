<template>
  <div class="flex items-center justify-between px-6 h-14 border-b border-border bg-background">
    <!-- Left: Breadcrumbs + Title -->
    <div>
      <h1 class="text-lg font-semibold leading-tight">{{ route.meta.title }}</h1>
      <p v-if="route.meta.subtitle" class="text-xs text-muted-foreground">
        {{ route.meta.subtitle }}
      </p>
      <nav v-if="breadcrumbs.length > 1" class="flex items-center gap-1.5 text-xs text-muted-foreground mb-0.5">
        <template v-for="(crumb, i) in breadcrumbs" :key="i">
          <router-link
            v-if="crumb.to"
            :to="crumb.to"
            class="hover:text-foreground transition-colors"
          >
            {{ crumb.label }}
          </router-link>
          <span v-else class="text-foreground font-medium">{{ crumb.label }}</span>
          <ChevronRight :stroke-width="1.5" v-if="i < breadcrumbs.length - 1" :size="12" />
        </template>
      </nav>
    </div>

    <!-- Right: Actions -->
    <div class="flex items-center gap-3">
      <!-- Notification Bell -->
      <Button variant="ghost" size="icon" class="relative">
        <Bell :size="18" :stroke-width="1.5" />
      </Button>

      <!-- User Dropdown -->
      <div class="relative">
        <Button
          variant="ghost"
          class="flex items-center gap-2 h-auto py-1.5"
          @click="open = !open"
        >
          <div class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
            <UserIcon :size="16" class="text-primary" :stroke-width="1.5" />
          </div>
          <div class="text-left hidden sm:block">
            <div class="text-sm font-medium leading-none">{{ user.username }}</div>
            <div class="text-xs text-muted-foreground">{{ user.role }}</div>
          </div>
          <ChevronDown :stroke-width="1.5" :size="14" class="text-muted-foreground hidden sm:block" />
        </Button>

        <!-- Dropdown -->
        <div
          v-if="open"
          class="absolute right-0 top-full mt-1 w-48 rounded-md border border-border bg-popover shadow-lg z-50"
        >
          <div class="px-3 py-2 border-b border-border">
            <div class="text-sm font-medium">{{ user.username }}</div>
            <div class="text-xs text-muted-foreground">{{ user.role }}</div>
          </div>
          <a
            href="/user/logout"
            class="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
          >
            <LogOut :size="14" :stroke-width="1.5" />
            Logout
          </a>
        </div>
      </div>

      <!-- Backdrop -->
      <div v-if="open" class="fixed inset-0 z-40" @click="open = false" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject } from "vue";
import { useRoute } from "vue-router";
import { Bell, User as UserIcon, ChevronDown, LogOut, ChevronRight } from "lucide-vue-next";
import { UserKey } from "../composables/usePermissions.js";
import { Button } from "@/components/ui/button";

const route = useRoute();
const user = inject(UserKey, { id: 0, username: "User", role: "" });
const open = ref(false);

const breadcrumbs = computed(() => {
  const crumbs = [];
  if (route.meta.parent) {
    crumbs.push({ label: route.meta.parent, to: { name: "dashboard" } });
  }
  crumbs.push({ label: route.meta.title });
  return crumbs;
});
</script>
