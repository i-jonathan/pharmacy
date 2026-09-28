<template>
  <div class="p-6 lg:p-8">
    <!-- ===== DASHBOARD ===== -->
    <template v-if="view === 'dashboard'">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h1 class="text-2xl font-bold text-foreground">Receive Items</h1>
          <p class="text-sm text-muted-foreground mt-1">Record incoming inventory from suppliers</p>
        </div>
        <Button size="lg" @click="startNewReceipt"><Plus :size="16" class="mr-2" />New Receipt</Button>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <div class="rounded-lg border border-border bg-card p-5"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-sky-100 dark:bg-sky-900/20 flex items-center justify-center"><CalendarCheck :size="20" class="text-sky-600" /></div><div><div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Received Today</div><div class="text-2xl font-bold text-foreground">{{ todayCount }}</div></div></div></div>
        <div class="rounded-lg border border-border bg-card p-5"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-900/20 flex items-center justify-center"><PauseCircle :size="20" class="text-amber-600" /></div><div><div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Held Drafts</div><div class="text-2xl font-bold text-foreground">—</div></div></div></div>
        <div class="rounded-lg border border-border bg-card p-5"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-emerald-100 dark:bg-emerald-900/20 flex items-center justify-center"><Package :size="20" class="text-emerald-600" /></div><div><div class="text-xs text-muted-foreground uppercase tracking-wider font-semibold">This Month</div><div class="text-2xl font-bold text-foreground">{{ monthCount }}</div></div></div></div>
      </div>
      <div class="border border-border rounded-lg overflow-hidden mb-6">
        <div class="flex items-center justify-between px-4 py-3 bg-muted/30 border-b border-border"><span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Recent Receipts</span><router-link to="/received-items-history" class="text-xs text-primary hover:underline">View all</router-link></div>
        <div v-if="recentReceipts.length" class="divide-y divide-border/50">
          <div v-for="r in recentReceipts" :key="r.id" class="flex items-center justify-between px-4 py-3 hover:bg-muted/20 transition-colors"><div class="flex items-center gap-2"><Truck :size="14" class="text-muted-foreground shrink-0" /><span class="text-sm font-medium text-foreground">{{ r.supplier_name }}</span></div><div class="text-xs text-muted-foreground">{{ formatDate(r.created_at) }} · {{ r.items?.length || 0 }} items</div></div>
        </div>
        <div v-else class="px-4 py-6 text-center text-sm text-muted-foreground">No receipts yet.</div>
      </div>
    </template>

    <!-- ===== NEW RECEIPT ===== -->
    <template v-if="view === 'receipt'">
      <div class="flex items-center gap-3 mb-6">
        <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground" @click="view = 'dashboard'"><ChevronLeft :size="16" /></Button>
        <div><h1 class="text-xl font-bold text-foreground">New Receipt</h1><p class="text-sm text-muted-foreground">Record incoming items from a supplier</p></div>
      </div>

      <!-- Step 1: Supplier -->
      <div class="rounded-lg border border-border bg-card p-5 mb-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">1</div>
          <span class="text-sm font-semibold text-foreground">Supplier</span>
          <span v-if="supplier" class="text-xs text-emerald-600 ml-2">✓ {{ supplier }}</span>
        </div>
        <div class="relative">
          <Building :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input v-model="supplier" type="text" placeholder="Enter supplier name..." ref="supplierInputRef" class="no-spinners w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring" :class="{ 'border-red-500': validationErrors.supplier }" @input="onSupplierInput" />
        </div>
        <ul v-if="supplierSuggestions.length" class="mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-hidden">
          <li v-for="s in supplierSuggestions" :key="s" class="flex items-center gap-2 px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors" @click="selectSupplier(s)"><Building :size="14" class="text-muted-foreground" /><span>{{ s }}</span></li>
        </ul>
      </div>

      <!-- Step 2: Add Products -->
      <div class="rounded-lg border border-border bg-card p-5 mb-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">2</div>
          <span class="text-sm font-semibold text-foreground">Add Products</span>
          <span v-if="items.length" class="text-xs text-muted-foreground ml-2">· {{ items.length }} item{{ items.length !== 1 ? 's' : '' }}</span>
        </div>
        <div class="relative">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input v-model="productQuery" type="text" placeholder="Search products by name or barcode..." ref="searchInputRef" class="no-spinners w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring" @input="onProductSearch" @keydown.escape="productSuggestions = []" />
        </div>
        <div v-if="!items.length" class="flex flex-col items-center justify-center py-12 text-center"><Package :size="32" class="text-muted-foreground/30 mb-3" /><p class="text-sm text-muted-foreground">Search and select products above. If a product doesn't exist yet, create it below.</p></div>
        <ul v-if="productSuggestions.length" class="mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-y-auto max-h-56">
          <li v-for="(p,i) in productSuggestions" :key="p.id" class="flex items-center justify-between px-4 py-2.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors" :class="i < productSuggestions.length - 1 ? 'border-b border-border/50' : ''" @click="addProduct(p)">
            <div class="flex items-center gap-2.5"><PillBottle :size="16" class="text-muted-foreground/60 shrink-0" /><div><div class="font-medium text-foreground">{{ p.name }}</div><div class="text-xs text-muted-foreground">{{ p.manufacturer || '—' }} · {{ p.barcode || 'no barcode' }}</div></div></div>
            <div class="text-xs text-muted-foreground font-mono">&#8358;{{ (p.default_price?.selling_price || 0).toLocaleString() }}</div>
          </li>
        </ul>

        <!-- Create new product button -->
        <div class="mt-2.5 pt-2.5 border-t border-border/30">
          <Button variant="outline" size="sm" class="w-full" @click="openNewProductModal"><Plus :size="14" class="mr-1.5" />Create New Product</Button>
        </div>

        <!-- New Product Modal -->
        <div v-if="showNewProductModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm overflow-y-auto" @click.self="showNewProductModal = false">
          <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-lg mx-4 p-6 my-8">
            <h3 class="text-base font-semibold mb-4">New Product</h3>
            <div class="space-y-3">
              <div>
                <label class="text-xs text-muted-foreground mb-1 block">Product Name <span class="text-destructive">*</span></label>
                <input v-model="newProduct.name" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" @input="checkDuplicateProduct" />
                <div v-if="newProduct.duplicateMsg" class="text-xs text-amber-600 mt-1">{{ newProduct.duplicateMsg }}</div>
              </div>
              <div>
                <label class="text-xs text-muted-foreground mb-1 block">Manufacturer <span class="text-destructive">*</span></label>
                <div class="relative">
                  <input v-model="newProduct.manufacturer" type="text" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" @input="onManufacturerInput" />
                  <ul v-if="manufacturerSuggestions.length" class="absolute z-10 w-full mt-1 bg-popover border border-border rounded-lg shadow-lg overflow-hidden">
                    <li v-for="m in manufacturerSuggestions" :key="m" class="px-4 py-2 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors" @click="selectManufacturer(m)">{{ m }}</li>
                  </ul>
                </div>
              </div>
              <div><label class="text-xs text-muted-foreground mb-1 block">Barcode</label><input v-model="newProduct.barcode" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              <div class="grid grid-cols-2 gap-3">
                <div><label class="text-xs text-muted-foreground mb-1 block">Selling Price (&#8358;) <span class="text-destructive">*</span></label><input v-model.number="newProduct.selling_price" type="number" step="10" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
                <div><label class="text-xs text-muted-foreground mb-1 block">Cost Price (&#8358;) <span class="text-destructive">*</span></label><input v-model.number="newProduct.cost_price" type="number" step="10" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div><label class="text-xs text-muted-foreground mb-1 block">Category</label>
                  <select v-model.number="newProduct.category_id" class="w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring">
                    <option :value="0">Select...</option>
                    <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
                  </select>
                </div>
                <div><label class="text-xs text-muted-foreground mb-1 block">Reorder Level</label><input v-model.number="newProduct.reorder_level" type="number" class="no-spinners w-full text-sm border border-border rounded-md px-3 py-2 bg-background outline-none focus:ring-1 focus:ring-ring" /></div>
              </div>
            </div>
            <div class="flex items-center gap-2 mt-5 justify-end">
              <Button variant="outline" size="sm" @click="showNewProductModal = false">Cancel</Button>
              <Button size="sm" :disabled="newProductSaving" @click="saveNewProduct"><RotateCw v-if="newProductSaving" :size="14" class="animate-spin mr-1.5" /><Plus v-else :size="14" class="mr-1.5" />Create & Add</Button>
            </div>
          </div>
        </div>
      </div>

      <!-- Step 3: Set Details -->
      <div v-if="items.length" class="mb-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">3</div>
          <span class="text-sm font-semibold text-foreground">Set Details</span>
          <span class="text-xs text-muted-foreground ml-2">Cost, price, quantity, and expiry</span>
        </div>
        <div class="rounded-lg border border-border overflow-hidden">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30 text-[11px]">
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1.5">Item</th>
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1.5 w-24">Cost</th>
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1.5 w-28">Sell Price</th>
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1.5 w-14">Qty</th>
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1.5 w-28">Expiry</th>
                <th class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1.5 w-22">Total</th>
                <th class="w-8"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <template v-for="(item, idx) in items" :key="item._key">
                <tr class="hover:bg-muted/20 transition-colors">
                  <td class="px-2 py-2">
                    <div class="flex items-center gap-2">
                      <div class="text-sm font-medium text-foreground">{{ item.name }}</div>
                      <Button variant="ghost" size="icon" class="h-6 w-6 text-muted-foreground hover:text-primary" title="Price options" @click="openPriceOptions(item)"><Settings2 :size="12" /></Button>
                    </div>
                    <div class="text-[11px] text-muted-foreground">{{ item.manufacturer || '' }}</div>
                  </td>
                  <td class="px-2 py-2"><div class="relative"><span class="text-[11px] text-muted-foreground absolute left-1.5 top-1.5">&#8358;</span><input :value="item.cost_price" @input="item.cost_price = num($event.target.value); suggestPrice(item)" type="number" step="10" min="0" class="no-spinners w-full pl-4 py-1.5 text-right text-sm border border-border rounded-md bg-background outline-none font-medium" :class="{ 'border-red-500': item._errors?.cost }" /></div></td>
                  <td class="px-2 py-2">
                    <div class="relative">
                      <span class="text-[11px] text-muted-foreground absolute left-1.5 top-1.5">&#8358;</span>
                      <input :value="item.selling_price" @input="item.selling_price = num($event.target.value)" type="number" step="10" min="0" class="no-spinners w-full pl-4 py-1.5 text-right text-sm border border-border rounded-md bg-background outline-none font-medium" :class="{ 'border-red-500': item._errors?.sell }" />
                      <button v-if="item._suggestedPrice !== undefined" class="absolute -bottom-4 left-0 text-[10px] text-primary/70 hover:text-primary cursor-pointer" @click="item.selling_price = item._suggestedPrice; item._suggestedPrice = undefined">Suggested: &#8358;{{ item._suggestedPrice }}</button>
                    </div>
                  </td>
                  <td class="px-2 py-2"><div class="inline-flex items-center border border-border rounded-md overflow-hidden"><button class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent text-sm transition-colors" @click="item.quantity = Math.max(0, (item.quantity || 0) - 1)">−</button><input :value="item.quantity" @input="item.quantity = Math.max(0, num($event.target.value))" class="no-spinners h-7 w-9 text-center text-sm bg-transparent border-x border-border outline-none" /><button class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent text-sm transition-colors" @click="item.quantity = (item.quantity || 0) + 1">+</button></div></td>
                  <td class="px-2 py-2"><input :value="item.expiry" @input="item.expiry = $event.target.value" type="date" class="no-spinners w-full px-2 py-1.5 text-sm text-center border border-border rounded-md bg-background outline-none" :class="{ 'border-red-500': item._errors?.expiry }" /></td>
                  <td class="px-2 py-2 text-right font-semibold text-sm text-foreground">&#8358;{{ num(item.cost_price||0) * num(item.quantity||0) }}</td>
                  <td class="px-2 py-2"><Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-destructive" @click="removeItem(idx)"><X :size="13" /></Button></td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Step 4: Review Totals -->
      <div v-if="items.length" class="border border-border bg-card rounded-lg p-5 mb-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">4</div>
          <span class="text-sm font-semibold text-foreground">Review Totals</span>
        </div>
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="text-sm space-y-1.5"><div class="flex items-center gap-2"><Package :size="16" class="text-muted-foreground shrink-0" /><span>{{ items.length }} item{{ items.length !== 1 ? 's' : '' }}</span></div><div class="flex items-center gap-2"><span class="font-semibold">&#8358;{{ items.reduce((s,i) => s + num(i.cost_price||0) * num(i.quantity||0), 0).toLocaleString() }}</span><span class="text-muted-foreground">total cost</span></div></div>
          <div class="flex items-center gap-2"><Button variant="outline" size="sm" :disabled="submitting" @click="holdReceipt"><PauseCircle :size="14" class="mr-1.5" />Hold Draft</Button><Button size="lg" class="px-6 gap-2" :disabled="submitting" @click="receiveItems"><CircleCheck :size="16" />Receive Items</Button></div>
        </div>
        <div v-if="validationErrors.global" class="text-xs text-destructive mt-1.5 flex items-center gap-1"><AlertTriangle :size="12" />{{ validationErrors.global }}</div>
      </div>

      <!-- Price Options Modal -->
      <div v-if="priceOptionsTarget" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="closePriceOptions">
        <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold">Price Options</h3>
            <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" @click="closePriceOptions"><X :size="15" /></Button>
          </div>
          <p class="text-xs text-muted-foreground mb-4">Set alternative pricing for different pack sizes or formulations.</p>
          <div v-for="(po, pi) in priceOptionsTarget._priceOptions" :key="pi" class="flex items-center gap-2 text-xs border border-border/50 rounded-md px-3 py-2 mb-2">
            <input :value="po.name" @input="po.name = $event.target.value" class="no-spinners w-24 px-2 py-1.5 border border-border rounded-md bg-background outline-none" placeholder="Name" />
            <input :value="po.price" @input="po.price = num($event.target.value)" type="number" step="10" class="no-spinners w-22 px-2 py-1.5 text-right border border-border rounded-md bg-background outline-none" placeholder="Price" />
            <input :value="po.qty" @input="po.qty = num($event.target.value)" type="number" class="no-spinners w-16 px-2 py-1.5 text-right border border-border rounded-md bg-background outline-none" placeholder="Qty" />
            <Button variant="ghost" size="icon" class="h-6 w-6 text-muted-foreground hover:text-destructive" @click="priceOptionsTarget._priceOptions.splice(pi, 1)"><X :size="11" /></Button>
          </div>
          <div class="flex items-center justify-between mt-2">
            <Button variant="outline" size="sm" @click="priceOptionsTarget._priceOptions.push({ id: null, name: '', price: 0, qty: 1 })"><Plus :size="12" class="mr-1" />Add Option</Button>
            <Button variant="outline" size="sm" @click="closePriceOptions">Done</Button>
          </div>
        </div>
      </div>
    </template>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast" class="fixed bottom-6 right-6 z-50 flex items-center gap-2.5 bg-card border border-border px-4 py-3 rounded-xl shadow-2xl text-sm font-medium"><CircleCheck v-if="toastType === 'success'" :size="16" class="text-emerald-600" /><AlertCircle v-else :size="16" class="text-destructive" />{{ toast }}</div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { Search, X, Building, Package, PillBottle, PauseCircle, CircleCheck, Plus, AlertTriangle, AlertCircle, ChevronLeft, CalendarCheck, Truck, RotateCw, Settings2 } from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const API = "";

const VIEW_DASHBOARD = "dashboard";
const VIEW_RECEIPT = "receipt";

const view = ref(VIEW_DASHBOARD);
const supplier = ref("");
const supplierSuggestions = ref([]);
const productQuery = ref("");
const productSuggestions = ref([]);
const showNewProductModal = ref(false);
const newProductSaving = ref(false);
const newProduct = ref({ name: "", manufacturer: "", barcode: "", selling_price: 0, cost_price: 0, category_id: 0, reorder_level: 5, duplicateMsg: "" });
const categories = ref([]);
const manufacturerSuggestions = ref([]);
const items = ref([]);
const recentReceipts = ref([]);
const todayCount = ref(0);
const monthCount = ref(0);
const submitting = ref(false);
const validationErrors = ref({ supplier: false, global: "" });
const toast = ref(null);
const toastType = ref("success");
const priceOptionsTarget = ref(null);
const searchInputRef = ref(null);

let tTimer = null; let keyCounter = 0; let msTimer = null;

function num(v) { return Number(v || 0); }
function showToast(msg, type = "success") { toast.value = msg; toastType.value = type; clearTimeout(tTimer); tTimer = setTimeout(() => { toast.value = null; }, 3000); }

function suggestPrice(item) {
  if (num(item.cost_price) > 0) {
    const suggested = Math.round(num(item.cost_price) * 1.3 / 10) * 10;
    if (!num(item.selling_price) || num(item.selling_price) <= num(item.cost_price)) { item.selling_price = suggested; }
    else if (num(item.selling_price) <= suggested) { item._suggestedPrice = suggested; }
  }
}

// === Dashboard ===
async function fetchDashboard() {
  try {
    const r = await fetch(`${API}/inventory/received-items-history/api`);
    if (r.ok) { const d = await r.json(); recentReceipts.value = (d.batches || []).slice(0, 5); }
  } catch {}
  todayCount.value = recentReceipts.value.length; monthCount.value = recentReceipts.value.length;
}

function formatDate(d) { if (!d) return ""; try { return new Date(d).toLocaleDateString(); } catch { return ""; } }

// === Navigation ===
function startNewReceipt() {
  view.value = VIEW_RECEIPT; supplier.value = ""; items.value.splice(0);
  productQuery.value = ""; validationErrors.value = { supplier: false, global: "" }; fetchCategories();
}

// === Categories ===
async function fetchCategories() {
  try {
    const r = await fetch(`${API}/inventory/item-list`);
    if (r.ok) { const d = await r.json(); categories.value = d.categories || []; }
  } catch {}
}

// === Supplier search (starts from 1 char) ===
function onSupplierInput() { validationErrors.value.supplier = false; debouncedSupplierSearch(); }
function selectSupplier(s) { supplier.value = s; supplierSuggestions.value = []; validationErrors.value.supplier = false; }
let ssTimer = null;
function debouncedSupplierSearch() {
  const q = supplier.value.trim();
  if (q.length < 1) { supplierSuggestions.value = []; return; }
  clearTimeout(ssTimer);
  ssTimer = setTimeout(async () => {
    try { const r = await fetch(`${API}/inventory/suppliers/search?query=${encodeURIComponent(q)}`); if (r.ok) { const d = await r.json(); supplierSuggestions.value = d.data || []; } } catch {}
  }, 200);
}

// === Product Search ===
function onProductSearch() { debouncedProductSearch(); }
let psTimer = null;
function debouncedProductSearch() {
  const q = productQuery.value.trim();
  if (q.length < 2) { productSuggestions.value = []; return; }
  clearTimeout(psTimer);
  psTimer = setTimeout(async () => {
    try { const r = await fetch(`${API}/inventory/search?query=${encodeURIComponent(q)}`); if (r.ok) productSuggestions.value = await r.json(); } catch {}
  }, 250);
}

// === Add Product ===
function addProduct(p) {
  productQuery.value = ""; productSuggestions.value = []; keyCounter++;
  const cost = p.cost_price || (p.default_price?.selling_price ? p.default_price.selling_price * 0.7 : 0);
  const sell = p.default_price?.selling_price || 0;
  const suggestedSell = sell > 0 ? sell : (cost > 0 ? Math.round(cost * 1.3 / 10) * 10 : 0);
  const priceOpts = (p.price_options || []).map((po) => ({ id: po.id, name: po.name, price: po.selling_price, qty: po.quantity_per_unit || 1 }));
  items.value.push({
    _key: p.id + "-" + keyCounter, id: p.id, name: p.name, manufacturer: p.manufacturer || "",
    barcode: p.barcode || "", cost_price: cost, selling_price: suggestedSell, quantity: 1, expiry: "",
    _errors: {}, _priceOptions: priceOpts, _suggestedPrice: undefined,
  });
  refocusSearch();
}
function removeItem(idx) { items.value.splice(idx, 1); }
function refocusSearch() { setTimeout(() => { const el = searchInputRef.value; if (el) el.focus(); }, 50); }

// === Price Options Modal ===
function openPriceOptions(item) { priceOptionsTarget.value = item; }
function closePriceOptions() { priceOptionsTarget.value = null; }

// === New Product Modal ===
function openNewProductModal() {
  productSuggestions.value = []; showNewProductModal.value = true;
  newProduct.value = { name: "", manufacturer: "", barcode: "", selling_price: 0, cost_price: 0, category_id: 0, reorder_level: 5, duplicateMsg: "" };
  manufacturerSuggestions.value = [];
}

function checkDuplicateProduct() {
  const name = newProduct.value.name.trim().toLowerCase();
  const mfr = newProduct.value.manufacturer.trim().toLowerCase();
  newProduct.value.duplicateMsg = "";
  // Check is done async on manufacturer input instead
}

// === Manufacturer autocomplete ===
function onManufacturerInput() {
  const q = newProduct.value.manufacturer.trim();
  if (q.length < 1) { manufacturerSuggestions.value = []; return; }
  clearTimeout(msTimer);
  msTimer = setTimeout(async () => {
    try {
      const r = await fetch(`${API}/inventory/search?query=${encodeURIComponent(q)}`);
      if (r.ok) {
        const products = await r.json();
        const seen = new Set();
        manufacturerSuggestions.value = products.filter(p => p.manufacturer).map(p => p.manufacturer).filter(m => { if (seen.has(m)) return false; seen.add(m); return true; }).slice(0, 8);
      }
    } catch {}
    checkDuplicateProduct();
  }, 200);
}
function selectManufacturer(m) {
  newProduct.value.manufacturer = m; manufacturerSuggestions.value = [];
  checkDuplicateProduct();
}

function checkDuplicateProduct() {
  const name = newProduct.value.name.trim();
  const mfr = newProduct.value.manufacturer.trim();
  if (name.length < 2 || mfr.length < 2) { newProduct.value.duplicateMsg = ""; return; }
  fetch(`${API}/inventory/search?query=${encodeURIComponent(name)}`)
    .then(r => r.ok ? r.json() : [])
    .then(products => {
      const dup = products.find(p => p.name?.toLowerCase() === name.toLowerCase() && p.manufacturer?.toLowerCase() === mfr.toLowerCase());
      newProduct.value.duplicateMsg = dup ? '⚠ "' + dup.name + '" by ' + dup.manufacturer + ' already exists' : "";
    }).catch(() => {});
}

async function saveNewProduct() {
  if (!newProduct.value.name.trim()) { showToast("Product name is required", "error"); return; }
  if (!newProduct.value.manufacturer.trim()) { showToast("Manufacturer is required", "error"); return; }
  if (num(newProduct.value.selling_price) <= 0) { showToast("Selling price is required", "error"); return; }
  if (newProduct.value.duplicateMsg) { showToast("A product with this name and manufacturer already exists", "error"); return; }
  newProductSaving.value = true;
  try {
    const r = await fetch(`${API}/inventory/add-item`, {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: newProduct.value.name.trim(), manufacturer: newProduct.value.manufacturer.trim(),
        barcode: newProduct.value.barcode.trim(), category_id: num(newProduct.value.category_id) || 1,
        reorder_level: num(newProduct.value.reorder_level) || 5, cost_price: num(newProduct.value.cost_price),
        selling_price: num(newProduct.value.selling_price),
      }),
    });
    if (!r.ok) throw new Error("Failed to create product");
    const created = await r.json();
    showNewProductModal.value = false;
    const suggestedSell = num(newProduct.value.selling_price) || Math.round(num(newProduct.value.cost_price) * 1.3 / 10) * 10;
    items.value.push({
      _key: "new-" + keyCounter++, id: created.id, name: newProduct.value.name.trim(),
      manufacturer: newProduct.value.manufacturer.trim(), barcode: newProduct.value.barcode.trim(),
      cost_price: num(newProduct.value.cost_price), selling_price: suggestedSell, quantity: 1, expiry: "",
      _errors: {}, _priceOptions: [], _suggestedPrice: undefined,
    });
    showToast("Product created and added"); refocusSearch();
  } catch (e) { showToast(e.message || "Failed to create product", "error"); }
  finally { newProductSaving.value = false; }
}

// === Hold (no validation) ===
async function holdReceipt() {
  submitting.value = true;
  try {
    const p = { reference: "", payload: JSON.stringify({ supplier: supplier.value.trim(), products: items.value.map(i => ({ id: i.id, barcode: i.barcode, cost_price: i.cost_price, selling_price: i.selling_price, quantity: i.quantity, expiry: i.expiry || null, price_options_changes: (i._priceOptions || []).map(po => ({ id: po.id, name: po.name, selling_price: po.price, quantity_per_unit: po.qty || 1 })) })) }) };
    const r = await fetch(`${API}/inventory/receive-items/hold`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(p) });
    if (!r.ok) throw new Error("Failed to hold");
    view.value = VIEW_DASHBOARD; showToast("Receipt saved as draft"); fetchDashboard();
  } catch (e) { showToast(e.message || "Failed to hold", "error"); }
  finally { submitting.value = false; }
}

// === Validate + Receive ===
function validate() {
  let valid = true; validationErrors.value = { supplier: false, global: "" };
  if (!supplier.value.trim()) { validationErrors.value.supplier = true; valid = false; }
  if (!items.value.length) { validationErrors.value.global = "Add at least one product."; valid = false; }
  for (const item of items.value) {
    item._errors = {};
    if (num(item.cost_price) <= 0) { item._errors.cost = true; valid = false; }
    if (num(item.selling_price) <= 0) { item._errors.sell = true; valid = false; }
    if (num(item.quantity) <= 0) { item._errors.qty = true; valid = false; }
    if (!item.expiry) { item._errors.expiry = true; valid = false; }
  }
  if (!valid && !validationErrors.value.global) validationErrors.value.global = "Fill in all required fields marked in red.";
  return valid;
}

async function receiveItems() {
  if (!validate()) return;
  submitting.value = true;
  try {
    const p = { supplier: supplier.value.trim(), products: items.value.map(i => ({ id: i.id, barcode: i.barcode, cost_price: i.cost_price, selling_price: i.selling_price, quantity: i.quantity, expiry: i.expiry || null, price_options_changes: (i._priceOptions || []).map(po => ({ id: po.id, name: po.name, selling_price: po.price, quantity_per_unit: po.qty || 1 })) })), idempotency_key: (() => { try { return crypto.randomUUID(); } catch { return Date.now() + "-" + Math.random().toString(36).slice(2); } })() };
    const r = await fetch(`${API}/inventory/receive-items`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(p) });
    if (!r.ok) { const e = await r.json().catch(() => ({})); throw new Error(e.error || `HTTP ${r.status}`); }
    view.value = VIEW_DASHBOARD; showToast("Items received successfully"); fetchDashboard();
  } catch (e) { showToast(e.message || "Failed to receive", "error"); }
  finally { submitting.value = false; }
}

onMounted(() => { fetchDashboard(); fetchCategories(); });
</script>

<style scoped>
input.no-spinners::-webkit-outer-spin-button,
input.no-spinners::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
input.no-spinners[type="number"] { -moz-appearance: textfield; }
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>