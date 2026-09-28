<template>
  <div class="flex h-[calc(100vh-3.5rem)]">
    <!-- Main table area -->
    <div class="flex-1 flex flex-col overflow-hidden" :class="{ 'border-r border-border': detailHeld }">
      <!-- Header -->
      <div class="p-6 pb-0">
        <h1 class="text-2xl font-bold text-foreground">Held Receive Items</h1>
        <p class="text-sm text-muted-foreground mt-1">Draft and incomplete inventory receipts</p>
      </div>

      <!-- Summary bar -->
      <div class="px-6 pb-3 flex items-center justify-between text-sm">
        <span class="text-muted-foreground">
          <template v-if="!loading">{{ totalItems }} held receipt{{ totalItems !== 1 ? 's' : '' }}</template>
        </span>
      </div>

      <!-- Scrollable table area -->
      <div class="flex-1 overflow-y-auto px-6 pb-4" ref="tableContainerRef" @keydown="onTableKeydown" tabindex="-1">
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
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr
                v-for="(held, i) in heldItems"
                :key="held.reference"
                :ref="(el) => rowRefs[i] = el"
                class="cursor-pointer transition-colors outline-none"
                :class="rowClass(i)"
                :data-index="i"
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
        class="flex flex-col h-full w-[40%] min-w-[360px] bg-card border-l border-border flex-shrink-0"
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

          <!-- Totals -->
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

        <!-- Footer actions -->
        <div class="flex items-center gap-3 px-4 py-3.5 border-t border-border shrink-0">
          <Button variant="outline" size="sm" class="gap-2" :disabled="deleting" @click="deleteHeld">
            <Trash2 :size="14" class="shrink-0" />
            Delete
          </Button>
          <div class="flex-1"></div>
          <Button size="sm" class="gap-2" @click="restoreHeld">
            <RefreshCw :size="14" class="shrink-0" />
            Restore
          </Button>
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
import { ref, computed, onMounted, nextTick } from "vue";
import { RotateCw, AlertCircle, PauseCircle, Truck, Trash2, RefreshCw, X, CircleCheck } from "lucide-vue-next";
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

// State
const heldItems = ref([]);
const loading = ref(false);
const error = ref(null);
const detailHeld = ref(null);
const selectedIndex = ref(-1);
const deleting = ref(false);
const rowRefs = ref([]);
const tableContainerRef = ref(null);
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
    const data = await res.json();
    heldItems.value = data || [];
  } catch (e) {
    error.value = e.message || "Failed to load";
  } finally {
    loading.value = false;
  }
}

function select(i) {
  if (i < 0 || i >= heldItems.value.length) return;
  selectedIndex.value = i;
  nextTick(() => {
    const el = rowRefs.value[i];
    if (el && typeof el === "object" && "$el" in el) {
      el.$el?.scrollIntoView?.({ block: "nearest" });
    } else if (el?.scrollIntoView) {
      el.scrollIntoView({ block: "nearest" });
    }
  });
}

function selectAndShow(i) {
  select(i);
  openDetailFor(i);
}

function openDetailFor(i) {
  const held = heldItems.value[i];
  if (!held) return;
  detailHeld.value = held;
}

function closeDetail() {
  detailHeld.value = null;
  selectedIndex.value = -1;
}

function restoreHeld() {
  if (!detailHeld.value) return;
  // Save held data to localStorage so ReceiveItems can pick it up
  const held = detailHeld.value;
  const payload = (typeof held.payload === "string")
    ? JSON.parse(held.payload)
    : held.payload;
  localStorage.setItem("heldReceiveItems", JSON.stringify({
    reference: held.reference,
    payload: payload,
  }));
  // Navigate to receive-items — the old JS will restore from localStorage
  // For the Vue receive-items, the ReceiveItems.vue doesn't have restore
  // logic yet, but the legacy JS flow does. Navigate to the server route.
  window.location.href = "/inventory/receive-items";
}

async function deleteHeld() {
  if (!detailHeld.value) return;
  const reference = detailHeld.value.reference;
  deleting.value = true;
  try {
    const res = await fetch(`${API}/inventory/receive-items/held/${encodeURIComponent(reference)}`, {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    });
    if (!res.ok) throw new Error(`Failed to delete (${res.status})`);
    // Remove from list and close detail
    const idx = heldItems.value.map(h => h.reference).indexOf(reference);
    if (idx >= 0) heldItems.value.splice(idx, 1);
    closeDetail();
    showToast("Held receipt deleted");
  } catch (e) {
    alert(`Failed to delete: ${e.message}`);
  } finally {
    deleting.value = false;
  }
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

// --- Keyboard navigation ---
function onTableKeydown(e) {
  if (heldItems.value.length === 0) return;
  if (e.key === "ArrowDown") {
    e.preventDefault();
    const next = Math.min(selectedIndex.value + 1, heldItems.value.length - 1);
    if (next >= 0) select(next);
    if (detailHeld.value) openDetailFor(next);
  }
  if (e.key === "ArrowUp") {
    e.preventDefault();
    const prev = Math.max(selectedIndex.value - 1, 0);
    if (prev >= 0) select(prev);
    if (detailHeld.value) openDetailFor(prev);
  }
  if (e.key === "Enter" || e.key === " ") {
    e.preventDefault();
    if (selectedIndex.value >= 0) {
      if (detailHeld.value && heldItems.value[selectedIndex.value]?.reference === detailHeld.value.reference) {
        closeDetail();
      } else {
        openDetailFor(selectedIndex.value);
      }
    }
  }
  if (e.key === "Escape") {
    closeDetail();
  }
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
</style>