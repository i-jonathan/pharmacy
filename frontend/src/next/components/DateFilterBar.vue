<template>
  <div class="flex items-center gap-2">
    <Button
      v-for="option in options"
      :key="option.value"
      :variant="modelValue === option.value ? 'default' : 'outline'"
      size="sm"
      @click="$emit('update:modelValue', option.value)"
    >
      {{ option.label }}
    </Button>
  </div>
</template>

<script setup>
import { Button } from "@/components/ui/button";

defineProps({
  modelValue: { type: String, default: "today" },
});

defineEmits(["update:modelValue"]);

const options = [
  { label: "Today", value: "today" },
  { label: "This Week", value: "week" },
  { label: "This Month", value: "month" },
  { label: "This Year", value: "year" },
];
</script>
