<template>
  <div class="p-6 lg:p-8">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-foreground">Receive Items</h1>
        <p class="text-sm text-muted-foreground mt-1">Record incoming inventory from suppliers</p>
      </div>
      <Button variant="outline" size="sm" @click="newReceipt" :disabled="items.length > 0">
        <Plus :size="14" class="mr-1.5" />
        New Receipt
      </Button>
    </div>

    <!-- Supplier Card -->
    <div class="rounded-lg border border-border bg-card p-4 mb-6">
      <div class="flex items-center gap-3">
        <Building :size="18" class="text-muted-foreground shrink-0" />
        <div class="flex-1 relative">
          <input
            v-model="supplier"
            type="text"
            placeholder="Enter or search for supplier name..."
            class="w-full px-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
            @input="debouncedSupplierSearch"
          />
          <ul
            v-if="supplierSuggestions.length"
            class="absolute z-10 w-full mt-0.5 bg-popover border border-border rounded-lg shadow-lg overflow-hidden"
          >
            <li
              v-for="s in supplierSuggestions"
              :key="s"
              class="flex items-center gap-2 px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors"
              @click="selectSupplier(s)"
            >
              <Building :size="14" class="text-muted-foreground" />
              <span>{{ s }}</span>
            </li>
          </ul>
        </div>
        <Button v-if="supplier" variant="ghost" size="icon" class="h-9 w-9 text-muted-foreground hover:text-destructive" @click="supplier = ''">
          <X :size="14" />
        </Button>
      </div>
    </div>

    <!-- Add Items Section -->
    <div class="rounded-lg border border-border bg-card p-4 mb-6">
      <div class="flex items-center gap-2 mb-3">
        <div class="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">1</div>
        <span class="text-sm font-semibold text-foreground">Add Products</span>
        <span v-if="items.length" class="text-xs text-muted-foreground">· {{ items.length }} item{{ items.length !== 1 ? 's' : '' }} · &#8358;{{ subtotal.toLocaleString() }}</span>
      </div>
      <div class="relative">
        <Search :size="16" class="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <input
          v-model="productQuery"
          type="text"
          placeholder="Search products by name or barcode..."
          class="w-full pl-10 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
          @input="debouncedProductSearch"
        />
        <ul
          v-if="productSuggestions.length"
          class="absolute z-10 w-full mt-0.5 bg-popover border border-border rounded-lg shadow-lg overflow-y-auto max-h-56"
        >
          <li
            v-for="p in productSuggestions"
            :key="p.id"
            class="flex items-center justify-between px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors border-b border-border/50"
            :class="{ 'border-b-0': p === productSuggestions[productSuggestions.length - 1] }"
            @click="addProduct(p)"
          >
            <div class="flex items-center gap-2">
              <PillBottle :size="16" class="text-muted-foreground/60 shrink-0" />
              <div>
                <div class="font-medium text-foreground">{{ p.name }}</div>
                <div class="text-xs text-muted-foreground">{{ p.manufacturer || '—' }} · {{ p.barcode || 'no barcode' }}</div>
              </div>
            </div>
            <div class="text-xs text-muted-foreground font-mono">&#8358;{{ (p.default_price?.selling_price || 0).toLocaleString() }}</div>
          </li>
        </ul>
      </div>

      <!-- Empty state -->
      <div v-if="!items.length" class="flex flex-col items-center justify-center py-16 text-center">
        <div class="w-16 h-16 rounded-full bg-muted/30 flex items-center justify-center mb-4">
          <Package :size="32" class="text-muted-foreground/40" />
        </div>
        <h3 class="text-base font-semibold text-foreground mb-1">No items yet</h3>
        <p class="text-sm text-muted-foreground max-w-sm">Search and select products above to add them to this receipt. You can adjust prices, quantities, and expiry dates for each item.</p>
      </div>
    </div>

    <!-- Items Table -->
    <div v-if="items.length" class="rounded-lg border border-border overflow-hidden mb-6">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border bg-muted/30">
            <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-44">Item</th>
            <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-22">Barcode</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-18">Cost</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-20">Sell Price</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-14">Qty</th>
            <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-26">Expiry</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5 w-18">Total</th>
            <th class="w-10"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border/50">
          <tr v-for="(item, idx) in items" :key="item._key" class="hover:bg-muted/20 group">
            <td class="px-3 py-2.5">
              <div class="text-sm font-medium text-foreground">{{ item.name }}</div>
              <div class="text-[10px] text-muted-foreground">{{ item.manufacturer || '' }}</div>
            </td>
            <td class="px-3 py-2.5"><input :value="item.barcode" @input="item.barcode = $event.target.value" class="w-full px-2 py-1 text-[11px] border border-border rounded bg-background font-mono outline-none" /></td>
            <td class="px-3 py-2.5">
              <div class="relative">
                <span class="absolute left-2 top-1/2 -translate-y-1/2 text-[10px] text-muted-foreground">&#8358;</span>
                <input :value="item.cost_price" @input="item.cost_price = Number($event.target.value) || 0; suggestSellingPrice(item)" type="number" step="0.01" class="w-full pl-5 py-1 text-[11px] text-right border border-border rounded bg-background outline-none" />
              </div>
            </td>
            <td class="px-3 py-2.5"><input :value="item.selling_price" @input="item.selling_price = Number($event.target.value) || 0" type="number" step="0.01" class="w-full px-2 py-1 text-[11px] text-right border border-border rounded bg-background outline-none" /></td>
            <td class="px-3 py-2.5">
              <div class="inline-flex items-center border border-border rounded overflow-hidden">
                <button class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent text-xs transition-colors" @click="item.quantity = Math.max(0, (item.quantity || 0) - 1)">−</button>
                <input :value="item.quantity" @input="item.quantity = Math.max(0, Number($event.target.value) || 0)" class="h-7 w-10 text-center text-xs bg-transparent border-x border-border outline-none" />
                <button class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent text-xs transition-colors" @click="item.quantity = (item.quantity || 0) + 1">+</button>
              </div>
            </td>
            <td class="px-3 py-2.5"><input :value="item.expiry" @input="item.expiry = $event.target.value" type="date" class="w-full px-2 py-1 text-[11px] text-center border border-border rounded bg-background outline-none" /></td>
            <td class="px-3 py-2.5 text-right text-xs font-medium text-foreground">&#8358;{{ (Number(item.cost_price || 0) * Number(item.quantity || 0)).toLocaleString() }}</td>
            <td class="px-3 py-2.5"><Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-destructive" @click="removeItem(idx)"><X :size="13" /></Button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Actions Footer -->
    <div v-if="items.length" class="rounded-lg border border-border bg-card p-4">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-3">
        <div class="flex items-center gap-3 text-sm">
          <div class="flex items-center gap-1.5">
            <Package :size="16" class="text-muted-foreground" />
            <span class="font-medium">{{ items.length }}</span>
            <span class="text-muted-foreground">item{{ items.length !== 1 ? 's' : '' }}</span>
          </div>
          <span class="text-muted-foreground">·</span>
          <div class="flex items-center gap-1.5">
            <span class="font-semibold">&#8358;{{ subtotal.toLocaleString() }}</span>
            <span class="text-muted-foreground">total cost</span>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" :disabled="!canSubmit" @click="holdReceipt">
            <PauseCircle :size="14" class="mr-1.5" />
            Hold Draft
          </Button>
          <Button size="lg" class="flex-1 sm:flex-none px-6 gap-2" :disabled="!canSubmit" @click="receiveItems">
            <CircleCheck :size="16" />
            Receive Items
          </Button>
        </div>
      </div>
      <div v-if="!supplier" class="text-xs text-amber-600 dark:text-amber-400 mt-2 flex items-center gap-1.5">
        <AlertTriangle :size="12" />
        Enter a supplier name to enable submission
      </div>
    </div>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast" class="fixed bottom-6 right-6 z-50 flex items-center gap-2 bg-card border border-border px-4 py-3 rounded-lg shadow-xl text-sm font-medium">
        <component :is="toastIcon" :size="16" :class="toastColor" />
        {{ toast }}
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { Search, X, Building, Package, PillBottle, PauseCircle, CircleCheck, Plus, AlertTriangle, CircleCheck as CheckCircle } from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const API = "";

const supplier = ref("");
const productQuery = ref("");
const items = ref([]);
const submitting = ref(false);
const toast = ref(null);
const toastType = ref("info");
const supplierSuggestions = ref([]);
const productSuggestions = ref([]);
const localItems = ref([]);

const canSubmit = computed(() => supplier.value.trim() && items.value.length > 0);
const toastIcon = computed(() => toastType.value === "success" ? CheckCircle : AlertTriangle);
const toastColor = computed(() => toastType.value === "success" ? "text-emerald-600" : "text-amber-600");

const subtotal = computed(() => items.value.reduce((s, i) => s + Number(i.cost_price || 0) * Number(i.quantity || 0), 0));

let keyCounter = 0;
let tTimer = null;
let psTimer = null;
let ssTimer = null;

function showToast(msg, type = "info") {
  toast.value = msg; toastType.value = type;
  clearTimeout(tTimer);
  tTimer = setTimeout(() => { toast.value = null; }, 2800);
}

function suggestSellingPrice(item) {
  // Auto-suggest 30% markup on cost price if selling price is 0 or equal to cost
  if (!item.selling_price || Number(item.selling_price) <= Number(item.cost_price || 0)) {
    item.selling_price = Math.round(Number(item.cost_price || 0) * 1.3 * 100) / 100;
  }
}

function newReceipt() { items.value.splice(0); supplier.value = ""; }
function selectSupplier(s) { supplier.value = s; supplierSuggestions.value = []; }
function removeItem(idx) { items.value.splice(idx, 1); }

function debouncedSupplierSearch() {
  const q = supplier.value.trim();
  if (q.length < 2) { supplierSuggestions.value = []; return; }
  clearTimeout(ssTimer);
  ssTimer = setTimeout(async () => {
    try {
      const r = await fetch(`${API}/inventory/suppliers/search?query=${encodeURIComponent(q)}`);
      if (r.ok) supplierSuggestions.value = await r.json();
    } catch {}
  }, 200);
}

function debouncedProductSearch() {
  const q = productQuery.value.trim();
  if (q.length < 2) { productSuggestions.value = []; return; }
  clearTimeout(psTimer);
  psTimer = setTimeout(async () => {
    try {
      const r = await fetch(`${API}/inventory/search?query=${encodeURIComponent(q)}`);
      if (r.ok) productSuggestions.value = await r.json();
    } catch {}
  }, 250);
}

function addProduct(p) {
  productQuery.value = ""; productSuggestions.value = []; keyCounter++;
  const cost = (p.cost_price || 0) / 100;
  const sell = p.default_price?.selling_price || Math.round(cost * 1.3 * 100) / 100;
  items.value.push({
    _key: p.id + "-" + keyCounter, id: p.id,
    name: p.name, manufacturer: p.manufacturer || "",
    barcode: p.barcode || "", cost_price: cost, selling_price: sell,
    quantity: 1, expiry: "",
  });
}

function getId() { try { return crypto.randomUUID(); } catch { return Date.now() + "-" + Math.random().toString(36).slice(2); } }

async function holdReceipt() {
  submitting.value = true;
  try {
    const p = { reference: "", payload: JSON.stringify({ supplier: supplier.value.trim(), products: items.value.map(i => ({ id: i.id, barcode: i.barcode, cost_price: i.cost_price, selling_price: i.selling_price, quantity: i.quantity, expiry: i.expiry || null, price_options_changes: [] })) }) };
    const r = await fetch(`${API}/inventory/receive-items/hold`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(p) });
    if (!r.ok) throw new Error("Failed to hold");
    newReceipt(); showToast("Receipt saved as draft", "success");
  } catch (e) { showToast(e.message || "Failed to hold"); }
  finally { submitting.value = false; }
}

async function receiveItems() {
  submitting.value = true;
  try {
    const p = { supplier: supplier.value.trim(), products: items.value.map(i => ({ id: i.id, barcode: i.barcode, cost_price: i.cost_price, selling_price: i.selling_price, quantity: i.quantity, expiry: i.expiry || null, price_options_changes: [] })), idempotency_key: getId() };
    const r = await fetch(`${API}/inventory/receive-items`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(p) });
    if (!r.ok) { const e = await r.json().catch(() => ({})); throw new Error(e.error || `HTTP ${r.status}`); }
    newReceipt(); showToast("Items received successfully", "success");
  } catch (e) { showToast(e.message || "Failed to receive"); }
  finally { submitting.value = false; }
}
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>