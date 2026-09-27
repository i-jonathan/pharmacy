<template>
  <div class="p-6 lg:p-8">
    <!-- ===== DASHBOARD VIEW ===== -->
    <template v-if="view === 'dashboard'">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h1 class="text-2xl font-bold text-foreground">Receive Items</h1>
          <p class="text-sm text-muted-foreground mt-1">Record incoming inventory from suppliers</p>
        </div>
        <Button size="lg" @click="startNewReceipt">
          <Plus :size="16" class="mr-2" />
          New Receipt
        </Button>
      </div>

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <div class="rounded-lg border border-border bg-card p-5">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-sky-100 dark:bg-sky-900/20 flex items-center justify-center">
              <CalendarCheck :size="20" class="text-sky-600" />
            </div>
            <div>
              <div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Received Today</div>
              <div class="text-2xl font-bold text-foreground">{{ todayCount }}</div>
            </div>
          </div>
        </div>
        <div class="rounded-lg border border-border bg-card p-5">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-900/20 flex items-center justify-center">
              <PauseCircle :size="20" class="text-amber-600" />
            </div>
            <div>
              <div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Held Drafts</div>
              <div class="text-2xl font-bold text-foreground">{{ heldDrafts.length }}</div>
            </div>
          </div>
        </div>
        <div class="rounded-lg border border-border bg-card p-5">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-emerald-100 dark:bg-emerald-900/20 flex items-center justify-center">
              <Package :size="20" class="text-emerald-600" />
            </div>
            <div>
              <div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">This Month</div>
              <div class="text-2xl font-bold text-foreground">{{ monthCount }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Receipts -->
      <div class="border border-border rounded-lg overflow-hidden mb-6">
        <div class="flex items-center justify-between px-4 py-3 bg-muted/30 border-b border-border">
          <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Recent Receipts</span>
          <router-link to="/received-items-history" class="text-xs text-primary hover:underline">View all</router-link>
        </div>
        <div v-if="recentReceipts.length" class="divide-y divide-border/50">
          <div v-for="r in recentReceipts" :key="r.id" class="flex items-center justify-between px-4 py-3 hover:bg-muted/20 transition-colors">
            <div class="flex items-center gap-2">
              <Truck :size="14" class="text-muted-foreground shrink-0" />
              <span class="text-sm font-medium text-foreground">{{ r.supplier_name }}</span>
            </div>
            <div class="text-xs text-muted-foreground">{{ r.created_at ? new Date(r.created_at).toLocaleDateString() : '' }} · {{ r.items?.length || 0 }} items</div>
          </div>
        </div>
        <div v-if="!hasReceipts" class="px-4 py-6 text-center text-sm text-muted-foreground">No receipts yet. Create your first one.</div>
      </div>
    </template>

    <!-- ===== NEW RECEIPT VIEW ===== -->
    <template v-if="view === 'receipt'">
      <!-- Header with back link -->
      <div class="flex items-center gap-3 mb-6">
        <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground" @click="view = 'dashboard'">
          <ChevronLeft :size="16" />
        </Button>
        <div>
          <h1 class="text-xl font-bold text-foreground">New Receipt</h1>
          <p class="text-sm text-muted-foreground">Record incoming items from a supplier</p>
        </div>
      </div>

      <!-- Step 1: Supplier -->
      <div class="rounded-lg border border-border bg-card p-4 mb-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">1</div>
          <span class="text-sm font-semibold text-foreground">Supplier</span>
          <span v-if="supplier" class="text-xs text-emerald-600 ml-2">✓ {{ supplier }}</span>
        </div>
        <div class="relative">
          <Building :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="supplier"
            type="text"
            placeholder="Enter supplier name..."
            class="w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
            :class="{ 'border-red-500 dark:border-red-400': validationErrors.supplier }"
            @input="onSupplierInput"
          />
        </div>
        <ul v-if="supplierSuggestions.length" class="mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-hidden">
          <li v-for="s in supplierSuggestions" :key="s" class="flex items-center gap-2 px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors" @click="selectSupplier(s)"><Building :size="14" class="text-muted-foreground" /><span>{{ s }}</span></li>
        </ul>
      </div>

      <!-- Step 2: Add Products -->
      <div class="rounded-lg border border-border bg-card p-4 mb-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">2</div>
          <span class="text-sm font-semibold text-foreground">Products</span>
          <span v-if="items.length" class="text-xs text-muted-foreground ml-2">· {{ items.length }} item{{ items.length !== 1 ? 's' : '' }}</span>
        </div>

        <div class="relative">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="productQuery"
            type="text"
            placeholder="Search products by name or barcode..."
            class="w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
            @input="onProductSearch"
          />
        </div>

        <!-- Product search results -->
        <ul v-if="productSuggestions.length" class="mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-y-auto max-h-56">
          <li v-for="p in productSuggestions" :key="p.id" class="flex items-center justify-between px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors border-b border-border/50" @click="addProduct(p)">
            <div class="flex items-center gap-2.5">
              <PillBottle :size="16" class="text-muted-foreground/60 shrink-0" />
              <div>
                <div class="font-medium text-foreground">{{ p.name }}</div>
                <div class="text-xs text-muted-foreground">{{ p.manufacturer || '—' }} · {{ p.barcode || 'no barcode' }}</div>
              </div>
            </div>
            <div class="text-xs text-muted-foreground font-mono">&#8358;{{ (p.default_price?.selling_price || 0).toLocaleString() }}</div>
          </li>
          <li class="border-t border-border">
            <div class="px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors text-primary font-medium" @click="openNewProductModal">
              <Plus :size="14" class="inline mr-1.5" />
              Create new product...
            </div>
          </li>
        </ul>

        <!-- New Product Modal -->
        <div v-if="showNewProductModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="showNewProductModal = false">
          <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
            <h3 class="text-base font-semibold mb-4">New Product</h3>
            <div class="space-y-3">
              <div><label class="text-xs text-muted-foreground mb-1 block">Product Name <span class="text-destructive">*</span></label><input v-model="newProduct.name" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              <div><label class="text-xs text-muted-foreground mb-1 block">Manufacturer</label><input v-model="newProduct.manufacturer" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              <div><label class="text-xs text-muted-foreground mb-1 block">Barcode</label><input v-model="newProduct.barcode" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              <div class="grid grid-cols-2 gap-3">
                <div><label class="text-xs text-muted-foreground mb-1 block">Selling Price <span class="text-destructive">*</span></label><input v-model.number="newProduct.selling_price" type="number" step="0.01" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
                <div><label class="text-xs text-muted-foreground mb-1 block">Cost Price <span class="text-destructive">*</span></label><input v-model.number="newProduct.cost_price" type="number" step="0.01" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              </div>
            </div>
            <div class="flex items-center gap-2 mt-5 justify-end">
              <Button variant="outline" size="sm" @click="showNewProductModal = false">Cancel</Button>
              <Button size="sm" :disabled="newProductSaving" @click="saveNewProduct">
                <RotateCw v-if="newProductSaving" :size="14" class="animate-spin mr-1.5" />
                <Plus v-else :size="14" class="mr-1.5" />
                Create & Add
              </Button>
            </div>
          </div>
        </div>

        <!-- Empty state -->
        <div v-if="!items.length" class="flex flex-col items-center justify-center py-12 text-center mt-2">
          <Package :size="32" class="text-muted-foreground/30 mb-3" />
          <p class="text-sm text-muted-foreground">Search and select products above to add them to this receipt.</p>
        </div>
      </div>

      <!-- Step 3: Items Table -->
      <div v-if="items.length" class="mb-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">3</div>
          <span class="text-sm font-semibold text-foreground">Review & Set Details</span>
        </div>
        <div class="rounded-lg border border-border overflow-hidden">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30 text-[11px]">
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2">Item</th>
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2" style="width: 13%">Cost</th>
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2" style="width: 13%">Sell Price</th>
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2" style="width: 10%">Qty</th>
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2" style="width: 16%">Expiry</th>
                <th class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2" style="width: 12%">Total</th>
                <th style="width: 4%"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr v-for="(item, idx) in items" :key="item._key" class="hover:bg-muted/20 transition-colors">
                <td class="px-2.5 py-2.5">
                  <div class="text-sm font-medium text-foreground">{{ item.name }}</div>
                  <div class="text-[11px] text-muted-foreground">{{ item.manufacturer || '' }}</div>
                </td>
                <td class="px-2.5 py-2.5">
                  <div class="relative">
                    <span class="text-[11px] text-muted-foreground absolute left-1.5 top-1.5">&#8358;</span>
                    <input :value="item.cost_price" @input="item.cost_price = num($event.target.value); clearRowError(item); suggestPrice(item)" type="number" step="0.01" min="0" class="w-full pl-4.5 py-1.5 text-right text-sm border border-border rounded-md bg-background outline-none focus:ring-1 focus:ring-ring font-medium" :class="{ 'border-red-500': item._errors?.cost }" />
                  </div>
                </td>
                <td class="px-2.5 py-2.5">
                  <div class="relative">
                    <span class="text-[11px] text-muted-foreground absolute left-1.5 top-1.5">&#8358;</span>
                    <input :value="item.selling_price" @input="item.selling_price = num($event.target.value); clearRowError(item)" type="number" step="0.01" min="0" class="w-full pl-4.5 py-1.5 text-right text-sm border border-border rounded-md bg-background outline-none focus:ring-1 focus:ring-ring font-medium" :class="{ 'border-red-500': item._errors?.sell }" />
                  </div>
                </td>
                <td class="px-2.5 py-2.5">
                  <div class="inline-flex items-center border border-border rounded-md overflow-hidden">
                    <button class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent text-sm transition-colors" @click="item.quantity = Math.max(0, (item.quantity || 0) - 1); clearRowError(item)">−</button>
                    <input :value="item.quantity" @input="item.quantity = Math.max(0, num($event.target.value)); clearRowError(item)" class="h-7 w-9 text-center text-sm bg-transparent border-x border-border outline-none" :class="{ 'text-destructive': !item.quantity }" />
                    <button class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent text-sm transition-colors" @click="item.quantity = (item.quantity || 0) + 1; clearRowError(item)">+</button>
                  </div>
                </td>
                <td class="px-2.5 py-2.5">
                  <input :value="item.expiry" @input="item.expiry = $event.target.value; clearRowError(item)" type="date" class="w-full px-2 py-1.5 text-sm text-center border border-border rounded-md bg-background outline-none focus:ring-1 focus:ring-ring" :class="{ 'border-red-500': item._errors?.expiry }" />
                </td>
                <td class="px-2.5 py-2.5 text-right font-semibold text-sm text-foreground">&#8358;{{ num(item.cost_price || 0) * num(item.quantity || 0) }}</td>
                <td class="px-2.5 py-2.5"><Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-destructive" @click="removeItem(idx)"><X :size="13" /></Button></td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Submit -->
      <div v-if="items.length" class="border border-border bg-card rounded-lg p-4">
        <div class="flex flex-col sm:flex-row items-start justify-between gap-3">
          <div class="flex items-center gap-3 text-sm">
            <span class="font-semibold">{{ items.length }}</span>
            <span class="text-muted-foreground">items</span>
            <span class="text-muted-foreground">·</span>
            <span class="font-semibold">&#8358;{{ items.reduce((s,i) => s + num(i.cost_price||0) * num(i.quantity||0), 0).toLocaleString() }}</span>
            <span class="text-muted-foreground">total cost</span>
          </div>
          <div class="flex items-center gap-2 mt-2">
            <Button variant="outline" size="sm" :disabled="submitting" @click="holdReceipt"><PauseCircle :size="14" class="mr-1.5" />Hold Draft</Button>
            <Button size="lg" class="px-6 gap-2" :disabled="submitting" @click="receiveItems"><CircleCheck :size="16" />Receive Items</Button>
          </div>
        </div>
        <div v-if="validationErrors.global" class="text-xs text-destructive mt-1.5 flex items-center gap-1">
          <AlertTriangle :size="12" />{{ validationErrors.global }}
        </div>
      </div>
    </template>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast" class="fixed bottom-6 right-6 z-50 flex items-center gap-2.5 bg-card border border-border px-4 py-3 rounded-xl shadow-2xl text-sm font-medium">
        <CircleCheck v-if="toastType === 'success'" :size="16" class="text-emerald-600" />
        <AlertCircle v-else :size="16" class="text-destructive" />
        {{ toast }}
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { Search, X, Building, Package, PillBottle, PauseCircle, CircleCheck, Plus, AlertTriangle, AlertCircle, ChevronLeft, Trash2, CalendarCheck, Truck, RotateCw } from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const API = "";

const hasReceipts = computed(() => recentReceipts.value.length > 0);
const VIEW_DASHBOARD = "dashboard";
const VIEW_RECEIPT = "receipt";

// View state
const view = ref(VIEW_DASHBOARD);

// Supplier
const supplier = ref("");
const supplierSuggestions = ref([]);

// Product search
const productQuery = ref("");
const productSuggestions = ref([]);

// New product modal
const showNewProductModal = ref(false);
const newProductSaving = ref(false);
const newProduct = ref({ name: "", manufacturer: "", barcode: "", selling_price: 0, cost_price: 0 });

// Items
const items = ref([]);

// Dashboard data
const recentReceipts = ref([]);
const heldDrafts = ref([]);
const todayCount = ref(0);
const monthCount = ref(0);

// Submission
const submitting = ref(false);

// Validation
const validationErrors = ref({ supplier: false, global: "" });

// Toast
const toast = ref(null);
const toastType = ref("success");

let tTimer = null;
let keyCounter = 0;

function num(v) { return Number(v || 0); }

function showToast(msg, type = "success") {
  toast.value = msg; toastType.value = type;
  clearTimeout(tTimer);
  tTimer = setTimeout(() => { toast.value = null; }, 3000);
}

function clearRowError(item) {
  if (item._errors) item._errors = {};
}

function suggestPrice(item) {
  if (num(item.cost_price) > 0 && (!num(item.selling_price) || num(item.selling_price) <= num(item.cost_price))) {
    item.selling_price = Math.round(num(item.cost_price) * 1.3 * 100) / 100;
  }
}

// === Dashboard ===
async function fetchDashboard() {
  try {
    const r = await fetch(`${API}/inventory/received-items-history/api`);
    if (r.ok) {
      const d = await r.json();
      recentReceipts.value = (d.batches || []).slice(0, 5);
    }
  } catch {}
  todayCount.value = recentReceipts.value.length;
}

function onSupplierInput() {
  validationErrors.value.supplier = false;
  debouncedSupplierSearch();
}

function onProductSearch() {
  debouncedProductSearch();
}

function getHeldSupplier(h) {
  try {
    const p = typeof h.payload === "string" ? JSON.parse(h.payload) : h.payload;
    return p.supplier || "—";
  } catch { return "—"; }
}
function getHeldItemCount(h) {
  try {
    const p = typeof h.payload === "string" ? JSON.parse(h.payload) : h.payload;
    return p.products?.length || 0;
  } catch { return 0; }
}
function formatDate(d) {
  if (!d) return "";
  try { return new Date(d).toLocaleDateString(); } catch { return ""; }
}
async function voidHeldDraft(ref) {
  if (!confirm("Delete this draft?")) return;
  try {
    await fetch(`${API}/inventory/receive-items/held/${encodeURIComponent(ref)}`, { method: "DELETE" });
    heldDrafts.value = heldDrafts.value.filter((h) => h.reference !== ref);
  } catch {}
}

// === Navigation ===
function startNewReceipt() {
  view.value = VIEW_RECEIPT;
  supplier.value = "";
  items.value.splice(0);
  productQuery.value = "";
  validationErrors.value = { supplier: false, global: "" };
}

// === Supplier ===
function selectSupplier(s) { supplier.value = s; supplierSuggestions.value = []; validationErrors.value.supplier = false; }

let ssTimer = null;
function debouncedSupplierSearch() {
  const q = supplier.value.trim();
  if (q.length < 2) { supplierSuggestions.value = []; return; }
  clearTimeout(ssTimer);
  ssTimer = setTimeout(async () => {
    try {
      const r = await fetch(`${API}/inventory/suppliers/search?query=${encodeURIComponent(q)}`);
      if (r.ok) { const d = await r.json(); supplierSuggestions.value = d.data || []; }
    } catch {}
  }, 200);
}

// === Product Search ===
let psTimer = null;
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

// === Add Product ===
function addProduct(p) {
  productQuery.value = ""; productSuggestions.value = []; keyCounter++;
  const cost = p.cost_price || (p.default_price?.selling_price ? p.default_price.selling_price * 0.7 : 0);
  const sell = p.default_price?.selling_price || Math.round(cost * 1.3 * 100) / 100;
  items.value.push({
    _key: p.id + "-" + keyCounter, id: p.id,
    name: p.name, manufacturer: p.manufacturer || "", barcode: p.barcode || "",
    cost_price: cost, selling_price: sell, quantity: 1, expiry: "", _errors: {},
  });
}

function removeItem(idx) { items.value.splice(idx, 1); }

// === New Product Modal ===
function openNewProductModal() {
  productSuggestions.value = [];
  showNewProductModal.value = true;
  newProduct.value = { name: "", manufacturer: "", barcode: "", selling_price: 0, cost_price: 0 };
}
async function saveNewProduct() {
  if (!newProduct.value.name.trim()) { showToast("Product name is required", "error"); return; }
  if (num(newProduct.value.selling_price) <= 0) { showToast("Selling price is required", "error"); return; }
  newProductSaving.value = true;
  try {
    const r = await fetch(`${API}/inventory/add-item`, {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: newProduct.value.name.trim(),
        manufacturer: newProduct.value.manufacturer.trim(),
        barcode: newProduct.value.barcode.trim(),
        category_id: 1,
        reorder_level: 5,
        cost_price: num(newProduct.value.cost_price),
        selling_price: num(newProduct.value.selling_price),
      }),
    });
    if (!r.ok) throw new Error("Failed to create product");
    const created = await r.json();
    showNewProductModal.value = false;
    addProduct({ id: created.id, name: newProduct.value.name.trim(), manufacturer: newProduct.value.manufacturer.trim(), barcode: newProduct.value.barcode.trim(), cost_price: num(newProduct.value.cost_price), default_price: { selling_price: num(newProduct.value.selling_price) } });
    showToast("Product created and added");
  } catch (e) {
    showToast(e.message || "Failed to create product", "error");
  } finally {
    newProductSaving.value = false;
  }
}

// === Validation ===
function validate() {
  let valid = true;
  validationErrors.value = { supplier: false, global: "" };

  if (!supplier.value.trim()) {
    validationErrors.value.supplier = true;
    valid = false;
  }
  if (!items.value.length) {
    validationErrors.value.global = "Add at least one product.";
    valid = false;
  }
  for (const item of items.value) {
    item._errors = {};
    if (num(item.cost_price) <= 0) { item._errors.cost = true; valid = false; }
    if (num(item.selling_price) <= 0) { item._errors.sell = true; valid = false; }
    if (num(item.quantity) <= 0) { item._errors.qty = true; valid = false; }
    if (!item.expiry) { item._errors.expiry = true; valid = false; }
  }
  if (!valid && !validationErrors.value.global) {
    valiDationErrors.value.global = "Fill in all required fields marked in red.";
  }
  return valid;
}

async function holdReceipt() {
  if (!validate()) return;
  submitting.value = true;
  try {
    const p = { reference: "", payload: JSON.stringify({ supplier: supplier.value.trim(), products: items.value.map(i => ({ id: i.id, barcode: i.barcode, cost_price: i.cost_price, selling_price: i.selling_price, quantity: i.quantity, expiry: i.expiry || null, price_options_changes: [] })) }) };
    const r = await fetch(`${API}/inventory/receive-items/hold`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(p) });
    if (!r.ok) throw new Error("Failed to hold");
    view.value = VIEW_DASHBOARD;
    showToast("Receipt saved as draft");
    fetchDashboard();
  } catch (e) { showToast(e.message || "Failed to hold", "error"); }
  finally { submitting.value = false; }
}

async function receiveItems() {
  if (!validate()) return;
  submitting.value = true;
  try {
    const p = {
      supplier: supplier.value.trim(),
      products: items.value.map(i => ({ id: i.id, barcode: i.barcode, cost_price: i.cost_price, selling_price: i.selling_price, quantity: i.quantity, expiry: i.expiry || null, price_options_changes: [] })),
      idempotency_key: (() => { try { return crypto.randomUUID(); } catch { return Date.now() + "-" + Math.random().toString(36).slice(2); } })(),
    };
    const r = await fetch(`${API}/inventory/receive-items`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(p) });
    if (!r.ok) { const e = await r.json().catch(() => ({})); throw new Error(e.error || `HTTP ${r.status}`); }
    view.value = VIEW_DASHBOARD;
    showToast("Items received successfully");
    fetchDashboard();
  } catch (e) { showToast(e.message || "Failed to receive", "error"); }
  finally { submitting.value = false; }
}

onMounted(() => { fetchDashboard(); });
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>