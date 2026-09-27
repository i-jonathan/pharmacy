<template>
  <div class="p-6">
    <!-- Header -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Receive Items</h1>
      <p class="text-sm text-muted-foreground mt-1">Record incoming inventory from suppliers</p>
    </div>

    <!-- Supplier -->
    <div class="flex items-end gap-3 mb-4 max-w-lg">
      <div class="relative flex-1">
        <label class="text-xs text-muted-foreground mb-1 block">Supplier</label>
        <input
          v-model="supplier"
          type="text"
          placeholder="Enter supplier name..."
          class="w-full px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
          @input="searchSuppliers"
        />
        <ul
          v-if="supplierSuggestions.length"
          class="absolute z-10 w-full mt-1 bg-popover border border-border rounded-md shadow-lg"
        >
          <li
            v-for="s in supplierSuggestions"
            :key="s"
            class="px-3 py-2 text-sm cursor-pointer hover:bg-accent transition-colors"
            @click="selectSupplier(s)"
          >{{ s }}</li>
        </ul>
      </div>
      <Button variant="outline" size="sm" v-if="supplier" @click="supplier = ''">
        <X :size="12" class="mr-1" />
        Clear
      </Button>
    </div>

    <!-- Product Search -->
    <div class="flex items-center gap-3 mb-6">
      <div class="relative flex-1">
        <Search :size="16" class="absolute left-3 top-3 text-muted-foreground" />
        <input
          v-model="productQuery"
          type="text"
          placeholder="Search product by name or barcode..."
          class="w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
          @input="searchProducts"
        />
        <ul
          v-if="productSuggestions.length"
          class="absolute z-10 w-full mt-1 bg-popover border border-border rounded-md shadow-lg max-h-48 overflow-y-auto"
        >
          <li
            v-for="p in productSuggestions"
            :key="p.id"
            class="flex items-center justify-between px-3 py-2 text-sm cursor-pointer hover:bg-accent transition-colors"
            @click="addProduct(p)"
          >
            <div>
              <div class="font-medium">{{ p.name }}</div>
              <div class="text-xs text-muted-foreground">{{ p.manufacturer || '' }}</div>
            </div>
            <div class="text-xs text-muted-foreground">&#8358;{{ (p.default_price?.selling_price || 0).toLocaleString() }}</div>
          </li>
        </ul>
      </div>
    </div>

    <!-- Items Table -->
    <div v-if="items.length" class="rounded-lg border border-border bg-card overflow-x-auto mb-6">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border bg-muted/30">
            <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-44">Item</th>
            <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-24">Barcode</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-20">Cost (&#8358;)</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-20">Sell Price</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-16">Qty</th>
            <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-28">Expiry</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-16 hidden">Subtotal</th>
            <th class="w-10"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border/50">
          <tr v-for="(item, idx) in items" :key="item._key" class="hover:bg-muted/20">
            <td class="px-3 py-2">
              <div class="text-sm font-medium">{{ item.name }}</div>
              <div class="text-xs text-muted-foreground">{{ item.manufacturer || '' }}</div>
            </td>
            <td class="px-3 py-2"><input :value="item.barcode" @input="item.barcode = $event.target.value" class="w-full px-2 py-1 text-xs border border-border rounded bg-background outline-none focus:ring-1 focus:ring-ring" /></td>
            <td class="px-3 py-2"><input :value="item.cost_price" @input="item.cost_price = Number($event.target.value) || 0" type="number" step="0.01" class="w-full px-2 py-1 text-xs text-right border border-border rounded bg-background outline-none focus:ring-1 focus:ring-ring" /></td>
            <td class="px-3 py-2"><input :value="item.selling_price" @input="item.selling_price = Number($event.target.value) || 0" type="number" step="0.01" class="w-full px-2 py-1 text-xs text-right border border-border rounded bg-background outline-none focus:ring-1 focus:ring-ring" /></td>
            <td class="px-3 py-2"><input :value="item.quantity" @input="item.quantity = Number($event.target.value) || 0" type="number" class="w-full px-2 py-1 text-xs text-right border border-border rounded bg-background outline-none focus:ring-1 focus:ring-ring" /></td>
            <td class="px-3 py-2"><input :value="item.expiry" @input="item.expiry = $event.target.value" type="date" class="w-full px-2 py-1 text-xs border border-border rounded bg-background outline-none focus:ring-1 focus:ring-ring" /></td>
            <td class="px-3 py-2 text-right text-xs font-medium">&#8358;{{ (item.cost_price * item.quantity).toLocaleString() }}</td>
            <td class="px-3 py-2"><Button variant="ghost" size="icon" class="h-6 w-6 text-muted-foreground hover:text-destructive" @click="items.splice(idx, 1)"><X :size="12" /></Button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Empty state -->
    <div v-if="!items.length" class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
      <Truck :size="48" class="text-muted-foreground/40 mb-4" />
      <h3 class="text-lg font-semibold text-foreground mb-1">No items added</h3>
      <p class="text-sm text-muted-foreground max-w-sm">Search for products above and add them to the receiving list.</p>
    </div>

    <!-- Summary + Actions -->
    <div v-if="items.length" class="flex items-center justify-between gap-4 border-t border-border pt-4">
      <div class="flex items-center gap-4">
        <span class="text-sm text-muted-foreground">{{ items.length }} item{{ items.length !== 1 ? 's' : '' }}</span>
        <span class="font-semibold text-sm">Total: &#8358;{{ subtotal.toLocaleString() }}</span>
      </div>
      <div class="flex items-center gap-2">
        <Button variant="outline" :disabled="!supplier || submitting" @click="holdReceipt">
          <PauseCircle :size="14" class="mr-1.5" />
          Hold
        </Button>
        <Button :disabled="!supplier || submitting" @click="receiveItems">
          <CircleCheck :size="14" class="mr-1.5" />
          Receive Items
        </Button>
      </div>
    </div>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast" class="fixed bottom-6 right-6 z-50 bg-foreground text-background px-4 py-2 rounded-lg shadow-lg text-sm font-medium">{{ toast }}</div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { Search, X, Truck, PauseCircle, CircleCheck, Plus } from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const API = "";

const supplier = ref("");
const productQuery = ref("");
const items = ref([]);
const submitting = ref(false);
const toast = ref(null);
const supplierSuggestions = ref([]);
const productSuggestions = ref([]);

let keyCounter = 0;
let toastTimer = null;
let productSearchTimer = null;
let supplierSearchTimer = null;

function showToast(msg) {
  toast.value = msg;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { toast.value = null; }, 2500);
}

const subtotal = computed(() => {
  return items.value.reduce((sum, item) => sum + Number(item.cost_price || 0) * Number(item.quantity || 0), 0);
});

function selectSupplier(s) {
  supplier.value = s;
  supplierSuggestions.value = [];
}

async function searchSuppliers() {
  const q = supplier.value.trim();
  if (q.length < 2) { supplierSuggestions.value = []; return; }
  clearTimeout(supplierSearchTimer);
  supplierSearchTimer = setTimeout(async () => {
    try {
      const res = await fetch(`${API}/inventory/suppliers/search?query=${encodeURIComponent(q)}`);
      if (res.ok) supplierSuggestions.value = await res.json();
    } catch {}
  }, 200);
}

async function searchProducts() {
  const q = productQuery.value.trim();
  if (q.length < 2) { productSuggestions.value = []; return; }
  clearTimeout(productSearchTimer);
  productSearchTimer = setTimeout(async () => {
    try {
      const res = await fetch(`${API}/inventory/search?query=${encodeURIComponent(q)}`);
      if (res.ok) productSuggestions.value = await res.json();
    } catch {}
  }, 200);
}

function addProduct(p) {
  productQuery.value = "";
  productSuggestions.value = [];
  keyCounter++;
  items.value.push({
    _key: p.id + "-" + keyCounter,
    id: p.id,
    name: p.name,
    manufacturer: p.manufacturer || "",
    barcode: p.barcode || "",
    cost_price: (p.cost_price || 0) / 100,
    selling_price: p.default_price?.selling_price || 0,
    quantity: 1,
    expiry: "",
  });
}

function getIdempotencyKey() {
  try { return crypto.randomUUID(); } catch { return Date.now() + "-" + Math.random().toString(36).slice(2); }
}

async function holdReceipt() {
  if (!supplier.value.trim()) { showToast("Enter a supplier name"); return; }
  if (!items.value.length) { showToast("Add at least one item"); return; }
  submitting.value = true;
  try {
    const payload = {
      reference: "",
      payload: JSON.stringify({
        supplier: supplier.value.trim(),
        products: items.value.map((item) => ({
          id: item.id,
          barcode: item.barcode,
          cost_price: item.cost_price,
          selling_price: item.selling_price,
          quantity: item.quantity,
          expiry: item.expiry || null,
          price_options_changes: [],
        })),
      }),
    };
    const res = await fetch(`${API}/inventory/receive-items/hold`, {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) throw new Error("Failed to hold receipt");
    items.value.splice(0);
    supplier.value = "";
    showToast("Receipt saved as draft");
  } catch (e) {
    showToast(e.message || "Failed to hold");
  } finally {
    submitting.value = false;
  }
}

async function receiveItems() {
  if (!supplier.value.trim()) { showToast("Enter a supplier name"); return; }
  if (!items.value.length) { showToast("Add at least one item"); return; }
  submitting.value = true;
  try {
    const payload = {
      supplier: supplier.value.trim(),
      products: items.value.map((item) => ({
        id: item.id,
        barcode: item.barcode,
        cost_price: item.cost_price,
        selling_price: item.selling_price,
        quantity: item.quantity,
        expiry: item.expiry || null,
        price_options_changes: [],
      })),
      idempotency_key: getIdempotencyKey(),
    };
    const res = await fetch(`${API}/inventory/receive-items`, {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    items.value.splice(0);
    supplier.value = "";
    showToast("Items received successfully");
  } catch (e) {
    showToast(e.message || "Failed to receive");
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>