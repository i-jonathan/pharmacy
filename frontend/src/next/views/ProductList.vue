<template>
  <div class="flex h-[calc(100vh-3.5rem)]">
    <!-- Main table area -->
    <div class="flex-1 flex flex-col overflow-hidden" :class="{ 'border-r border-border': detailProduct }">
      <!-- Header -->
      <div class="p-6 pb-0">
        <h1 class="text-2xl font-bold text-foreground">Products</h1>
        <p class="text-sm text-muted-foreground mt-1">Manage your pharmacy inventory items</p>
      </div>

      <!-- Toolbar -->
      <div class="p-6 pb-4 flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by product name, manufacturer, or barcode..."
            class="w-full pl-9 pr-4 py-2 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
          />
        </div>
        <div class="flex gap-2">
          <select
            v-model="categoryFilter"
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring"
          >
            <option value="">All Categories</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
          </select>
          <Button variant="outline" :disabled="loading" @click="fetchProducts">
            <RotateCw :size="14" :class="{ 'animate-spin': loading }" class="mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      <!-- Content area -->
      <div class="flex-1 overflow-y-auto px-6 pb-4">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-24 text-muted-foreground">
          <RotateCw :size="20" class="animate-spin mr-3" />
          <span class="text-sm">Loading products...</span>
        </div>

        <!-- Error -->
        <div v-else-if="error" class="flex flex-col items-center justify-center py-24 text-center">
          <AlertCircle :size="40" class="text-destructive/40 mb-3" />
          <p class="text-sm text-muted-foreground mb-3">{{ error }}</p>
          <Button variant="outline" size="sm" @click="fetchProducts">Retry</Button>
        </div>

        <!-- Empty -->
        <div v-else-if="filteredProducts.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
          <Package :size="40" class="text-muted-foreground/40 mb-3" />
          <h3 class="text-base font-semibold text-foreground mb-1">No products found</h3>
          <p class="text-sm text-muted-foreground max-w-sm">
            {{ searchQuery || categoryFilter ? 'No products match your filters.' : 'No products in inventory yet.' }}
          </p>
        </div>

        <!-- Product Table -->
        <div v-else class="border border-border rounded-lg overflow-hidden">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Name</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Manufacturer</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Category</th>
                <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Price</th>
                <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Stock</th>
                <th v-if="canViewReorder" class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Reorder At</th>
                <th v-if="canEditInventory" class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Cost Price</th>
                <th v-if="canEditInventory" class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Expiry</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr
                v-for="p in paginatedProducts"
                :key="p.id"
                class="cursor-pointer transition-colors hover:bg-muted/20 outline-none"
                tabindex="0"
                @click="selectAndFetch(p.id)"
                @keydown.enter.prevent="selectAndFetch(p.id)"
              >
                <td class="px-4 py-3">
                  <div class="text-sm font-medium text-foreground">{{ p.name }}</div>
                </td>
                <td class="px-4 py-3 text-sm text-muted-foreground">{{ p.manufacturer || '—' }}</td>
                <td class="px-4 py-3 text-sm text-muted-foreground">
                  <span class="bg-muted text-muted-foreground px-2 py-0.5 rounded-full text-xs">{{ p.category }}</span>
                </td>
                <td class="px-4 py-3 text-sm text-right font-medium">&#8358;{{ (p.default_price / 100).toLocaleString() }}</td>
                <td class="px-4 py-3 text-sm text-right">
                  <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="stockClass(p.stock, p.reorder_level)">{{ p.stock }}</span>
                </td>
                <td v-if="canViewReorder" class="px-4 py-3 text-sm text-right text-muted-foreground">{{ p.reorder_level }}</td>
                <td v-if="canEditInventory" class="px-4 py-3 text-sm text-right text-muted-foreground">&#8358;{{ (p.cost_price / 100).toLocaleString() }}</td>
                <td v-if="canEditInventory" class="px-4 py-3 text-sm text-right text-muted-foreground">{{ formatDate(p.earliest_expiry) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="flex items-center justify-between pt-4">
          <span class="text-xs text-muted-foreground">{{ filteredProducts.length }} products</span>
          <div class="flex items-center gap-4">
            <span class="text-xs text-muted-foreground">Page {{ currentPage }} of {{ totalPages }}</span>
            <div class="flex gap-1">
              <Button variant="outline" size="sm" :disabled="currentPage <= 1" @click="goPage(currentPage - 1)">
                <ChevronLeft :size="14" />
              </Button>
              <Button v-for="p in visiblePages" :key="p" variant="outline" size="sm" :class="{ 'bg-primary/10 text-primary border-primary/30': p === currentPage }" @click="goPage(p)">{{ p }}</Button>
              <Button variant="outline" size="sm" :disabled="currentPage >= totalPages" @click="goPage(currentPage + 1)">
                <ChevronRight :size="14" />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Detail Sidebar -->
    <Transition name="slide-panel">
      <div v-if="detailProduct" class="flex flex-col h-full w-[40%] min-w-[360px] bg-card border-l border-border flex-shrink-0">
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-border">
          <div>
            <h2 class="text-sm font-semibold">{{ editMode ? 'Edit Product' : 'Product Details' }}</h2>
            <p class="text-xs text-muted-foreground font-mono">#{{ detailProduct.id }}</p>
          </div>
          <div class="flex items-center gap-1">
            <Button v-if="canEditInventory && !editMode" variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" title="Edit" @click="editMode = true">
              <Pencil :size="14" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" @click="closeDetail">
              <X :size="15" />
            </Button>
          </div>
        </div>

        <!-- Scrollable content -->
        <div class="flex-1 overflow-y-auto">
          <!-- View Mode -->
          <template v-if="!editMode">
            <div class="px-4 py-4 space-y-4">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <div class="text-xs text-muted-foreground mb-1">Name</div>
                  <div class="text-sm font-medium">{{ detailProduct.name }}</div>
                </div>
                <div>
                  <div class="text-xs text-muted-foreground mb-1">Manufacturer</div>
                  <div class="text-sm">{{ detailProduct.manufacturer || '—' }}</div>
                </div>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <div class="text-xs text-muted-foreground mb-1">Barcode / SKU</div>
                  <div class="text-sm font-mono">{{ detailProduct.barcode || '—' }}</div>
                </div>
                <div>
                  <div class="text-xs text-muted-foreground mb-1">Category</div>
                  <div class="text-sm">{{ detailProduct.category || categoryName(detailProduct.category_id) }}</div>
                </div>
              </div>
            </div>

            <div class="border-t border-border px-4 py-4">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <div class="text-xs text-muted-foreground mb-1">Selling Price</div>
                  <div class="text-sm font-semibold">&#8358;{{ detailProduct.default_price?.selling_price?.toLocaleString() }}</div>
                </div>
                <div>
                  <div class="text-xs text-muted-foreground mb-1">Cost Price</div>
                  <div class="text-sm">&#8358;{{ detailProduct.cost_price?.toLocaleString() || '—' }}</div>
                </div>
              </div>
              <div class="grid grid-cols-2 gap-4 mt-3">
                <div>
                  <div class="text-xs text-muted-foreground mb-1">Stock</div>
                  <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="stockClass(detailProduct.stock, detailProduct.reorder_level)">{{ detailProduct.stock }}</span>
                </div>
                <div>
                  <div class="text-xs text-muted-foreground mb-1">Reorder Level</div>
                  <div class="text-sm">{{ detailProduct.reorder_level }}</div>
                </div>
              </div>
            </div>

            <!-- Price Options -->
            <div v-if="detailProduct.price_options?.length" class="border-t border-border px-4 py-4">
              <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Price Options</div>
              <div class="border border-border rounded-sm divide-y divide-border">
                <div v-for="opt in detailProduct.price_options" :key="opt.id" class="flex items-center justify-between px-3 py-2">
                  <div>
                    <div class="text-sm font-medium">{{ opt.name || 'Base' }}</div>
                    <div v-if="opt.quantity_per_unit" class="text-xs text-muted-foreground">{{ opt.quantity_per_unit }} per unit</div>
                  </div>
                  <div class="text-sm font-semibold">&#8358;{{ (opt.selling_price / 100).toLocaleString() }}</div>
                </div>
              </div>
            </div>
          </template>

          <!-- Edit Mode -->
          <template v-else>
            <div class="px-4 py-4 space-y-4">
              <div>
                <label class="text-xs text-muted-foreground mb-1 block">Name</label>
                <input v-model="editForm.name" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" />
              </div>
              <div>
                <label class="text-xs text-muted-foreground mb-1 block">Manufacturer</label>
                <input v-model="editForm.manufacturer" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" />
              </div>
              <div>
                <label class="text-xs text-muted-foreground mb-1 block">Barcode / SKU</label>
                <input v-model="editForm.barcode" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" />
              </div>
              <div>
                <label class="text-xs text-muted-foreground mb-1 block">Category</label>
                <select v-model="editForm.category_id" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring">
                  <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
                </select>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="text-xs text-muted-foreground mb-1 block">Cost Price (&#8358;)</label>
                  <input v-model.number="editForm.cost_price" type="number" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" />
                </div>
                <div>
                  <label class="text-xs text-muted-foreground mb-1 block">Reorder Level</label>
                  <input v-model.number="editForm.reorder_level" type="number" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" />
                </div>
              </div>
            </div>
          </template>
        </div>

        <!-- Bottom Actions (edit mode) -->
        <div v-if="editMode" class="border-t border-border px-4 py-3 flex items-center gap-2">
          <Button variant="outline" class="flex-1" @click="cancelEdit">Cancel</Button>
          <Button class="flex-1" :disabled="saving" @click="saveProduct">
            <RotateCw v-if="saving" :size="14" class="animate-spin mr-2" />
            <Check v-else :size="14" class="mr-2" />
            Save
          </Button>
        </div>
      </div>
    </Transition>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast" class="fixed bottom-6 right-6 z-50 bg-foreground text-background px-4 py-2 rounded-lg shadow-lg text-sm font-medium">
        {{ toast }}
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from "vue";
import { Search, RotateCw, AlertCircle, Package, ChevronLeft, ChevronRight, X, Pencil, Check } from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const API = "";

const permissions = ref(window.__PERMISSIONS__ ?? {});
const canViewReorder = computed(() => permissions.value["reorderlevel:view"]);
const canEditInventory = computed(() => permissions.value["inventory:edit"]);

const allProducts = ref([]);
const categories = ref([]);
const loading = ref(false);
const error = ref(null);
const searchQuery = ref("");
const categoryFilter = ref("");
const currentPage = ref(1);
const perPage = ref(20);
const detailProduct = ref(null);
const detailLoading = ref(false);
const editMode = ref(false);
const saving = ref(false);
const toast = ref(null);

const editForm = ref({
  name: "",
  manufacturer: "",
  barcode: "",
  category_id: null,
  cost_price: 0,
  reorder_level: 0,
});

let toastTimer = null;
function showToast(msg) {
  toast.value = msg;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { toast.value = null; }, 2500);
}

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
  if (categoryFilter.value) {
    items = items.filter((p) => p.category_id === Number(categoryFilter.value));
  }
  return items;
});

const totalPages = computed(() => Math.max(1, Math.ceil(filteredProducts.value.length / perPage.value)));

const paginatedProducts = computed(() => {
  const start = (currentPage.value - 1) * perPage.value;
  return filteredProducts.value.slice(start, start + perPage.value);
});

const visiblePages = computed(() => {
  const total = totalPages.value;
  const cur = currentPage.value;
  if (total <= 5) return Array.from({ length: total }, (_, i) => i + 1);
  if (cur <= 3) return [1, 2, 3, 4, 5];
  if (cur >= total - 2) return [total - 4, total - 3, total - 2, total - 1, total];
  return [cur - 2, cur - 1, cur, cur + 1, cur + 2];
});

function formatDate(date) {
  if (!date) return "—";
  const d = new Date(date);
  return d.toLocaleDateString("en-US", { year: "numeric", month: "short", day: "numeric" });
}

function stockClass(stock, reorderLevel) {
  if (stock <= 0) return "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400";
  if (stock <= reorderLevel) return "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400";
  return "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400";
}

function categoryName(id) {
  return categories.value.find((c) => c.id === id)?.name || "—";
}

function goPage(p) {
  if (p < 1 || p > totalPages.value) return;
  currentPage.value = p;
}

function closeDetail() {
  detailProduct.value = null;
  editMode.value = false;
}

async function selectAndFetch(id) {
  detailProduct.value = null;
  editMode.value = false;
  detailLoading.value = true;
  try {
    const res = await fetch(`${API}/inventory/product/${id}`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    detailProduct.value = await res.json();
  } catch (e) {
    showToast("Failed to load product details");
  } finally {
    detailLoading.value = false;
  }
}

function cancelEdit() {
  editMode.value = false;
}

async function saveProduct() {
  if (!detailProduct.value) return;
  saving.value = true;
  try {
    const payload = {
      id: detailProduct.value.id,
      name: editForm.value.name,
      manufacturer: editForm.value.manufacturer,
      barcode: editForm.value.barcode,
      category_id: editForm.value.category_id,
      cost_price: editForm.value.cost_price,
      reorder_level: editForm.value.reorder_level,
      prices: [],
    };
    const res = await fetch(`${API}/inventory/product/${detailProduct.value.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    showToast("Product updated");
    editMode.value = false;
    // Refresh product list and detail
    await Promise.all([fetchProducts(), selectAndFetch(detailProduct.value.id)]);
  } catch (e) {
    showToast(e.message || "Failed to save");
  } finally {
    saving.value = false;
  }
}

async function fetchProducts() {
  loading.value = true;
  error.value = null;
  try {
    const res = await fetch(`${API}/inventory/item-list`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    const data = await res.json();
    allProducts.value = data.items || [];
    categories.value = data.categories || [];
  } catch (e) {
    error.value = e.message || "Failed to load";
  } finally {
    loading.value = false;
  }
}

watch([searchQuery, categoryFilter], () => { currentPage.value = 1; });

watch(detailProduct, (p) => {
  if (p) {
    editForm.value = {
      name: p.name || "",
      manufacturer: p.manufacturer || "",
      barcode: p.barcode || "",
      category_id: p.category_id || null,
      cost_price: p.cost_price || 0,
      reorder_level: p.reorder_level || 0,
    };
  }
});

onMounted(() => { fetchProducts(); });
</script>

<style scoped>
.slide-panel-enter-active,
.slide-panel-leave-active {
  transition: opacity 0.15s ease;
}
.slide-panel-enter-from,
.slide-panel-leave-to {
  opacity: 0;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>