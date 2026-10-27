<template>
  <div class="space-y-2">
    <select
      :value="preset"
      class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm md:max-w-sm"
      aria-label="Reason"
      @change="$emit('update:preset', $event.target.value)"
    >
      <option value="" disabled>Choose a reason</option>
      <option v-for="r in reasons" :key="r" :value="r">{{ r }}</option>
    </select>
    <input
      v-if="preset"
      :value="detail"
      maxlength="250"
      class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm md:max-w-sm"
      :placeholder="preset === 'Other' ? 'Describe the reason' : 'More detail (optional)'"
      aria-label="Reason detail"
      @input="$emit('update:detail', $event.target.value)"
    />
  </div>
</template>

<script setup>
import { REASONS } from "./reasons.js";
defineProps({
  preset: { type: String, default: "" },
  detail: { type: String, default: "" },
  reasons: { type: Array, default: () => REASONS },
});
defineEmits(["update:preset", "update:detail"]);

</script>
