<template>
  <div class="flex flex-col h-full">
    <!-- Search Bar -->
    <div class="relative">
      <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
      <input
        ref="searchInput"
        v-model="searchQuery"
        type="text"
        placeholder="Search medicine by name, brand or generic..."
        class="w-full pl-9 pr-14 py-2.5 rounded-lg border border-border bg-background text-sm focus:outline-none focus:border-primary"
      />
      <kbd class="absolute right-2 top-1/2 -translate-y-1/2 px-1.5 py-0.5 text-xs rounded bg-muted text-muted-foreground border border-border">
        F3
      </kbd>
    </div>

    <!-- Frequently Sold -->
    <div v-if="!debouncedQuery" class="mt-5">
      <h2 class="text-sm font-semibold mb-2">Frequently Sold</h2>
      <ProductCarousel
        v-if="frequentlySold.length"
        :items="frequentlySold"
        :slides-per-view="4"
        @add-item="onAddItem($event, null, null)"
      />
      <div v-else class="text-sm text-muted-foreground text-center py-4">
        No products available
      </div>
    </div>

    <!-- All Products -->
    <div class="mt-5 flex-1 overflow-auto">
      <div class="flex items-center gap-2 mb-2">
        <h2 class="text-sm font-semibold">All Products</h2>
        <button
          class="w-5 h-5 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
          :class="{ 'animate-spin': refreshing }"
          @click="refreshProducts"
        >
          <RotateCw :size="13" />
        </button>
      </div>
      <div v-if="filteredProducts.length">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Product</TableHead>
              <TableHead>Category</TableHead>
              <TableHead>Price</TableHead>
              <TableHead>Stock</TableHead>
              <TableHead class="w-12"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="product in filteredProducts"
              :key="product.id"
              class="cursor-pointer hover:bg-accent/50"
              :class="{ 'opacity-60': product.stock <= 0 }"
              @click="product.stock > 0 && onAddItem(product, null, null)"
            >
              <TableCell>
                <div class="text-sm font-medium">{{ product.name }}</div>
                <div v-if="product.manufacturer" class="text-xs text-muted-foreground">{{ product.manufacturer }}</div>
              </TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ product.category }}</TableCell>
              <TableCell class="text-sm font-medium">&#8358;{{ product.price.toLocaleString() }}</TableCell>
              <TableCell class="text-sm" :class="product.stock > 0 ? 'text-emerald-600' : 'text-destructive'">{{ product.stock }}</TableCell>
              <TableCell class="relative">
                <Button
                  variant="outline"
                  size="icon"
                  class="h-7 w-7"
                  :disabled="product.stock <= 0"
                  @click.stop="product.stock > 0 && onPlusClick($event, product)"
                >
                  <Plus :size="14" />
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <div v-else class="text-sm text-muted-foreground text-center py-6">
        {{ debouncedQuery ? 'No products match your search' : 'No products available' }}
      </div>
    </div>

    <!-- Price Options Popover -->
    <Teleport to="body">
      <div
        v-if="popover.product"
        class="fixed z-60 w-48 rounded-sm border border-border bg-popover shadow-lg p-1"
        :style="{ top: popover.y + 'px', left: popover.x + 'px' }"
        @click.stop
      >
        <div class="text-xs text-muted-foreground px-2 py-1.5 border-b border-border">
          Select price option
        </div>
        <button
          v-for="opt in priceOptionsForCurrent"
          :key="opt.id"
          class="flex items-center justify-between w-full px-2 py-1.5 text-sm rounded-sm hover:bg-primary/15 transition-colors"
          @click="selectPriceOption(opt)"
        >
          <span>{{ opt.name || 'Base' }}</span>
          <span class="font-medium">&#8358;{{ (opt.selling_price / 100).toLocaleString() }}</span>
        </button>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { Search, Plus, RotateCw } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import ProductCarousel from "./ProductCarousel.vue";

const emit = defineEmits(["add-item", "search-ref"]);

const searchQuery = ref("");
const debouncedQuery = ref("");
const allProducts = ref([]);
const topSelling = ref([]);
const searchInput = ref(null);

let debounceTimer = null;
watch(searchQuery, (val) => {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    debouncedQuery.value = val;
  }, 300);
});

const popover = ref({ product: null, x: 0, y: 0 });
const refreshing = ref(false);

async function loadProducts() {
  try {
    const [prodResp, topResp] = await Promise.all([
      fetch("/inventory/item-list"),
      fetch("/inventory/top-selling?limit=10"),
    ]);

    if (prodResp.ok) {
      const data = await prodResp.json();
      allProducts.value = (data.items || []).map((p) => ({
        ...p,
        price: p.default_price / 100,
        priceId: p.default_price_id,
      }));
    }
    if (topResp.ok) {
      topSelling.value = await topResp.json();
    }
  } catch (e) {
    console.error("Failed to load POS data:", e);
  }
}

async function refreshProducts() {
  refreshing.value = true;
  await loadProducts();
  refreshing.value = false;
}

onMounted(async () => {
  emit("search-ref", searchInput.value);
  searchInput.value?.focus();
  await loadProducts();
});

const priceOptionsForCurrent = computed(() => {
  const p = popover.value.product;
  if (!p || !p.price_options) return [];
  return p.price_options.filter((opt) => opt.id !== p.priceId);
});

function hasMultiplePriceOptions(product) {
  return product.price_options && product.price_options.length > 1;
}

function onAddItem(product, priceId, price) {
  emit("add-item", product, priceId, price);
  searchQuery.value = "";
  debouncedQuery.value = "";
}

function onPlusClick(event, product) {
  if (hasMultiplePriceOptions(product)) {
    const rect = event.target.getBoundingClientRect();
    popover.value = {
      product,
      x: Math.min(rect.left, window.innerWidth - 200),
      y: Math.min(rect.bottom + 4, window.innerHeight - 200),
    };
  } else {
    onAddItem(product, null, null);
  }
}

function selectPriceOption(opt) {
  const product = popover.value.product;
  onAddItem(product, opt.id, opt.selling_price / 100);
  closePopover();
}

function closePopover() {
  popover.value = { product: null, x: 0, y: 0 };
}

function onDocumentClick(e) {
  if (!popover.value.product) return;
  if (e.target.closest("button")) return;
  const el = document.querySelector(".fixed.z-\\[60\\]");
  if (el && !el.contains(e.target)) {
    closePopover();
  }
}

onMounted(() => document.addEventListener("click", onDocumentClick));
onUnmounted(() => document.removeEventListener("click", onDocumentClick));

const frequentlySold = computed(() => {
  return topSelling.value
    .map((ts) => allProducts.value.find((p) => p.name === ts.product_name))
    .filter(Boolean)
    .slice(0, 10);
});

const filteredProducts = computed(() => {
  let items = allProducts.value;

  if (debouncedQuery.value.trim()) {
    const q = debouncedQuery.value.toLowerCase();
    items = items.filter(
      (p) =>
        p.name?.toLowerCase().includes(q) ||
        p.manufacturer?.toLowerCase().includes(q) ||
        p.barcode?.toLowerCase().includes(q)
    );
  }

  return items;
});
</script>
