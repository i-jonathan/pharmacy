<template>
  <div>
    <div class="relative">
      <Building :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
      <input
        :value="modelValue"
        type="text"
        :placeholder="placeholder"
        :aria-label="ariaLabel"
        autocomplete="off"
        class="no-spinners w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
        :class="{ 'border-red-500': invalid }"
        @input="onInput($event.target.value)"
        @keydown.escape="suggestions = []"
      />
    </div>
    <ul v-if="suggestions.length" class="mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-hidden">
      <li
        v-for="name in suggestions"
        :key="name"
        class="flex items-center gap-2 px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors"
        @click="select(name)"
      >
        <Building :size="14" class="text-muted-foreground" aria-hidden="true" />
        <span>{{ name }}</span>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { onBeforeUnmount, ref } from "vue";
import { Building } from "lucide-vue-next";

// The supplier field shared by Receive Items and Purchase Orders: type a name, pick an existing
// supplier from the list or keep what was typed (a new name creates the supplier on save).
defineProps({
  modelValue: { type: String, default: "" },
  placeholder: { type: String, default: "Enter supplier name..." },
  ariaLabel: { type: String, default: "Supplier" },
  invalid: { type: Boolean, default: false },
});
const emit = defineEmits(["update:modelValue"]);

const suggestions = ref([]);
let timer = null;

function onInput(text) {
  emit("update:modelValue", text);
  clearTimeout(timer);
  const q = text.trim();
  if (!q) {
    suggestions.value = [];
    return;
  }
  timer = setTimeout(async () => {
    try {
      const r = await fetch(`/inventory/suppliers/search?query=${encodeURIComponent(q)}`);
      if (r.ok) suggestions.value = ((await r.json()).data) || [];
    } catch {
      suggestions.value = [];
    }
  }, 200);
}

function select(name) {
  clearTimeout(timer);
  emit("update:modelValue", name);
  suggestions.value = [];
}

onBeforeUnmount(() => clearTimeout(timer));
</script>
