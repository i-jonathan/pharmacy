<template>
  <div class="relative">
    <div class="overflow-hidden rounded-lg">
      <div
        ref="scrollContainer"
        class="flex overflow-x-auto snap-x snap-mandatory scroll-smooth gap-3"
        @scroll="onScroll"
      >
        <div
          v-for="(product, i) in items"
          :key="product.id"
          class="snap-start shrink-0"
          :style="{ width: `${100 / props.slidesPerView}%` }"
        >
          <ProductCard :product="product" @add="$emit('add-item', product)" />
        </div>
      </div>
    </div>

    <!-- Prev Button -->
    <button
      v-if="canScrollPrev"
      class="absolute -left-3 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full border border-border bg-background shadow-sm flex items-center justify-center hover:bg-accent transition-colors"
      @click="scrollPrev"
    >
      <ChevronLeft :size="16" />
    </button>

    <!-- Next Button -->
    <button
      v-if="canScrollNext"
      class="absolute -right-3 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full border border-border bg-background shadow-sm flex items-center justify-center hover:bg-accent transition-colors"
      @click="scrollNext"
    >
      <ChevronRight :size="16" />
    </button>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import ProductCard from "./ProductCard.vue";

const props = defineProps({
  items: { type: Array, default: () => [] },
  slidesPerView: { type: Number, default: 4 },
});

defineEmits(["add-item"]);

const scrollContainer = ref(null);
const canScrollPrev = ref(false);
const canScrollNext = ref(true);

function onScroll() {
  const el = scrollContainer.value;
  if (!el) return;
  canScrollPrev.value = el.scrollLeft > 2;
  canScrollNext.value = el.scrollLeft + el.clientWidth < el.scrollWidth - 2;
}

function scrollPrev() {
  const el = scrollContainer.value;
  if (!el) return;
  const width = el.clientWidth;
  el.scrollBy({ left: -width, behavior: "smooth" });
}

function scrollNext() {
  const el = scrollContainer.value;
  if (!el) return;
  const width = el.clientWidth;
  el.scrollBy({ left: width, behavior: "smooth" });
}

onMounted(() => {
  const el = scrollContainer.value;
  if (el) {
    el.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
  }
});

onUnmounted(() => {
  const el = scrollContainer.value;
  if (el) {
    el.removeEventListener("scroll", onScroll);
  }
});
</script>
