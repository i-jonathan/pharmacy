<template>
  <Transition name="whats-new-fade">
    <div
      v-if="open && slides.length"
      class="fixed inset-0 z-[70] flex items-end justify-center bg-neutral-950/55 backdrop-blur-sm sm:items-center sm:p-6"
      role="presentation"
      @click.self="close"
    >
      <section
        ref="panel"
        class="flex max-h-[96dvh] w-full max-w-3xl flex-col overflow-hidden rounded-t-2xl border border-border bg-card shadow-2xl sm:rounded-2xl"
        role="dialog"
        aria-modal="true"
        aria-labelledby="whats-new-title"
        aria-describedby="whats-new-text"
        tabindex="-1"
      >
        <header class="flex items-start justify-between gap-3 border-b border-border px-5 py-4 sm:px-6">
          <div>
            <p class="text-xs font-medium text-muted-foreground">
              {{ release.title }}
            </p>
            <h2 id="whats-new-title" class="mt-0.5 text-lg font-semibold text-foreground">{{ slide.title }}</h2>
          </div>
          <Button variant="ghost" size="icon" aria-label="Close" @click="close">
            <X :size="16" />
          </Button>
        </header>

        <div class="min-h-0 flex-1 overflow-y-auto bg-muted/30 px-4 pt-4 sm:px-6">
          <a :href="slide.image" target="_blank" rel="noopener" title="Open the picture full size">
            <img
              :key="slide.id"
              :src="slide.image"
              :alt="slide.alt"
              class="mx-auto block aspect-[11/7] w-full rounded-lg border border-border bg-white object-cover object-top shadow-sm"
            />
          </a>
          <p id="whats-new-text" class="py-4 text-sm leading-6 text-foreground">{{ slide.text }}</p>
        </div>

        <footer class="flex items-center justify-between gap-3 border-t border-border px-5 py-4 sm:px-6">
          <div class="flex items-center gap-2" role="group" aria-label="Choose a slide">
            <button
              v-for="(s, i) in slides"
              :key="s.id"
              type="button"
              class="h-2.5 w-2.5 rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :class="i === index ? 'bg-primary' : 'bg-border hover:bg-muted-foreground'"
              :aria-label="`Slide ${i + 1} of ${slides.length}: ${s.title}`"
              :aria-current="i === index ? 'true' : undefined"
              @click="go(i)"
            ></button>
          </div>
          <div class="flex gap-2">
            <Button variant="outline" :disabled="index === 0" @click="go(index - 1)">
              <ChevronLeft :size="15" class="mr-1" /> Back
            </Button>
            <Button v-if="index < slides.length - 1" @click="go(index + 1)">
              Next <ChevronRight :size="15" class="ml-1" />
            </Button>
            <Button v-else @click="close">Done</Button>
          </div>
        </footer>
      </section>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, watch, nextTick, onUnmounted } from "vue";
import { X, ChevronLeft, ChevronRight } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { useWhatsNew } from "../composables/useWhatsNew.js";

const { open, release, slides, close } = useWhatsNew();

const panel = ref(null);
const index = ref(0);
const slide = computed(() => slides.value[index.value] ?? slides.value[0]);

function go(i) {
  if (i < 0 || i >= slides.value.length) return;
  index.value = i;
  preload(i + 1);
}

// The next picture is fetched while the current one is read, so Next doesn't wait.
function preload(i) {
  const next = slides.value[i];
  if (next) new Image().src = next.image;
}

// Keys are heard on the window while the dialog is open: swapping Next for Done on the last slide
// removes the focused button, and a keydown handler on the dialog would stop hearing Esc.
function onKeydown(e) {
  if (e.defaultPrevented) return; // another overlay already took this key
  if (e.key === "Tab") return trapTab(e);
  // Arrow keys belong to a field when one has focus.
  if (e.target.closest?.("input, textarea, select, [contenteditable]")) return;
  if (e.key === "Escape") close();
  else if (e.key === "ArrowLeft") go(index.value - 1);
  else if (e.key === "ArrowRight") go(index.value + 1);
  else return;
  e.preventDefault();
}

// Tab stays inside the dialog: the page behind it is covered, so it must not take focus.
function trapTab(e) {
  const items = [...(panel.value?.querySelectorAll("button:not(:disabled), a[href]") ?? [])];
  if (!items.length) return;
  const first = items[0];
  const last = items[items.length - 1];
  const active = document.activeElement;
  if (!panel.value.contains(active)) {
    e.preventDefault();
    first.focus();
  } else if (e.shiftKey && (active === first || active === panel.value)) {
    e.preventDefault();
    last.focus();
  } else if (!e.shiftKey && active === last) {
    e.preventDefault();
    first.focus();
  }
}

let opener = null; // what had focus before the dialog opened, so closing gives it back
let pageOverflow = ""; // the page's scrolling before the dialog locked it

watch(open, async (isOpen) => {
  if (!isOpen) {
    window.removeEventListener("keydown", onKeydown);
    document.body.style.overflow = pageOverflow;
    await nextTick();
    if (opener?.isConnected) opener.focus();
    opener = null;
    return;
  }
  opener = document.activeElement;
  pageOverflow = document.body.style.overflow;
  document.body.style.overflow = "hidden"; // the page behind the backdrop must not scroll
  index.value = 0;
  preload(1);
  window.addEventListener("keydown", onKeydown);
  await nextTick();
  panel.value?.focus();
});

onUnmounted(() => {
  window.removeEventListener("keydown", onKeydown);
  if (open.value) document.body.style.overflow = pageOverflow;
});
</script>

<style scoped>
.whats-new-fade-enter-active,
.whats-new-fade-leave-active {
  transition: opacity 0.15s ease;
}
.whats-new-fade-enter-from,
.whats-new-fade-leave-to {
  opacity: 0;
}
@media (prefers-reduced-motion: reduce) {
  .whats-new-fade-enter-active,
  .whats-new-fade-leave-active {
    transition: none;
  }
}
</style>
