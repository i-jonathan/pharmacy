<template>
  <div class="p-6 lg:p-8">
    <!-- ===== DASHBOARD ===== -->
    <template v-if="view === 'dashboard'">
      <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-6">
        <div>
          <h1 class="text-2xl font-bold text-foreground">Stock Taking</h1>
          <p class="text-sm text-muted-foreground mt-1">Conduct and manage inventory counts</p>
        </div>
        <Button size="lg" class="w-full sm:w-auto" @click="showCreateModal = true"><Plus :size="16" class="mr-2" />New Stock Count</Button>
      </div>

      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 lg:gap-4 mb-6">
        <div class="rounded-lg border border-border bg-card p-3 lg:p-4">
          <div class="flex items-center gap-1.5 mb-1">
            <ClipboardCheck :size="14" class="text-emerald-500 shrink-0" />
            <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Completed</span>
          </div>
          <div class="text-xl lg:text-2xl font-bold text-foreground">{{ metricCompleted }}</div>
        </div>
        <div class="rounded-lg border border-border bg-card p-3 lg:p-4">
          <div class="flex items-center gap-1.5 mb-1">
            <Clock :size="14" class="text-amber-500 shrink-0" />
            <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">In Progress</span>
          </div>
          <div class="text-xl lg:text-2xl font-bold text-foreground">{{ metricInProgress }}</div>
        </div>
        <div class="rounded-lg border border-border bg-card p-3 lg:p-4">
          <div class="flex items-center gap-1.5 mb-1">
            <AlertTriangle :size="14" class="text-destructive shrink-0" />
            <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Issues</span>
          </div>
          <div class="text-xl lg:text-2xl font-bold text-foreground">{{ metricDiscrepancies }}</div>
        </div>
        <div class="rounded-lg border border-border bg-card p-3 lg:p-4">
          <div class="flex items-center gap-1.5 mb-1">
            <Package :size="14" class="text-primary shrink-0" />
            <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Items</span>
          </div>
          <div class="text-xl lg:text-2xl font-bold text-foreground">{{ metricCounted }}</div>
        </div>
      </div>

      <div v-if="!loading && stockTakings.length === 0" class="rounded-lg border border-border bg-card p-8 lg:p-12 flex flex-col items-center justify-center text-center">
        <ClipboardCheck :size="36" class="text-muted-foreground/40 mb-3" />
        <h3 class="text-base font-semibold text-foreground mb-1">No stock counts yet</h3>
        <p class="text-sm text-muted-foreground max-w-sm">Start a new stock count to track inventory levels and identify discrepancies.</p>
      </div>

      <div v-else-if="loading" class="flex items-center justify-center py-12 text-muted-foreground">
        <RotateCw :size="20" class="animate-spin mr-3" />
        <span class="text-sm">Loading...</span>
      </div>

      <div v-else-if="error" class="flex flex-col items-center justify-center py-12 text-center rounded-lg border border-border bg-card">
        <AlertCircle :size="28" class="text-destructive/40 mb-2" />
        <p class="text-sm text-muted-foreground">{{ error }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetchStockTakings">Retry</Button>
      </div>

      <div v-else class="overflow-x-auto rounded-lg border border-border">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border bg-muted/30 text-xs">
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5">Name</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5">Status</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5">Created By</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5">Started</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5">Completed</th>
              <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-3 py-2.5">By</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/50">
            <tr v-for="st in stockTakings" :key="st.id" class="cursor-pointer hover:bg-muted/20 transition-colors" @click="openStockTaking(st.id)">
              <td class="px-3 py-2.5 font-medium text-foreground text-sm">{{ st.name }}</td>
              <td class="px-3 py-2.5"><span :class="statusBadge(st.status)">{{ st.status }}</span></td>
              <td class="px-3 py-2.5 text-sm text-muted-foreground">{{ st.created_by }}</td>
              <td class="px-3 py-2.5 text-sm text-muted-foreground whitespace-nowrap">{{ formatDate(st.started_at) }}</td>
              <td class="px-3 py-2.5 text-sm text-muted-foreground whitespace-nowrap">{{ st.completed_at ? formatDate(st.completed_at) : '—' }}</td>
              <td class="px-3 py-2.5 text-sm text-muted-foreground">{{ st.completed_by || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <!-- ===== STOCK TAKING COUNTING VIEW ===== -->
    <template v-if="view === 'counting'">
      <!-- Back + Header (stacks on mobile, side-by-side on desktop) -->
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-2 mb-4">
        <div class="flex items-center gap-2 sm:gap-3 w-full sm:w-auto">
          <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground shrink-0" @click="closeStockTaking"><ChevronLeft :size="16" /></Button>
          <div class="min-w-0 flex-1">
            <h1 class="text-base lg:text-xl font-bold text-foreground">{{ countingName || 'Stock Taking' }}</h1>
            <p class="text-xs text-muted-foreground">
              {{ countingCreatedBy }} · {{ formatDate(countingStarted) }} ·
              <span class="font-medium" :class="countingStatus === 'Completed' ? 'text-emerald-600' : 'text-primary'">{{ countingStatus }}</span>
            </p>
          </div>
        </div>
        <div class="flex gap-2 w-full sm:w-auto sm:ml-auto">
          <Button v-if="showQuantityAndVariance" variant="outline" size="sm" class="flex-1 sm:flex-none gap-1.5" @click="toggleVarianceFilter">
            <ListFilter :size="12" class="shrink-0" />
            <span>{{ filterVariancesOnly ? 'Show All' : 'Variances' }}</span>
          </Button>
          <Button v-if="completeStockPermission" size="sm" class="flex-1 sm:flex-none gap-1.5" :disabled="countingStatus === 'Completed'" @click="completeStockTaking">
            <CircleCheck :size="12" class="shrink-0" />
            <span>{{ countingStatus === 'Completed' ? 'Complete' : 'Complete' }}</span>
          </Button>
        </div>
      </div>

      <!-- Sticky Search -->
      <div class="sticky top-0 z-20 bg-card -mx-6 lg:-mx-8 px-6 lg:px-8 py-3 border-b border-border" style="position:sticky;top:0;">
        <div class="relative">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search items..."
            class="w-full pl-9 pr-4 py-3 text-sm border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
          />
        </div>
      </div>

      <!-- Loading / Error / Empty -->
      <div v-if="countingLoading" class="flex items-center justify-center py-12 text-muted-foreground mt-4">
        <RotateCw :size="20" class="animate-spin mr-3" />
        <span class="text-sm">Loading items...</span>
      </div>
      <div v-else-if="countingError" class="flex flex-col items-center justify-center py-12 text-center rounded-lg border border-border bg-card mt-4">
        <AlertCircle :size="28" class="text-destructive/40 mb-2" />
        <p class="text-sm text-muted-foreground">{{ countingError }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="loadStockTaking(countingId)">Retry</Button>
      </div>

      <!-- Items Table -->
      <div v-else-if="filteredItems.length" class="overflow-x-auto rounded-lg border border-border mt-4">
        <div class="max-h-[65vh] overflow-y-auto">
          <table class="w-full min-w-[700px] text-sm">
            <thead class="bg-muted/30 sticky top-0 z-10">
              <tr class="text-xs">
                <th class="text-left font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2.5 sticky left-0 bg-muted/30 z-20 min-w-[160px]">Item</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2.5 w-20">Sys</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2.5 w-24">Disp.</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2.5 w-24">Store</th>
                <th v-if="showQuantityAndVariance" class="text-right font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2.5 w-20">Var</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2.5 w-32">Expiry</th>
                <th class="text-center font-semibold text-muted-foreground uppercase tracking-wider px-2.5 py-2.5 w-28">Notes</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <template v-for="(row, ri) in filteredItems" :key="ri">
                <!-- Category header (not sticky — thead + search bar are sufficient) -->
                <tr v-if="row._isCategory" class="bg-muted/20">
                  <td colspan="7" class="px-3 py-1.5">
                    <div class="flex items-center gap-2 flex-wrap">
                      <span class="text-xs font-bold text-foreground uppercase tracking-wider">{{ row.category }}</span>
                      <span class="text-xs text-muted-foreground">· {{ row.count }} item{{ row.count !== 1 ? 's' : '' }}</span>
                      <span v-if="showQuantityAndVariance" class="text-xs" :class="row.totalVariance !== 0 ? 'text-destructive font-medium' : 'text-muted-foreground'">· Var: {{ row.totalVariance > 0 ? '+' : '' }}{{ row.totalVariance }}</span>
                    </div>
                  </td>
                </tr>
                <!-- Item row -->
                <tr v-else class="hover:bg-muted/20 transition-colors" :class="{ 'bg-destructive/5': hasVariance(row) }">
                  <td class="px-2.5 py-2.5 sticky left-0 bg-card z-10">
                    <div class="text-xs sm:text-sm font-medium text-foreground leading-tight">{{ row.product_name }}</div>
                    <div v-if="row.manufacturer" class="text-xs sm:text-xs text-muted-foreground">{{ row.manufacturer }}</div>
                    <div v-if="row.last_updated_by" class="text-xs text-muted-foreground/60">by {{ row.last_updated_by }} {{ row._timeAgoStr || '' }}</div>
                  </td>
                  <td class="px-2.5 py-2.5 text-center text-xs font-mono text-muted-foreground">{{ row.snapshot_quantity ?? '—' }}</td>
                  <td class="px-2.5 py-2.5">
                    <input
                      v-if="countingStatus !== 'Completed'"
                      :value="row.dispensary_count"
                      @input="row.dispensary_count = Math.max(0, num($event.target.value)); row._dispEntered = true; queueUpdate(row)"
                      type="number" min="0"
                      class="no-spinners w-full px-1.5 py-1.5 text-xs text-center border border-border rounded-md bg-background outline-none font-mono"
                      :class="{ 'border-ring/50 font-semibold': row.dispensary_count !== row.snapshot_quantity }"
                    />
                    <span v-else class="text-xs font-mono text-foreground block text-center">{{ row.dispensary_count ?? '—' }}</span>
                  </td>
                  <td class="px-2.5 py-2.5">
                    <input
                      v-if="countingStatus !== 'Completed'"
                      :value="row.store_count"
                      @input="row.store_count = Math.max(0, num($event.target.value)); queueUpdate(row)"
                      type="number" min="0"
                      class="no-spinners w-full px-1.5 py-1.5 text-xs text-center border border-border rounded-md bg-background outline-none font-mono"
                      :class="{ 'border-ring/50 font-semibold': row.store_count !== row.snapshot_quantity }"
                      :disabled="!row._dispEntered"
                    />
                    <span v-else class="text-xs font-mono text-foreground block text-center">{{ row.store_count ?? '—' }}</span>
                  </td>
                  <td v-if="showQuantityAndVariance" class="px-2.5 py-2.5 text-center text-xs font-mono" :class="varianceClass(row)">{{ variance(row) > 0 ? '+' : '' }}{{ variance(row) }}</td>
                  <td class="px-2.5 py-2.5">
                    <select
                      v-if="countingStatus !== 'Completed'"
                      v-model="row._expiry"
                      @change="queueUpdate(row)"
                      class="w-full text-xs text-center border border-border rounded-md bg-background outline-none px-1 py-1.5"
                    >
                      <option :value="null" selected>—</option>
                      <option v-for="d in (row.expiry_options || [])" :key="d" :value="d">{{ formatMonthYear(d) }}</option>
                    </select>
                    <span v-else class="text-xs text-center text-muted-foreground">{{ formatMonthYear(row._expiry || row.earliest_expiry) }}</span>
                  </td>
                  <td class="px-2.5 py-2.5">
                    <input
                      v-if="countingStatus !== 'Completed'"
                      :value="row.notes"
                      @input="row.notes = $event.target.value; queueUpdate(row)"
                      type="text"
                      class="no-spinners w-full px-1.5 py-1.5 text-xs border border-border rounded-md bg-background outline-none"
                      placeholder="—"
                    />
                    <span v-else class="text-xs text-muted-foreground">{{ row.notes || '—' }}</span>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>

      <div v-else class="flex flex-col items-center justify-center py-12 text-center rounded-lg border border-border bg-card mt-4">
        <ClipboardCheck :size="28" class="text-muted-foreground/40 mb-2" />
        <p class="text-sm text-muted-foreground">No items match your search.</p>
      </div>

      <!-- Summary -->
      <div v-if="!countingLoading && displayItems.length && showQuantityAndVariance" class="flex flex-col sm:flex-row items-start sm:items-center gap-1.5 sm:gap-3 px-3 py-2.5 bg-muted/30 rounded-lg border border-border text-sm mt-3">
        <span class="text-xs text-muted-foreground">Total Variance: <span class="font-semibold" :class="totalVariance !== 0 ? 'text-destructive' : 'text-foreground'">{{ totalVariance > 0 ? '+' : '' }}{{ totalVariance }}</span></span>
        <span class="text-xs text-muted-foreground">Issues: <span class="font-semibold">{{ totalIssues }}</span></span>
        <span class="text-xs text-muted-foreground">Items: <span class="font-semibold">{{ displayItems.length }}</span></span>
      </div>
    </template>

    <!-- ===== CREATE MODAL ===== -->
    <Transition name="fade">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" @click.self="showCreateModal = false">
        <div class="bg-card border border-border rounded-xl shadow-xl w-full max-w-md mx-auto p-6">
          <h2 class="text-base font-semibold mb-4">New Stock Taking</h2>
          <div v-if="createError" class="mb-3 text-xs text-destructive bg-destructive/10 rounded-md px-3 py-2.5">{{ createError }}</div>
          <label class="text-xs text-muted-foreground mb-1.5 block">Name</label>
          <input
            v-model="newName"
            @keydown.enter="handleCreate"
            type="text"
            placeholder="e.g. September 2026 Stock Count"
            class="w-full px-3 py-2.5 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
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
let rawItems = [];
const countingLoading = ref(false);
const countingError = ref(null);
const showQuantityAndVariance = ref(false);
const completeStockPermission = ref(false);
const searchQuery = ref("");
const filterVariancesOnly = ref(false);
let websocket = null;
let updateTimers = new Map();
let timeAgoTimer = null;
const timeTick = ref(0);

// === Enriched items (with _isCategory, _expiry, _dispEntered) ===
const displayItems = ref([]);

// === Rebuild enriched items whenever raw items change ===
function rebuildItems() {
  const items = [];
  for (const item of rawItems) {
    if (item._expiry === undefined) {
      item._expiry = item.earliest_expiry || null;
    }
    if (item._dispEntered === undefined) {
      item._dispEntered = item.dispensary_count !== null && item.dispensary_count !== undefined && item.dispensary_count !== "";
    }
    item._timeAgoStr = timeAgo(item.last_updated_at);
    items.push(item);
  }
  displayItems.value = items;
}

function refreshTimeAgo() {
  for (const item of rawItems) {
    item._timeAgoStr = timeAgo(item.last_updated_at);
  }
  // Force Vue to re-render by cloning the array reference
  displayItems.value = displayItems.value.slice();
}

function startTimeAgoTimer() {
  stopTimeAgoTimer();
  timeAgoTimer = setInterval(() => {
    timeTick.value = timeTick.value + 1;
    refreshTimeAgo();
  }, 10000);
}

function stopTimeAgoTimer() {
  if (timeAgoTimer) {
    clearInterval(timeAgoTimer);
    timeAgoTimer = null;
  }
}

// === Metrics ===
const metricCompleted = computed(() => stockTakings.value.filter(s => s.status === "Completed" || s.status === "completed").length);
const metricInProgress = computed(() => stockTakings.value.filter(s => {
  const st = s.status.toLowerCase();
  return st === "in progress" || st === "inprogress";
}).length);
const metricDiscrepancies = computed(() => displayItems.value.filter(i => variance(i) !== 0).length);
const metricCounted = computed(() => displayItems.value.length);

// === Computed items (filtered + categorized) ===
const filteredItems = computed(() => {
  let list = displayItems.value;
  const q = searchQuery.value.trim().toLowerCase();
  if (q) {
    list = list.filter(i =>
      (i.product_name || "").toLowerCase().includes(q) ||
      (i.manufacturer || "").toLowerCase().includes(q)
    );
  }
  if (filterVariancesOnly.value) {
    list = list.filter(i => variance(i) !== 0);
  }
  const groups = {};
  for (const item of list) {
    const cat = item.category || "Uncategorized";
    if (!groups[cat]) groups[cat] = [];
    groups[cat].push(item);
  }
  const result = [];
  for (const [category, catItems] of Object.entries(groups)) {
    const catVariance = catItems.reduce((s, i) => s + varianceRaw(
      i.dispensary_count, i.store_count, i.snapshot_quantity
    ), 0);
    result.push({ _isCategory: true, category, count: catItems.length, totalVariance: catVariance });
    result.push(...catItems);
  }
  return result;
});

function varianceRaw(disp, store, snap) {
  return (disp || 0) + (store || 0) - (snap || 0);
}
function variance(item) {
  return varianceRaw(item.dispensary_count, item.store_count, item.snapshot_quantity);
}
function hasVariance(item) { return variance(item) !== 0; }
function num(v) { return Number(v || 0); }

const totalVariance = computed(() => displayItems.value.reduce((s, i) => s + variance(i), 0));
const totalIssues = computed(() => displayItems.value.filter(i => variance(i) !== 0).length);

// === Dashboard ===
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
  rawItems = [];
  displayItems.value = [];
  searchQuery.value = "";
  filterVariancesOnly.value = false;
  stopTimeAgoTimer();
  closeWebSocket();
}

// === Counting ===
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
    rawItems = data.items || [];
    const perms = data.permissions || {};
    showQuantityAndVariance.value = perms["stock:view"] === true;
    completeStockPermission.value = perms["stock:complete"] === true;
    filterVariancesOnly.value = false;
    rebuildItems();
    startTimeAgoTimer();
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
  if (filterVariancesOnly.value) {
    const hasAny = displayItems.value.some(i => variance(i) !== 0);
    if (!hasAny) filterVariancesOnly.value = false;
  }
}

// === Debounced update ===
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
        updated_expiry: item._expiry ? item._expiry.split("T")[0] || item._expiry.slice(0, 10) : null,
        notes: item.notes || "",
      }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: `HTTP ${res.status}` }));
      console.error("Stock update failed:", err);
      // Item counts were saved server-side (before expiry validation), so
      // optimistically update the timestamp anyway.
      item.last_updated_at = new Date().toISOString();
      item._timeAgoStr = timeAgo(item.last_updated_at);
      displayItems.value = displayItems.value.slice();
    } else {
      // Optimistically set last_updated so the time-ago text updates immediately
      // rather than waiting for the WebSocket echo.
      item.last_updated_at = new Date().toISOString();
      item._timeAgoStr = timeAgo(item.last_updated_at);
      // Trigger re-render
      displayItems.value = displayItems.value.slice();
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
    stopTimeAgoTimer();
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
          const local = rawItems.find(i => i.product_id === server.product_id);
          if (local) {
            Object.assign(local, server);
            rebuildItems();
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

function formatMonthYear(d) {
  if (!d) return "—";
  try { return new Date(d).toLocaleDateString("en-US", { month: "short", year: "numeric" }); }
  catch { return "—"; }
}

function timeAgo(d) {
  if (!d) return "";
  try {
    const diff = Math.floor((new Date().getTime() - new Date(d).getTime()) / 60000);
    if (diff < 1) return "just now";
    if (diff < 60) return diff + "m ago";
    const hrs = Math.floor(diff / 60);
    if (hrs < 24) return hrs + "h ago";
    return Math.floor(hrs / 24) + "d ago";
  } catch { return ""; }
}

function statusBadge(status) {
  const base = "inline-block px-1.5 py-0.5 text-xs font-semibold rounded-full whitespace-nowrap";
  switch ((status || "").toLowerCase().replace(" ", "")) {
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
  const idParam = route.params.id;
  if (idParam) {
    const id = Number(idParam);
    if (id > 0) { openStockTaking(id); return; }
  }
  fetchStockTakings();
});

onUnmounted(() => {
  stopTimeAgoTimer();
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