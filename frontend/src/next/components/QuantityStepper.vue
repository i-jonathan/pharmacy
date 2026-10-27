<template>
  <div
    class="inline-flex items-center overflow-hidden rounded-md border border-border"
    :class="{ 'opacity-50': disabled }"
  >
    <button
      type="button"
      class="flex h-7 w-7 items-center justify-center text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
      :disabled="disabled || modelValue <= min"
      :aria-label="`Decrease ${label}`"
      @click="set(modelValue - 1)"
    >
      −
    </button>
    <input
      :value="text"
      inputmode="numeric"
      class="no-spinners h-7 w-12 border-x border-border bg-transparent text-center text-sm tabular-nums outline-none"
      :aria-label="label"
      :disabled="disabled"
      @input="onInput"
      @blur="commit"
      @keydown.enter.prevent="commit"
    />
    <button
      type="button"
      class="flex h-7 w-7 items-center justify-center text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
      :disabled="disabled || modelValue >= max"
      :aria-label="`Increase ${label}`"
      @click="set(modelValue + 1)"
    >
      +
    </button>
  </div>
</template>

<script setup>
import { ref, watch } from "vue";

const props = defineProps({
  modelValue: { type: Number, default: 1 },
  min: { type: Number, default: 1 },
  max: { type: Number, default: 1_000_000 },
  label: { type: String, default: "quantity" },
  disabled: { type: Boolean, default: false },
});
const emit = defineEmits(["update:modelValue"]);

const text = ref(String(props.modelValue));
watch(() => props.modelValue, (v) => { text.value = String(v); });

function clamp(v) {
  return Math.min(props.max, Math.max(props.min, v));
}

function set(v) {
  emit("update:modelValue", clamp(v));
}

// Typing is free-form (it may be empty mid-edit); the value is clamped when the field loses focus.
function onInput(e) {
  const digits = e.target.value.replace(/\D/g, "");
  text.value = digits;
  e.target.value = digits;
  if (digits) emit("update:modelValue", Math.min(props.max, parseInt(digits, 10)));
}

function commit() {
  const value = clamp(text.value ? parseInt(text.value, 10) : props.min);
  text.value = String(value);
  emit("update:modelValue", value);
}
</script>
