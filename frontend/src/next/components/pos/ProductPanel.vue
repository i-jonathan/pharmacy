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
    <div class="mt-5">
      <h2 class="text-sm font-semibold mb-2">Frequently Sold</h2>
      <ProductCarousel
        v-if="frequentlySold.length"
        :items="frequentlySold"
        :slides-per-view="4"
        @add-item="$emit('add-item', $event)"
      />
      <div v-else class="text-sm text-muted-foreground text-center py-4">
        No products available
      </div>
    </div>

    <!-- All Products -->
    <div class="mt-5 flex-1 overflow-auto">
      <h2 class="text-sm font-semibold mb-2">All Products</h2>
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
              @click="product.stock > 0 && $emit('add-item', product)"
            >
              <TableCell>
                <div class="text-sm font-medium">{{ product.name }}</div>
                <div v-if="product.manufacturer" class="text-xs text-muted-foreground">{{ product.manufacturer }}</div>
              </TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ product.category }}</TableCell>
              <TableCell class="text-sm font-medium">&#8358;{{ product.price.toLocaleString() }}</TableCell>
              <TableCell class="text-sm" :class="product.stock > 0 ? 'text-emerald-600' : 'text-destructive'">{{ product.stock }}</TableCell>
              <TableCell>
                <Button
                  variant="outline"
                  size="icon"
                  class="h-7 w-7"
                  :disabled="product.stock <= 0"
                  @click.stop="product.stock > 0 && $emit('add-item', product)"
                >
                  <Plus :size="14" />
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <div v-else class="text-sm text-muted-foreground text-center py-6">
        {{ searchQuery ? 'No products match your search' : 'No products available' }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { Search, Plus } from "lucide-vue-next";
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
const allProducts = ref([]);
const topSelling = ref([]);
const searchInput = ref(null);

onMounted(async () => {
  emit("search-ref", searchInput.value);

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
});

const frequentlySold = computed(() => {
  return topSelling.value
    .map((ts) => allProducts.value.find((p) => p.name === ts.product_name))
    .filter(Boolean)
    .slice(0, 10);
});

const filteredProducts = computed(() => {
  let items = allProducts.value;

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase();
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
