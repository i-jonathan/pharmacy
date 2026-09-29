<template>
  <slot v-if="hasPermission(permission)" />
</template>

<script setup>
import { usePermissions } from "../composables/usePermissions.js";

const props = defineProps({
  permission: { type: String, required: true },
});

const { hasPermission } = usePermissions();
</script>
