<template>
  <Transition name="fade">
    <div
      v-if="state.open"
      class="fixed inset-0 z-[70] flex items-center justify-center bg-neutral-950/55 p-4 backdrop-blur-sm"
      role="presentation"
      @click.self="cancel"
      @keydown.esc="cancel"
    >
      <section
        ref="panel"
        class="w-full max-w-md rounded-xl border border-border bg-card p-5 shadow-2xl sm:p-6"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-title"
        aria-describedby="confirm-description"
        tabindex="-1"
      >
        <div class="mb-4 flex items-start gap-3">
          <span
            class="grid h-10 w-10 shrink-0 place-items-center rounded-full"
            :class="danger ? 'bg-destructive/10' : 'bg-amber-100 dark:bg-amber-900/30'"
          >
            <AlertTriangle
              :size="20"
              :class="danger ? 'text-destructive' : 'text-amber-700 dark:text-amber-300'"
              aria-hidden="true"
            />
          </span>
          <div>
            <h2 id="confirm-title" class="text-base font-semibold">{{ state.title }}</h2>
            <p v-if="state.message" id="confirm-description" class="mt-1 text-sm leading-5 text-muted-foreground">
              {{ state.message }}
            </p>
          </div>
        </div>
        <p v-if="state.detail" class="mb-4 text-sm font-medium text-foreground">{{ state.detail }}</p>
        <ul v-if="state.lines.length" class="mb-5 space-y-1 rounded-md border border-border bg-muted/30 px-3 py-2 text-sm">
          <li v-for="line in state.lines" :key="line">{{ line }}</li>
        </ul>
        <p
          v-if="state.error"
          class="mb-4 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive"
          role="alert"
        >
          {{ state.error }}
        </p>
        <div class="flex flex-col-reverse gap-2 sm:flex-row sm:flex-wrap sm:justify-end">
          <Button v-if="state.alternate" :variant="state.alternate.danger ? 'destructive' : 'ghost'" :disabled="state.busy" @click="chooseAlternate">
            {{ state.alternate.label }}
          </Button>
          <Button variant="outline" :disabled="state.busy" @click="cancel">{{ state.cancelLabel }}</Button>
          <Button :variant="danger ? 'destructive' : 'default'" :disabled="state.busy" @click="accept">
            <span
              v-if="state.busy"
              class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-r-transparent"
              aria-hidden="true"
            ></span>
            {{ state.busy ? state.busyLabel : state.confirmLabel }}
          </Button>
        </div>
      </section>
    </div>
  </Transition>
</template>

<script setup>
import { computed, nextTick, ref, watch } from "vue";
import { AlertTriangle } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { useConfirm } from "../composables/useConfirm.js";

const { state, accept, chooseAlternate, cancel } = useConfirm();
const panel = ref(null);
const danger = computed(() => state.tone === "danger");

// Focus the dialog so Escape works without a click, and so Enter on the focused button confirms.
watch(
  () => state.open,
  async (open) => {
    if (!open) return;
    await nextTick();
    panel.value?.focus();
  },
);
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
