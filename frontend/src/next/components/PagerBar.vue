<template>
  <div v-if="total > size" class="flex items-center justify-between gap-3 text-xs text-muted-foreground">
    <span>Showing {{ from }}–{{ to }} of {{ total }}</span>
    <div class="flex items-center gap-1">
      <button
        type="button"
        class="rounded-md border border-border px-2.5 py-1 hover:bg-accent disabled:cursor-not-allowed disabled:opacity-40"
        :disabled="page <= 1"
        @click="$emit('update:page', page - 1)"
      >
        Previous
      </button>
      <span class="px-2 tabular-nums">{{ page }} / {{ pages }}</span>
      <button
        type="button"
        class="rounded-md border border-border px-2.5 py-1 hover:bg-accent disabled:cursor-not-allowed disabled:opacity-40"
        :disabled="page >= pages"
        @click="$emit('update:page', page + 1)"
      >
        Next
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  page: { type: Number, required: true },
  total: { type: Number, required: true },
  size: { type: Number, required: true },
});
defineEmits(["update:page"]);

const pages = computed(() => Math.max(1, Math.ceil(props.total / props.size)));
const from = computed(() => (props.page - 1) * props.size + 1);
const to = computed(() => Math.min(props.total, props.page * props.size));
</script>
