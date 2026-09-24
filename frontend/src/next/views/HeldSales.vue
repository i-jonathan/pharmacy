<template>
  <div class="flex h-[calc(100vh-3.5rem)]">
    <!-- Main table area -->
    <div
      class="flex-1 flex flex-col overflow-hidden"
      :class="{ 'border-r border-border': detailHeld }"
    >
      <!-- Header -->
      <div class="p-6 pb-0">
        <h1 class="text-2xl font-bold text-foreground">Held Sales</h1>
        <p class="text-sm text-muted-foreground mt-1">Resume or void temporarily paused transactions</p>
      </div>

      <!-- Toolbar -->
      <div class="p-6 pb-4 flex items-center gap-2">
        <Button variant="outline" size="sm" :disabled="loading" @click="fetchHeld" class="gap-2">
          <RotateCw :size="14" :class="{ 'animate-spin': loading }" />
          Refresh
        </Button>
        <span v-if="!loading" class="text-xs text-muted-foreground ml-2">
          {{ heldSales.length }} held sale{{ heldSales.length !== 1 ? 's' : '' }}
        </span>
      </div>

      <!-- Content area -->
      <div class="flex-1 overflow-y-auto px-6 pb-4">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-24 text-muted-foreground">
          <RotateCw :size="20" class="animate-spin mr-3" />
          <span class="text-sm">Loading held sales...</span>
        </div>

        <!-- Error -->
        <div v-else-if="error" class="flex flex-col items-center justify-center py-24 text-center">
          <AlertCircle :size="40" class="text-destructive/40 mb-3" />
          <p class="text-sm text-muted-foreground mb-3">{{ error }}</p>
          <Button variant="outline" size="sm" @click="fetchHeld">Retry</Button>
        </div>

        <!-- Empty -->
        <div v-else-if="heldSales.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
          <PauseCircle :size="40" class="text-muted-foreground/40 mb-3" />
          <h3 class="text-base font-semibold text-foreground mb-1">No held sales</h3>
          <p class="text-sm text-muted-foreground max-w-sm">
            Sales paused from the Point of Sale will appear here. You can resume or void them later.
          </p>
          <Button variant="outline" class="mt-4" @click="$router.push('/pos')">
            <ShoppingCart :size="16" class="mr-2" />
            Go to POS
          </Button>
        </div>

        <!-- Held Sales Table -->
        <div v-else class="border border-border rounded-lg overflow-hidden">
          <table class="w-full text-sm" role="grid">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Reference</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Date</th>
                <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-16">Items</th>
                <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-28">Total</th>
                <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-24">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr
                v-for="(held, i) in heldSales"
                :key="held.reference"
                class="cursor-pointer transition-colors outline-none"
                :class="rowClass(i)"
                tabindex="0"
                @click="selectAndShow(i)"
                @keydown.enter.prevent="selectAndShow(i)"
                @keydown.space.prevent="selectAndShow(i)"
              >
                <td class="px-4 py-3 font-mono text-xs font-medium text-foreground">{{ held.reference }}</td>
                <td class="px-4 py-3 text-muted-foreground text-xs whitespace-nowrap">{{ formatDate(held.updated_at) }}</td>
                <td class="px-4 py-3 text-center text-muted-foreground">{{ getCart(held).length }}</td>
                <td class="px-4 py-3 text-right font-semibold text-foreground">&#8358;{{ computeTotal(held).toLocaleString() }}</td>
                <td class="px-4 py-3 text-right">
                  <div class="flex items-center justify-end gap-1">
                    <Button variant="ghost" size="icon" class="h-8 w-8 text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-900/20" title="Resume sale" @click.stop="resumeSale(held)">
                      <Play :size="15" />
                    </Button>
                    <Button variant="ghost" size="icon" class="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10" title="Void this held sale" @click.stop="voidSale(held)">
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
            <h2 class="text-sm font-semibold">Held Sale</h2>
            <p class="text-xs text-muted-foreground font-mono">{{ detailHeld.reference }}</p>
          </div>
          <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" @click="closeDetail">
            <X :size="15" />
          </Button>
        </div>

        <!-- Date / Info -->
        <div class="px-4 py-3 border-b border-border">
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <span>Held {{ formatDate(detailHeld.updated_at) }}</span>
          </div>
        </div>

        <!-- Scrollable content -->
        <div class="flex-1 overflow-y-auto">
          <!-- Items Table -->
          <div class="p-3">
            <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 px-1">Items</div>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead class="text-[10px]">Item</TableHead>
                  <TableHead class="w-12 text-[10px] text-center">Qty</TableHead>
                  <TableHead class="w-16 text-[10px] text-right">Price</TableHead>
                  <TableHead class="w-14 text-[10px] text-right">Total</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="(item, idx) in getCart(detailHeld)" :key="idx">
                  <TableCell class="py-1.5">
                    <div class="text-xs font-medium">{{ item.name }}</div>
                    <div v-if="item.manufacturer" class="text-[10px] text-muted-foreground">{{ item.manufacturer }}</div>
                  </TableCell>
                  <TableCell class="py-1.5 text-center text-xs text-muted-foreground">{{ item.qty }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs text-muted-foreground">&#8358;{{ Number(item.price || 0).toLocaleString() }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs font-medium">&#8358;{{ (item.price * item.qty).toLocaleString() }}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>

          <!-- Summary -->
          <div class="px-4 py-3 border-t border-border">
            <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Summary</div>
            <div class="bg-muted/50 rounded-sm px-3 py-3 space-y-2">
              <div class="flex justify-between text-xs">
                <span class="text-muted-foreground">Items</span>
                <span class="font-medium">{{ getCart(detailHeld).length }} product{{ getCart(detailHeld).length !== 1 ? 's' : '' }}</span>
              </div>
              <div class="flex justify-between font-bold text-sm pt-2 border-t border-border">
                <span>Total</span>
                <span>&#8358;{{ computeTotal(detailHeld).toLocaleString() }}</span>
              </div>
            </div>
          </div>

          <!-- Note -->
          <div v-if="getPayload(detailHeld).orderNote" class="px-4 py-3 border-t border-border">
            <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">Note</div>
            <p class="text-xs text-foreground bg-muted/30 rounded-sm px-3 py-2">{{ getPayload(detailHeld).orderNote }}</p>
          </div>
        </div>

        <!-- Bottom Actions -->
        <div class="border-t border-border px-4 py-3 flex items-center gap-2">
          <Button class="flex-1 gap-2" @click="resumeSale(detailHeld)">
            <Play :size="15" />
            Resume
          </Button>
          <Button variant="destructive" class="flex-1 gap-2" @click="confirmVoid = detailHeld.reference">
            <Trash2 :size="15" />
            Void
          </Button>
        </div>
      </div>
    </Transition>

    <!-- Void Confirm Modal -->
    <Transition name="fade">
      <div v-if="confirmVoid" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="confirmVoid = null">
        <div class="bg-card border border-border rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-full bg-destructive/10 flex items-center justify-center shrink-0">
              <AlertTriangle :size="20" class="text-destructive" />
            </div>
            <div>
              <h3 class="text-sm font-semibold">Void Held Sale</h3>
              <p class="text-xs text-muted-foreground">This cannot be undone</p>
            </div>
          </div>
          <p class="text-sm text-foreground mb-6">
            Are you sure you want to void <span class="font-mono font-medium">{{ confirmVoid }}</span>?
          </p>
          <div class="flex items-center gap-2 justify-end">
            <Button variant="outline" size="sm" @click="confirmVoid = null">Cancel</Button>
            <Button variant="destructive" size="sm" :disabled="confirmVoid === '__loading__'" @click="executeVoid">
              <Trash2 :size="14" class="mr-1.5" />
              Void
            </Button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { PauseCircle, ShoppingCart, Play, Trash2, RotateCw, AlertCircle, AlertTriangle, X } from "lucide-vue-next";
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

const heldSales = ref([]);
const loading = ref(false);
const error = ref(null);
const detailHeld = ref(null);
const selectedIndex = ref(-1);
const confirmVoid = ref(null);

function getPayload(held) {
  if (!held || !held.payload) return {};
  return held.payload;
}

function getCart(held) {
  return getPayload(held).cart || [];
}

function computeTotal(held) {
  const cart = getCart(held);
  return cart.reduce((sum, item) => sum + Number(item.price || 0) * Number(item.qty || 0), 0);
}

function rowClass(i) {
  return [
    "hover:bg-muted/20 focus:ring-1 focus:ring-ring focus:ring-inset",
    selectedIndex.value === i
      ? "bg-primary/5 ring-1 ring-inset ring-primary/20 font-medium"
      : "",
  ];
}

function select(i) {
  if (i < 0 || i >= heldSales.value.length) return;
  selectedIndex.value = i;
}

function selectAndShow(i) {
  select(i);
  detailHeld.value = heldSales.value[i];
}

function closeDetail() {
  detailHeld.value = null;
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

function resumeSale(held) {
  // Save the held transaction to localStorage so the POS can restore it
  localStorage.setItem("resumeHeldSale", JSON.stringify(held));
  router.push({ name: "pos" });
}

async function executeVoid() {
  if (!confirmVoid.value) return;
  const ref = confirmVoid.value;
  confirmVoid.value = "__loading__";
  try {
    const res = await fetch(`${API}/sales/held/${ref}`, { method: "DELETE" });
    if (!res.ok) throw new Error("Failed to void held sale");
    heldSales.value = heldSales.value.filter((h) => h.reference !== ref);
    if (detailHeld.value?.reference === ref) closeDetail();
    selectedIndex.value = -1;
    confirmVoid.value = null;
  } catch (e) {
    error.value = e.message || "Failed to void";
    confirmVoid.value = ref;
  }
}

async function fetchHeld() {
  loading.value = true;
  error.value = null;
  detailHeld.value = null;
  selectedIndex.value = -1;
  try {
    const res = await fetch(`${API}/sales/api/held`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    heldSales.value = await res.json();
  } catch (e) {
    error.value = e.message || "Failed to load";
  } finally {
    loading.value = false;
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
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>