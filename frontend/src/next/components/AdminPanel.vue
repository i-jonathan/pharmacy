<template>
  <Teleport to="body">
    <Transition name="slide">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex justify-end"
        @click.self="$emit('close')"
      >
        <div class="absolute inset-0 bg-black/40" />
        <div class="relative w-full max-w-2xl bg-background border-l shadow-2xl overflow-y-auto h-full">
          <div class="sticky top-0 bg-background border-b px-6 py-4 flex items-center justify-between z-10">
            <div>
              <h2 class="text-lg font-semibold">Administration</h2>
              <p class="text-sm text-muted-foreground">Manage permissions, roles, users, and categories</p>
            </div>
            <Button variant="ghost" size="icon" @click="$emit('close')">
              <X :size="18" />
            </Button>
          </div>

          <div class="p-6">
            <div v-if="!currentModule">
              <div class="grid grid-cols-1 gap-4">
                <Card
                  v-for="mod in modules"
                  :key="mod.path"
                  class="cursor-pointer hover:border-primary/50 transition-colors"
                  @click="currentModule = mod.path"
                >
                  <CardHeader class="pb-2">
                    <CardTitle class="text-base flex items-center gap-2">
                      <component :is="mod.icon" :size="18" class="text-primary" />
                      {{ mod.name }}
                    </CardTitle>
                    <CardDescription>{{ mod.description }}</CardDescription>
                  </CardHeader>
                </Card>
              </div>
            </div>

            <div v-else>
              <Button variant="ghost" size="sm" class="mb-4" @click="currentModule = null">
                <ArrowLeft :size="14" class="mr-1" /> Back
              </Button>

              <PermissionsManager v-if="currentModule === 'permissions'" />
              <RolesViewer v-else-if="currentModule === 'roles'" />
              <UsersManager v-else-if="currentModule === 'users'" />
              <CategoriesManager v-else-if="currentModule === 'categories'" />
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch } from "vue";
import { X, ArrowLeft, Shield, Users, UserCog, Tags } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import PermissionsManager from "../../admin/views/PermissionsManager.vue";
import RolesViewer from "../../admin/views/RolesViewer.vue";
import UsersManager from "../../admin/views/UsersManager.vue";
import CategoriesManager from "../../admin/views/CategoriesManager.vue";

const props = defineProps({
  open: { type: Boolean, default: false },
  initialModule: { type: String, default: null },
});

defineEmits(["close"]);

const currentModule = ref(null);

watch(() => props.open, (isOpen) => {
  if (isOpen && props.initialModule) {
    currentModule.value = props.initialModule;
  } else if (!isOpen) {
    currentModule.value = null;
  }
});

const modules = [
  { name: "Permissions", description: "Manage access permissions and assign them to roles", icon: Shield, path: "permissions" },
  { name: "Roles", description: "View roles and their permission assignments", icon: Users, path: "roles" },
  { name: "Users", description: "Manage users, change roles, reset passwords", icon: UserCog, path: "users" },
  { name: "Categories", description: "Manage product categories", icon: Tags, path: "categories" },
];
</script>

<style scoped>
.slide-enter-active,
.slide-leave-active {
  transition: opacity 0.2s ease;
}
.slide-enter-active > div:last-child,
.slide-leave-active > div:last-child {
  transition: transform 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
}
.slide-enter-from > div:last-child {
  transform: translateX(100%);
}
.slide-leave-to > div:last-child {
  transform: translateX(100%);
}
</style>
