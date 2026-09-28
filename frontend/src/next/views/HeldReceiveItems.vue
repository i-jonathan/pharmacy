<template>
  <div class="flex h-[calc(100vh-3.5rem)]">
    <!-- Main table area -->
    <div class="flex-1 flex flex-col overflow-hidden" :class="{ 'border-r border-border': detailHeld }">
      <!-- Header -->
      <div class="p-6 pb-0">
        <h1 class="text-2xl font-bold text-foreground">Held Receive Items</h1>
        <p class="text-sm text-muted-foreground mt-1">Draft and incomplete inventory receipts</p>
      </div>

      <!-- Toolbar -->
      <div class="p-6 pb-4 flex items-center gap-2">
        <Button variant="outline" size="sm" :disabled="loading" @click="fetchHeld" class="gap-2">
          <RotateCw :size="14" :class="{ 'animate-spin': loading }" />
          Refresh
        </Button>
        <span v-if="!loading" class="text-xs text-muted-foreground ml-2">
          {{ heldItems.length }} held receipt{{ heldItems.length !== 1 ? 's' : '' }}
        </span>
      </div>

      <!-- Content area -->
      <div class="flex-1 overflow-y-auto px-6 pb-4">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-24 text-muted-foreground">
          <RotateCw :size="20" class="animate-spin mr-3" />
          <span class="text-sm">Loading held receipts...</span>
        </div>

        <!-- Error -->
        <div v-else-if="error" class="flex flex-col items-center justify-center py-24 text-center">
          <AlertCircle :size="40" class="text-destructive/40 mb-3" />
          <p class="text-sm text-muted-foreground mb-3">{{ error }}</p>
          <Button variant="outline" size="sm" @click="fetchHeld">Retry</Button>
        </div>

        <!-- Empty -->
        <div v-else-if="heldItems.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
          <PauseCircle :size="48" class="text-muted-foreground/40 mb-4" />
          <h3 class="text-lg font-semibold text-foreground mb-1">No held receipts</h3>
          <p class="text-sm text-muted-foreground max-w-sm">
            Incomplete inventory receipts will appear here so you can resume or discard them later.
          </p>
          <Button variant="outline" class="mt-4" @click="$router.push('/receive-items')">
            <Truck :size="16" class="mr-2" />
            Go to Receive Items
          </Button>
        </div>

        <!-- Held Table -->
        <div v-else class="border border-border rounded-lg overflow-hidden">
          <table class="w-full text-sm" role="grid">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Reference</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Date</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Supplier</th>
                <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-16">Items</th>
                <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-28">Total Cost</th>
                <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-24">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr
                v-for="(held, i) in heldItems"
                :key="held.reference"
                class="cursor-pointer transition-colors outline-none"
                :class="rowClass(i)"
                tabindex="0"
                @click="selectAndShow(i)"
                @keydown.enter.prevent="selectAndShow(i)"
                @keydown.space.prevent="selectAndShow(i)"
              >
                <td class="px-4 py-3 font-mono text-xs font-medium text-foreground">{{ held.reference }}</td>
                <td class="px-4 py-3 text-muted-foreground text-xs whitespace-nowrap">{{ formatDate(held.updated_at || held.created_at) }}</td>
                <td class="px-4 py-3 text-muted-foreground text-xs">{{ getSupplier(held) }}</td>
                <td class="px-4 py-3 text-center text-muted-foreground">{{ getProducts(held).length }}</td>
                <td class="px-4 py-3 text-right font-semibold text-foreground">&#8358;{{ totalCost(held).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</td>
                <td class="px-4 py-3 text-right">
                  <div class="flex items-center justify-end gap-1">
                    <Button variant="ghost" size="icon" class="h-8 w-8 text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-900/20" title="Restore receipt" @click.stop="restoreHeld(held)">
                      <Play :size="15" />
                    </Button>
                    <Button variant="ghost" size="icon" class="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10" title="Delete held receipt" @click.stop="confirmDelete = held.reference">
                      <Trash2 :size="15" />
                    </Button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Detail Sidebar -->
    <Transition name="slide-panel">
      <div
        v-if="detailHeld"
        class="flex flex-col h-full w-[40%] min-w-[360px] max-w-[600px] bg-card border-l border-border flex-shrink-0"
      >
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-border">
          <div>
            <h2 class="text-sm font-semibold">Receipt Details</h2>
            <p class="text-xs text-muted-foreground font-mono">{{ detailHeld.reference }}</p>
          </div>
          <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" @click="closeDetail">
            <X :size="15" />
          </Button>
        </div>

        <!-- Supplier / Date -->
        <div class="px-4 py-3 border-b border-border">
          <div class="flex items-center gap-2 text-xs text-muted-foreground mb-1">
            <span class="font-medium text-foreground">{{ getSupplier(detailHeld) }}</span>
            <span>&middot;</span>
            <span>{{ formatDate(detailHeld.updated_at || detailHeld.created_at) }}</span>
          </div>
        </div>

        <!-- Scrollable content area -->
        <div class="flex-1 overflow-y-auto">
          <!-- Items Table -->
          <div class="p-3">
            <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 px-1">Items</div>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead class="text-[10px]">Item</TableHead>
                  <TableHead class="w-10 text-[10px] text-center">Qty</TableHead>
                  <TableHead class="w-16 text-[10px] text-right">Cost</TableHead>
                  <TableHead class="w-16 text-[10px] text-right">Sell</TableHead>
                  <TableHead class="w-14 text-[10px] text-right">Total</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in getProducts(detailHeld)" :key="(item.id || item.name) + item.quantity">
                  <TableCell class="py-1.5">
                    <div class="text-xs font-medium">{{ item.name }}</div>
                    <div v-if="item.manufacturer" class="text-[10px] text-muted-foreground">{{ item.manufacturer }}</div>
                  </TableCell>
                  <TableCell class="py-1.5 text-center text-xs text-muted-foreground">{{ item.quantity }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs text-muted-foreground">&#8358;{{ Number(item.cost_price || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs text-muted-foreground">&#8358;{{ Number(item.selling_price || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs font-medium">&#8358;{{ lineTotal(item).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>

          <!-- Summary -->
          <div class="px-4 py-3 border-t border-border">
            <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Summary</div>
            <div class="space-y-2 bg-muted/50 rounded-sm px-3 py-3">
              <div class="flex justify-between text-xs">
                <span class="text-muted-foreground">Items</span>
                <span class="font-medium">{{ getProducts(detailHeld).length }}</span>
              </div>
              <div class="flex justify-between font-bold text-sm pt-2 border-t border-border">
                <span>Total Cost</span>
                <span>&#8358;{{ detailTotalCost.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Bottom Actions -->
        <div class="border-t border-border px-4 py-3 flex items-center gap-2 shrink-0">
          <Button class="flex-1 gap-2" @click="restoreHeld(detailHeld)">
            <Play :size="15" />
            Restore
          </Button>
          <Button variant="destructive" class="flex-1 gap-2" @click="confirmDelete = detailHeld.reference">
            <Trash2 :size="15" />
            Delete
          </Button>
        </div>
      </div>
    </Transition>

    <!-- Delete Confirm Modal -->
    <Transition name="fade">
      <div v-if="confirmDelete" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="confirmDelete = null">
        <div class="bg-card border border-border rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-full bg-destructive/10 flex items-center justify-center shrink-0">
              <AlertTriangle :size="20" class="text-destructive" />
            </div>
            <div>
              <h3 class="text-sm font-semibold">Delete Held Receipt</h3>
              <p class="text-xs text-muted-foreground">This cannot be undone</p>
            </div>
          </div>
          <p class="text-sm text-foreground mb-6">
            Are you sure you want to delete <span class="font-mono font-medium">{{ confirmDelete }}</span>?
          </p>
          <div class="flex items-center gap-2 justify-end">
            <Button variant="outline" size="sm" @click="confirmDelete = null">Cancel</Button>
            <Button variant="destructive" size="sm" :disabled="confirmDelete === '__loading__'" @click="executeDelete">
              <Trash2 :size="14" class="mr-1.5" />
              Delete
            </Button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast" class="fixed bottom-6 right-6 z-50 flex items-center gap-2.5 bg-card border border-border px-4 py-3 rounded-xl shadow-2xl text-sm font-medium"><CircleCheck :size="16" class="text-emerald-600" />{{ toast }}</div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { RotateCw, AlertCircle, AlertTriangle, PauseCircle, Truck, Play, Trash2, X, CircleCheck } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const API = "";
const router = useRouter();

// State
const heldItems = ref([]);
const loading = ref(false);
const error = ref(null);
const detailHeld = ref(null);
const selectedIndex = ref(-1);
const confirmDelete = ref(null);
const toast = ref(null);
let tTimer = null;

const totalItems = computed(() => heldItems.value.length);

const detailTotalCost = computed(() => {
  if (!detailHeld.value) return 0;
  return getProducts(detailHeld.value).reduce(
    (sum, item) => sum + Number(item.cost_price || 0) * Number(item.quantity || 0), 0
  );
});

// Helpers on payload
function getSupplier(held) {
  const p = held.payload;
  if (!p) return "";
  return (typeof p === "string") ? (JSON.parse(p).supplier || "") : (p.supplier || "");
}

function getProducts(held) {
  const p = held.payload;
  if (!p) return [];
  const products = (typeof p === "string") ? (JSON.parse(p).products || []) : (p.products || []);
  return products;
}

function totalCost(held) {
  return getProducts(held).reduce(
    (sum, item) => sum + Number(item.cost_price || 0) * Number(item.quantity || 0), 0
  );
}

function lineTotal(item) {
  return Number(item.cost_price || 0) * Number(item.quantity || 0);
}

function rowClass(i) {
  return [
    "hover:bg-muted/20 focus:ring-1 focus:ring-ring focus:ring-inset",
    selectedIndex.value === i
      ? "bg-primary/5 ring-1 ring-inset ring-primary/20 font-medium"
      : "",
  ];
}

async function fetchHeld() {
  loading.value = true;
  error.value = null;
  detailHeld.value = null;
  selectedIndex.value = -1;
  try {
    const res = await fetch(`${API}/inventory/receive-items/held/api`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    heldItems.value = await res.json();
  } catch (e) {
    error.value = e.message || "Failed to load";
  } finally {
    loading.value = false;
  }
}

function select(i) {
  if (i < 0 || i >= heldItems.value.length) return;
  selectedIndex.value = i;
}

function selectAndShow(i) {
  select(i);
  detailHeld.value = heldItems.value[i];
}

function closeDetail() {
  detailHeld.value = null;
}

function restoreHeld(held) {
  if (!held) return;
  const payload = (typeof held.payload === "string") ? JSON.parse(held.payload) : held.payload;
  localStorage.setItem("heldReceiveItems", JSON.stringify({
    reference: held.reference,
    payload: payload,
  }));
  router.push("/receive-items");
}

function executeDelete() {
  if (!confirmDelete.value) return;
  const ref = confirmDelete.value;
  confirmDelete.value = "__loading__";
  fetch(`${API}/inventory/receive-items/held/${encodeURIComponent(ref)}`, { method: "DELETE" })
    .then((res) => {
      if (!res.ok) throw new Error(`Failed to delete (${res.status})`);
      heldItems.value = heldItems.value.filter((h) => h.reference !== ref);
      if (detailHeld.value?.reference === ref) closeDetail();
      selectedIndex.value = -1;
      confirmDelete.value = null;
      showToast("Held receipt deleted");
    })
    .catch((e) => {
      error.value = e.message || "Failed to delete";
      confirmDelete.value = ref;
    });
}

function formatDate(iso) {
  if (!iso) return "—";
  const d = new Date(iso);
  return d.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function showToast(msg, type = "success") {
  toast.value = msg;
  clearTimeout(tTimer);
  tTimer = setTimeout(() => { toast.value = null; }, 3000);
}

onMounted(() => {
  fetchHeld();
});
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
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>