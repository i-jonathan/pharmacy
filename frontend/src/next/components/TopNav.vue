<template>
  <header class="sticky top-0 z-30 flex min-h-14 items-center justify-between gap-3 border-b border-border bg-background/95 px-4 backdrop-blur sm:px-6">
    <!-- Left: Breadcrumbs + Title -->
    <div class="flex min-w-0 items-center gap-3">
      <Button variant="ghost" size="icon" class="shrink-0 lg:hidden" aria-label="Open navigation" @click="$emit('toggle-navigation')">
        <Menu :size="19" />
      </Button>
      <div class="min-w-0">
        <template v-if="route.meta.parent">
          <nav class="flex items-center gap-1.5 text-xs text-muted-foreground" aria-label="Breadcrumb">
            <router-link v-if="route.meta.parentRoute" :to="{ name: route.meta.parentRoute }" class="transition-colors hover:text-foreground">{{ route.meta.parent }}</router-link>
            <span v-else>{{ route.meta.parent }}</span>
            <ChevronRight :size="12" :stroke-width="1.5" aria-hidden="true" />
            <span class="truncate font-medium text-foreground">{{ route.meta.title }}</span>
          </nav>
        </template>
        <template v-else>
          <h1 class="truncate text-base font-semibold leading-tight sm:text-lg">{{ route.meta.title }}</h1>
          <p v-if="route.meta.subtitle" class="text-xs text-muted-foreground">{{ route.meta.subtitle }}</p>
        </template>
      </div>
    </div>

    <!-- Right: Actions -->
    <div class="flex items-center gap-3">
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
            @click.prevent="logout"
          >
            <LogOut :size="14" :stroke-width="1.5" />
            Logout
          </a>
        </div>
      </div>

      <!-- Backdrop -->
      <div v-if="open" class="fixed inset-0 z-40" @click="open = false" />
    </div>
  </header>
</template>

<script setup>
import { ref, inject } from "vue";
import { useRoute } from "vue-router";
import { Menu, User as UserIcon, ChevronDown, LogOut, ChevronRight } from "lucide-vue-next";
import { UserKey } from "../composables/usePermissions.js";
import { discardPersistedPosState } from "../composables/usePos.js";
import { Button } from "@/components/ui/button";

defineEmits(["toggle-navigation"]);
const route = useRoute();
const user = inject(UserKey, { id: 0, username: "User", role: "" });
const open = ref(false);

function logout() {
  let intercepted = false;
  const request = new CustomEvent("pharmacy:logout-request", {
    detail: {
      intercept() { intercepted = true; },
      proceed: completeLogout,
    },
  });
  window.dispatchEvent(request);
  if (!intercepted) completeLogout();
}

function completeLogout() {
  discardPersistedPosState();
  window.location.assign("/user/logout");
}
</script>
