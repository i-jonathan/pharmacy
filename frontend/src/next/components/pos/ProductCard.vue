<template>
  <div
    class="flex h-full min-h-[6.5rem] flex-col rounded-md border border-border bg-card p-3 transition-colors hover:border-primary/35 hover:bg-primary/[0.025] cursor-pointer"
    @click="$emit('add')"
  >
    <div class="text-sm font-semibold truncate">{{ product.name }}</div>
    <div class="text-xs text-muted-foreground truncate mt-0.5">
      {{ product.manufacturer || product.category }}
    </div>
    <div class="mt-auto flex flex-col items-start gap-1.5 pt-3 xl:flex-row xl:items-center xl:justify-between xl:gap-2">
      <span class="text-sm font-semibold tabular-nums">&#8358;{{ (product.price || 0).toLocaleString() }}</span>
      <span class="whitespace-nowrap text-xs tabular-nums text-emerald-700 dark:text-emerald-400">Stock · {{ product.stock }}</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  product: { type: Object, required: true },
});

defineEmits(["add"]);
</script>
