<template>
  <div
    class="flex flex-col p-3 rounded-lg border border-border bg-card hover:shadow-sm transition-shadow cursor-pointer"
    @click="$emit('add')"
  >
    <div class="text-sm font-semibold truncate">{{ product.name }}</div>
    <div class="text-xs text-muted-foreground truncate mt-0.5">
      {{ product.manufacturer || product.category }}
    </div>
    <div class="mt-2 flex items-center justify-between">
      <span class="text-sm font-bold">&#8358;{{ (product.price || 0).toLocaleString() }}</span>
      <span class="text-xs text-emerald-600">Stock: {{ product.stock }}</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  product: { type: Object, required: true },
});

defineEmits(["add"]);
</script>
