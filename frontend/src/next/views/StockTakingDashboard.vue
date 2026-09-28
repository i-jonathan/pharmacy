<template>
  <div class="p-6 lg:p-8">
    <!-- ===== DASHBOARD ===== -->
    <template v-if="view === 'dashboard'">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h1 class="text-2xl font-bold text-foreground">Stock Taking</h1>
          <p class="text-sm text-muted-foreground mt-1">Conduct and manage inventory counts</p>
        </div>
        <Button size="lg" @click="showCreateModal = true"><Plus :size="16" class="mr-2" />New Stock Count</Button>
      </div>

      <!-- Metric Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div class="rounded-lg border border-border bg-card p-4">
          <div class="flex items-center gap-2 mb-1">
            <ClipboardCheck :size="16" class="text-emerald-500" />
            <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Completed</span>
          </div>
          <div class="text-2xl font-bold text-foreground">{{ metricCompleted }}</div>
        </div>
        <div class="rounded-lg border border-border bg-card p-4">
          <div class="flex items-center gap-2 mb-1">
            <Clock :size="16" class="text-amber-500" />
            <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">In Progress</span>
          </div>
          <div class="text-2xl font-bold text-foreground">{{ metricInProgress }}</div>
        </div>
        <div class="rounded-lg border border-border bg-card p-4">
          <div class="flex items-center gap-2 mb-1">
            <AlertTriangle :size="16" class="text-destructive" />
            <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Discrepancies</span>
          </div>
          <div class="text-2xl font-bold text-foreground">{{ metricDiscrepancies }}</div>
        </div>
        <div class="rounded-lg border border-border bg-card p-4">
          <div class="flex items-center gap-2 mb-1">
            <Package :size="16" class="text-primary" />
            <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Items Counted</span>
          </div>
          <div class="text-2xl font-bold text-foreground">{{ metricCounted }}</div>
        </div>
      </div>

      <!-- Table Section -->
      <div v-if="!loading && stockTakings.length === 0" class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
        <ClipboardCheck :size="48" class="text-muted-foreground/40 mb-4" />
        <h3 class="text-lg font-semibold text-foreground mb-1">No stock counts yet</h3>
        <p class="text-sm text-muted-foreground max-w-sm">Start a new stock count to track inventory levels and identify discrepancies.</p>
      </div>

      <div v-else-if="loading" class="flex items-center justify-center py-12 text-muted-foreground">
        <RotateCw :size="20" class="animate-spin mr-3" />
        <span class="text-sm">Loading...</span>
      </div>

      <div v-else-if="error" class="flex flex-col items-center justify-center py-12 text-center rounded-lg border border-border bg-card">
        <AlertCircle :size="32" class="text-destructive/40 mb-2" />
        <p class="text-sm text-muted-foreground">{{ error }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetchStockTakings">Retry</Button>
      </div>

      <div v-else class="rounded-lg border border-border overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border bg-muted/30 text-[11px]">
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Name</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Status</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Created By</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Started</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Completed</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Completed By</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/50">
            <tr
              v-for="st in stockTakings"
              :key="st.id"
              class="cursor-pointer hover:bg-muted/20 transition-colors"
              @click="openStockTaking(st.id)"
            >
              <td class="px-4 py-3 font-medium text-foreground">{{ st.name }}</td>
              <td class="px-4 py-3"><span :class="statusBadge(st.status)">{{ st.status }}</span></td>
              <td class="px-4 py-3 text-xs text-muted-foreground">{{ st.created_by }}</td>
              <td class="px-4 py-3 text-xs text-muted-foreground whitespace-nowrap">{{ formatDate(st.started_at) }}</td>
              <td class="px-4 py-3 text-xs text-muted-foreground whitespace-nowrap">{{ st.completed_at ? formatDate(st.completed_at) : '—' }}</td>
              <td class="px-4 py-3 text-xs text-muted-foreground">{{ st.completed_by || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <!-- ===== STOCK TAKING VIEW ===== -->
    <template v-if="view === 'counting'">
      <!-- Back -->
      <div class="flex items-center gap-3 mb-6">
        <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground" @click="closeStockTaking"><ChevronLeft :size="16" /></Button>
        <div>
          <h1 class="text-xl font-bold text-foreground">Stock Taking: {{ countingName }}</h1>
          <p class="text-sm text-muted-foreground">
            Started {{ formatDate(countingStarted) }} · {{ countingCreatedBy }} ·
            <span class="text-xs font-medium" :class="countingStatus === 'Completed' ? 'text-emerald-600' : 'text-primary'">{{ countingStatus }}</span>
          </p>
        </div>
        <div class="flex gap-2 ml-auto">
          <Button v-if="showQuantityAndVariance" variant="outline" size="sm" class="gap-2" @click="toggleVarianceFilter">
            <ListFilter :size="14" class="shrink-0" />
            {{ filterVariancesOnly ? 'Show All' : 'Filter Variances' }}
          </Button>
          <Button v-if="completeStockPermission" size="sm" :disabled="countingStatus === 'Completed'" class="gap-2" @click="completeStockTaking">
            <CircleCheck :size="14" class="shrink-0" />
            {{ countingStatus === 'Completed' ? 'Completed' : 'Complete' }}
          </Button>
        </div>
      </div>

      <!-- Search -->
      <div class="relative mb-4">
        <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search items..."
          class="w-full pl-9 pr-4 py-2.5 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
        />
      </div>

      <!-- Loading -->
      <div v-if="countingLoading" class="flex items-center justify-center py-12 text-muted-foreground">
        <RotateCw :size="20" class="animate-spin mr-3" />
        <span class="text-sm">Loading items...</span>
      </div>

      <!-- Error -->
      <div v-else-if="countingError" class="flex flex-col items-center justify-center py-12 text-center rounded-lg border border-border bg-card">
        <AlertCircle :size="32" class="text-destructive/40 mb-2" />
        <p class="text-sm text-muted-foreground">{{ countingError }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="loadStockTaking(countingId)">Retry</Button>
      </div>

      <!-- Items Table -->
      <div v-else-if="filteredItems.length" class="rounded-lg border border-border overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border bg-muted/30 text-[11px]">
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2">Item</th>
              <th class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-20">System Qty</th>
              <th class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-24">Dispensary</th>
              <th class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-24">Store</th>
              <th class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-20">Variance</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-24">Expiry</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2 w-28">Notes</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/50">
            <template v-for="(row, ri) in filteredItems" :key="ri">
              <!-- Category header -->
              <tr v-if="row._isCategory" class="bg-muted/20">
                <td colspan="7" class="px-3 py-2">
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-bold text-foreground uppercase tracking-wider">{{ row.category }}</span>
                    <span class="text-[11px] text-muted-foreground">· {{ row.count }} item{{ row.count !== 1 ? 's' : '' }}</span>
                    <span v-if="showQuantityAndVariance" class="text-[11px]" :class="row.totalVariance !== 0 ? 'text-destructive font-medium' : 'text-muted-foreground'">· Variance: {{ row.totalVariance > 0 ? '+' : '' }}{{ row.totalVariance }}</span>
                  </div>
                </td>
              </tr>
              <!-- Item row -->
              <tr v-else class="hover:bg-muted/20 transition-colors" :class="{ 'bg-destructive/5': hasVariance(row) }">
                <td class="px-3 py-2.5">
                  <div class="text-sm font-medium text-foreground">{{ row.product_name }}</div>
                  <div v-if="row.manufacturer" class="text-[11px] text-muted-foreground">{{ row.manufacturer }}</div>
                </td>
                <td class="px-3 py-2.5 text-right text-sm font-mono text-muted-foreground">{{ row.snapshot_quantity ?? '—' }}</td>
                <td class="px-3 py-2.5">
                  <input
                    v-if="countingStatus !== 'Completed'"
                    :value="row.dispensary_count"
                    @input="row.dispensary_count = num($event.target.value); queueUpdate(row)"
                    type="number" min="0"
                    class="no-spinners w-full px-2 py-1 text-sm text-right border border-border rounded-md bg-background outline-none font-mono"
                    :class="{ 'border-ring/50': row.dispensary_count !== row.snapshot_quantity }"
                  />
                  <span v-else class="text-sm font-mono text-foreground block text-right">{{ row.dispensary_count ?? '—' }}</span>
                </td>
                <td class="px-3 py-2.5">
                  <input
                    v-if="countingStatus !== 'Completed'"
                    :value="row.store_count"
                    @input="row.store_count = num($event.target.value); queueUpdate(row)"
                    type="number" min="0"
                    class="no-spinners w-full px-2 py-1 text-sm text-right border border-border rounded-md bg-background outline-none font-mono"
                    :class="{ 'border-ring/50': row.store_count !== row.snapshot_quantity }"
                  />
                  <span v-else class="text-sm font-mono text-foreground block text-right">{{ row.store_count ?? '—' }}</span>
                </td>
                <td class="px-3 py-2.5 text-right text-sm font-mono" :class="varianceClass(row)">{{ variance(row) > 0 ? '+' : '' }}{{ variance(row) }}</td>
                <td class="px-3 py-2.5 text-xs text-muted-foreground">{{ formatExpiry(row.earliest_expiry) }}</td>
                <td class="px-3 py-2.5">
                  <input
                    v-if="countingStatus !== 'Completed'"
                    :value="row.notes"
                    @input="row.notes = $event.target.value; queueUpdate(row)"
                    type="text"
                    class="no-spinners w-full px-2 py-1 text-xs border border-border rounded-md bg-background outline-none"
                    placeholder="—"
                  />
                  <span v-else class="text-xs text-muted-foreground">{{ row.notes || '—' }}</span>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <!-- Empty items -->
      <div v-else class="flex flex-col items-center justify-center py-12 text-center rounded-lg border border-border bg-card">
        <ClipboardCheck :size="32" class="text-muted-foreground/40 mb-2" />
        <p class="text-sm text-muted-foreground">No items match your search.</p>
      </div>

      <!-- Summary -->
      <div v-if="!countingLoading && filteredItems.length && showQuantityAndVariance" class="flex items-center justify-between px-4 py-3 bg-muted/30 rounded-lg border border-border text-sm mt-4">
        <span class="text-xs text-muted-foreground">Total Variance: <span class="font-semibold" :class="totalVariance !== 0 ? 'text-destructive' : 'text-foreground'">{{ totalVariance > 0 ? '+' : '' }}{{ totalVariance }}</span></span>
        <span class="text-xs text-muted-foreground">Issues: <span class="font-semibold">{{ totalIssues }}</span></span>
      </div>
    </template>

    <!-- ===== CREATE MODAL ===== -->
    <Transition name="fade">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="showCreateModal = false">
        <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
          <h2 class="text-base font-semibold mb-4">New Stock Taking</h2>
          <div v-if="createError" class="mb-3 text-xs text-destructive bg-destructive/10 rounded-md px-3 py-2">{{ createError }}</div>
          <label class="text-xs text-muted-foreground mb-1.5 block">Name</label>
          <input
            v-model="newName"
            @keydown.enter="handleCreate"
            type="text"
            placeholder="e.g. September 2026 Stock Count"
            class="w-full px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
          />
          <div class="flex items-center gap-2 mt-5 justify-end">
            <Button variant="outline" size="sm" @click="showCreateModal = false">Cancel</Button>
            <Button size="sm" :disabled="creating" @click="handleCreate">{{ creating ? 'Creating...' : 'Create' }}</Button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import {
  ClipboardCheck, Clock, AlertTriangle, Package, Plus, RotateCw, AlertCircle,
  ChevronLeft, Search, ListFilter, CircleCheck,
} from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const API = "";

// === View state ===
const VIEW_DASHBOARD = "dashboard";
const VIEW_COUNTING = "counting";
const view = ref(VIEW_DASHBOARD);

// === Dashboard state ===
const stockTakings = ref([]);
const loading = ref(false);
const error = ref(null);
const showCreateModal = ref(false);
const newName = ref("");
const creating = ref(false);
const createError = ref(null);

// === Counting state ===
const countingId = ref(0);
const countingName = ref("");
const countingCreatedBy = ref("");
const countingStarted = ref("");
const countingStatus = ref("");
const items = ref([]);
const countingLoading = ref(false);
const countingError = ref(null);
const showQuantityAndVariance = ref(false);
const completeStockPermission = ref(false);
const searchQuery = ref("");
const filterVariancesOnly = ref(false);
let websocket = null;
let updateTimers = new Map(); // product_id -> setTimeout

// === Metrics ===
const metricCompleted = computed(() => stockTakings.value.filter(s => s.status === "Completed").length);
const metricInProgress = computed(() => stockTakings.value.filter(s => s.status === "In Progress" || s.status === "InProgress").length);
const metricDiscrepancies = computed(() => {
  // Approximate from completed sessions — relies on items being loaded per session
  return stockTakings.value.length;
});
const metricCounted = computed(() => stockTakings.value.length);

// === Computed items (categorized + filtered) ===
const filteredItems = computed(() => {
  let list = items.value;

  // Search filter
  const q = searchQuery.value.trim().toLowerCase();
  if (q) {
    list = list.filter(i =>
      (i.product_name || "").toLowerCase().includes(q) ||
      (i.manufacturer || "").toLowerCase().includes(q)
    );
  }

  // Variance filter
  if (filterVariancesOnly.value) {
    list = list.filter(i => variance(i) !== 0);
  }

  // Group by category
  const groups = {};
  for (const item of list) {
    const cat = item.category || "Uncategorized";
    if (!groups[cat]) groups[cat] = [];
    groups[cat].push(item);
  }

  const result = [];
  for (const [category, catItems] of Object.entries(groups)) {
    const catVariance = catItems.reduce((s, i) => s + variance(i), 0);
    result.push({ _isCategory: true, category, count: catItems.length, totalVariance: catVariance });
    result.push(...catItems);
  }
  return result;
});

function variance(item) {
  return (item.dispensary_count || 0) + (item.store_count || 0) - (item.snapshot_quantity || 0);
}
function hasVariance(item) { return variance(item) !== 0; }
function num(v) { return Number(v || 0); }

const totalVariance = computed(() => {
  return items.value.reduce((s, i) => s + variance(i), 0);
});
const totalIssues = computed(() => {
  return items.value.filter(i => variance(i) !== 0).length;
});

// === Dashboard functions ===
async function fetchStockTakings() {
  loading.value = true;
  error.value = null;
  try {
    const res = await fetch(`${API}/stock-taking/api/list`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    const data = await res.json();
    stockTakings.value = data.stock_takings || [];
  } catch (e) {
    error.value = e.message || "Failed to load";
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  const name = newName.value.trim();
  if (!name) return;
  creating.value = true;
  createError.value = null;
  try {
    const res = await fetch(`${API}/stock-taking/api/create`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.message || "Failed to create");
    showCreateModal.value = false;
    newName.value = "";
    openStockTaking(data.id);
  } catch (e) {
    createError.value = e.message;
  } finally {
    creating.value = false;
  }
}

function openStockTaking(id) {
  view.value = VIEW_COUNTING;
  countingId.value = id;
  loadStockTaking(id);
}

function closeStockTaking() {
  view.value = VIEW_DASHBOARD;
  countingId.value = 0;
  countingName.value = "";
  items.value = [];
  searchQuery.value = "";
  filterVariancesOnly.value = false;
  closeWebSocket();
}

// === Counting functions ===
async function loadStockTaking(id) {
  countingLoading.value = true;
  countingError.value = null;
  try {
    const res = await fetch(`${API}/stock-taking/api/${id}`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    const data = await res.json();
    const st = data.stock_taking_data;
    countingName.value = st.name;
    countingCreatedBy.value = st.created_by;
    countingStarted.value = st.started_at;
    countingStatus.value = st.status;
    items.value = data.items || [];
    const perms = data.permissions || {};
    showQuantityAndVariance.value = perms["stock:view"] === true;
    completeStockPermission.value = perms["stock:complete"] === true;
    filterVariancesOnly.value = false;
    if (countingStatus.value !== "Completed") {
      initWebSocket(id);
    }
  } catch (e) {
    countingError.value = e.message || "Failed to load";
  } finally {
    countingLoading.value = false;
  }
}

function toggleVarianceFilter() {
  filterVariancesOnly.value = !filterVariancesOnly.value;
  // If the filter would show nothing, keep it off (like old UI)
  if (filterVariancesOnly.value) {
    const hasAny = items.value.some(i => variance(i) !== 0);
    if (!hasAny) {
      filterVariancesOnly.value = false;
    }
  }
}

// === Debounced update via API ===
function queueUpdate(item) {
  const pid = item.product_id;
  if (updateTimers.has(pid)) clearTimeout(updateTimers.get(pid));
  updateTimers.set(pid, setTimeout(() => sendUpdate(item), 800));
}

async function sendUpdate(item) {
  try {
    const res = await fetch(`${API}/stock-taking/api/${countingId.value}/item/${item.product_id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        dispensary_count: item.dispensary_count,
        store_count: item.store_count,
        updated_expiry: null,
        notes: item.notes || "",
      }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      console.error("Update failed:", err);
    }
  } catch (e) {
    console.error("Update error:", e);
  }
}

async function completeStockTaking() {
  if (countingStatus.value === "Completed") return;
  try {
    const res = await fetch(`${API}/stock-taking/api/${countingId.value}`, { method: "POST" });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || "Failed to complete");
    }
    countingStatus.value = "Completed";
    closeWebSocket();
  } catch (e) {
    alert(e.message);
  }
}

// === WebSocket ===
function initWebSocket(id) {
  closeWebSocket();
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsUrl = `${protocol}//${window.location.host}/ws?stockTakingId=${id}`;
  try {
    websocket = new WebSocket(wsUrl);
    websocket.onopen = () => {};
    websocket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === "stock_item_update") {
          const server = msg.data;
          const local = items.value.find(i => i.product_id === server.product_id);
          if (local) {
            Object.assign(local, server);
          }
        } else if (msg.type === "stock_taking_complete") {
          countingStatus.value = "Completed";
          closeWebSocket();
        }
      } catch {}
    };
    websocket.onclose = () => {
      if (countingStatus.value !== "Completed") {
        setTimeout(() => initWebSocket(id), 3000);
      }
    };
    websocket.onerror = () => {};
  } catch {}
}

function closeWebSocket() {
  if (websocket) {
    websocket.close();
    websocket = null;
  }
}

// === Formatting ===
function formatDate(d) {
  if (!d) return "—";
  try { return new Date(d).toLocaleDateString("en-US", { year: "numeric", month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" }); }
  catch { return "—"; }
}
function formatExpiry(d) {
  if (!d) return "—";
  try { return new Date(d).toLocaleDateString("en-US", { month: "short", year: "numeric" }); }
  catch { return "—"; }
}

function statusBadge(status) {
  const base = "inline-block px-2 py-0.5 text-[10px] font-semibold rounded-full";
  switch ((status || "").toLowerCase()) {
    case "in progress":
    case "inprogress":
      return `${base} bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300`;
    case "completed":
      return `${base} bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300`;
    default:
      return `${base} bg-muted/50 text-muted-foreground`;
  }
}

function varianceClass(item) {
  const v = variance(item);
  if (v > 0) return "text-emerald-600 font-medium";
  if (v < 0) return "text-destructive font-medium";
  return "text-muted-foreground";
}

// === Lifecycle ===
const route = useRoute();

onMounted(() => {
  // If navigated to /stock-taking/:id, open that session directly
  const idParam = route.params.id;
  if (idParam) {
    const id = Number(idParam);
    if (id > 0) {
      openStockTaking(id);
      return;
    }
  }
  fetchStockTakings();
});

onUnmounted(() => {
  closeWebSocket();
  for (const t of updateTimers.values()) clearTimeout(t);
  updateTimers.clear();
});
</script>

<style scoped>
input.no-spinners::-webkit-outer-spin-button,
input.no-spinners::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
input.no-spinners[type="number"] { -moz-appearance: textfield; }
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
</style>