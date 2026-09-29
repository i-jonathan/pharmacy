<template>
  <div class="flex min-h-[calc(100dvh-3.5rem)] flex-col lg:h-[calc(100dvh-3.5rem)] lg:min-h-0 lg:flex-row">
    <!-- Main table area -->
    <div class="flex-1 flex flex-col overflow-hidden" :class="{ 'border-r border-border': detailBatch }">
      <!-- Header -->
      <div class="px-4 pb-0 pt-5 sm:px-6">
        <h1 class="text-2xl font-bold text-foreground">Received Items History</h1>
        <p class="text-sm text-muted-foreground mt-1">Browse and search past inventory receipts</p>
      </div>

      <!-- Filters -->
      <div class="flex flex-col gap-3 px-4 pb-4 pt-4 sm:flex-row sm:px-6">
        <div class="relative flex-1">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by supplier..."
            class="w-full pl-9 pr-4 py-2 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
          />
        </div>
        <div class="flex gap-2 flex-wrap">
          <select
            v-model="rangePreset"
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring"
            @change="onRangePreset"
          >
            <option value="today">Today</option>
            <option value="yesterday">Yesterday</option>
            <option value="this-week">This Week</option>
            <option value="last-week">Last Week</option>
            <option value="this-month">This Month</option>
            <option value="last-month">Last Month</option>
            <option value="all">All Time</option>
          </select>
          <input
            v-model="dateStart"
            type="date"
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring w-36"
          />
          <input
            v-model="dateEnd"
            type="date"
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring w-36"
          />
          <Button variant="outline" size="icon" :disabled="loading" @click="fetchBatches">
            <RotateCw :size="16" :class="{ 'animate-spin': loading }" />
          </Button>
        </div>
      </div>

      <!-- Summary bar -->
      <div class="flex items-center justify-between px-4 pb-3 text-sm sm:px-6">
        <span class="text-muted-foreground">
          <template v-if="!loading">{{ totalItems }} receipt{{ totalItems !== 1 ? 's' : '' }}</template>
        </span>
      </div>

      <!-- Scrollable table area -->
      <div class="min-h-0 flex-1 overflow-y-auto px-4 pb-4 sm:px-6">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-24 text-muted-foreground">
          <RotateCw :size="20" class="animate-spin mr-3" />
          <span class="text-sm">Loading receipts...</span>
        </div>

        <!-- Error -->
        <div v-else-if="error" class="flex flex-col items-center justify-center py-24 text-center">
          <AlertCircle :size="40" class="text-destructive/40 mb-3" />
          <p class="text-sm text-muted-foreground mb-3">{{ error }}</p>
          <Button variant="outline" size="sm" @click="fetchBatches">Retry</Button>
        </div>

        <!-- Empty -->
        <div v-else-if="filteredBatches.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
          <ClipboardList :size="40" class="text-muted-foreground/40 mb-3" />
          <h3 class="text-base font-semibold text-foreground mb-1">No receipts found</h3>
          <p class="text-sm text-muted-foreground max-w-sm">
            {{ searchQuery ? 'No receipts match your search.' : 'No receipts in this date range.' }}
          </p>
        </div>

        <!-- Batches Table -->
        <div v-else class="border border-border rounded-lg overflow-hidden">
          <table class="w-full text-sm" role="grid">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Supplier</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Date</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Received By</th>
                <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-16">Items</th>
                <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-28">Total Cost</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr
                v-for="(batch, i) in filteredBatches"
                :key="batch.id"
                class="cursor-pointer transition-colors hover:bg-muted/20 outline-none"
                :class="{ 'bg-primary/5 ring-1 ring-inset ring-primary/20 font-medium': selectedIndex === i }"
                tabindex="0"
                @click="selectAndShow(i)"
              >
                <td class="px-4 py-3">
                  <div class="font-medium text-foreground text-sm">{{ batch.supplier_name }}</div>
                </td>
                <td class="px-4 py-3 text-muted-foreground text-xs whitespace-nowrap">{{ formatDate(batch.created_at) }}</td>
                <td class="px-4 py-3 text-muted-foreground text-xs">{{ batch.received_by }}</td>
                <td class="px-4 py-3 text-center text-muted-foreground">{{ batch.items?.length || 0 }}</td>
                <td class="px-4 py-3 text-right font-semibold text-foreground">&#8358;{{ totalCost(batch).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Detail Sidebar -->
    <Transition name="slide-panel">
      <div
        v-if="detailBatch"
        class="flex w-full max-w-full flex-col border-t border-border bg-card lg:h-full lg:w-[40%] lg:min-w-[360px] lg:border-l lg:border-t-0 lg:flex-shrink-0"
      >
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-border">
          <div>
            <h2 class="text-sm font-semibold">Receipt Details</h2>
            <p class="text-xs text-muted-foreground">{{ detailBatch.supplier_name }}</p>
          </div>
          <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" @click="closeDetail">
            <X :size="15" />
          </Button>
        </div>

        <!-- Batch Info -->
        <div class="px-4 py-3 border-b border-border">
          <div class="flex items-center gap-2 text-xs text-muted-foreground mb-1">
            <span class="font-medium text-foreground">{{ detailBatch.received_by }}</span>
            <span>&middot;</span>
            <span>{{ formatDate(detailBatch.created_at) }}</span>
          </div>
        </div>

        <!-- Scrollable content area -->
        <div class="flex-1 overflow-y-auto">
          <!-- Items Table -->
          <div class="p-3">
            <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 px-1">Items Received</div>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead class="text-[10px]">Item</TableHead>
                  <TableHead class="w-10 text-[10px] text-center">Qty</TableHead>
                  <TableHead class="w-16 text-[10px] text-right">Cost</TableHead>
                  <TableHead class="w-14 text-[10px] text-right">Total</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in detailBatch.items" :key="item.product_id">
                  <TableCell class="py-1.5">
                    <div class="text-xs font-medium">{{ item.product_name }}</div>
                    <div v-if="item.manufacturer" class="text-[10px] text-muted-foreground">{{ item.manufacturer }}</div>
                  </TableCell>
                  <TableCell class="py-1.5 text-center text-xs text-muted-foreground">{{ item.quantity }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs text-muted-foreground">&#8358;{{ Number(item.cost_price).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</TableCell>
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
                <span class="font-medium">{{ detailBatch.items?.length || 0 }}</span>
              </div>
              <div class="flex justify-between font-bold text-sm pt-2 border-t border-border">
                <span>Total Cost</span>
                <span>&#8358;{{ batchTotalCost.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { Search, RotateCw, AlertCircle, ClipboardList, X } from "lucide-vue-next";
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
const batches = ref([]);
const loading = ref(false);
const error = ref(null);
const searchQuery = ref("");
const rangePreset = ref("today");
const dateStart = ref("");
const dateEnd = ref("");
const detailBatch = ref(null);
const selectedIndex = ref(-1);

const totalCount = ref(0);
const totalItems = computed(() => totalCount.value);

const filteredBatches = computed(() => {
  if (!searchQuery.value) return batches.value;
  const q = searchQuery.value.toLowerCase();
  return batches.value.filter(
    (b) =>
      b.supplier_name?.toLowerCase().includes(q) ||
      b.received_by?.toLowerCase().includes(q)
  );
});

const batchTotalCost = computed(() => {
  const b = detailBatch.value;
  if (!b?.items) return 0;
  return b.items.reduce((sum, item) => sum + Number(item.cost_price || 0) * Number(item.quantity || 0), 0);
});

function totalCost(batch) {
  if (!batch?.items) return 0;
  return batch.items.reduce((sum, item) => sum + Number(item.cost_price || 0) * Number(item.quantity || 0), 0);
}

function lineTotal(item) {
  return Number(item.cost_price || 0) * Number(item.quantity || 0);
}

function getDateRange(preset) {
  const now = new Date();
  const y = now.getFullYear();
  const m = String(now.getMonth() + 1).padStart(2, "0");
  const d = String(now.getDate()).padStart(2, "0");
  const today = `${y}-${m}-${d}`;
  switch (preset) {
    case "today":
      return { start: today, end: today };
    case "yesterday": {
      const yest = new Date(now);
      yest.setDate(yest.getDate() - 1);
      return { start: yest.toISOString().slice(0, 10), end: yest.toISOString().slice(0, 10) };
    }
    case "this-week": {
      const sun = new Date(now);
      sun.setDate(now.getDate() - now.getDay());
      return { start: sun.toISOString().slice(0, 10), end: today };
    }
    case "last-week": {
      const start = new Date(now);
      start.setDate(now.getDate() - now.getDay() - 7);
      const end = new Date(start);
      end.setDate(start.getDate() + 6);
      return { start: start.toISOString().slice(0, 10), end: end.toISOString().slice(0, 10) };
    }
    case "this-month":
      return { start: `${y}-${m}-01`, end: today };
    case "last-month": {
      const lm = new Date(now.getFullYear(), now.getMonth() - 1, 1);
      const lme = new Date(now.getFullYear(), now.getMonth(), 0);
      return { start: lm.toISOString().slice(0, 10), end: lme.toISOString().slice(0, 10) };
    }
    default:
      return { start: "", end: "" };
  }
}

function onRangePreset() {
  const { start, end } = getDateRange(rangePreset.value);
  dateStart.value = start;
  dateEnd.value = end;
  fetchBatches();
}

async function fetchBatches() {
  loading.value = true;
  error.value = null;
  detailBatch.value = null;
  selectedIndex.value = -1;
  try {
    const params = new URLSearchParams();
    if (dateStart.value) params.set("start", dateStart.value);
    if (dateEnd.value) params.set("end", dateEnd.value);
    const res = await fetch(`${API}/inventory/received-items-history/api?${params}`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    const data = await res.json();
    batches.value = data.batches || [];
    totalCount.value = batches.value.length;
  } catch (e) {
    error.value = e.message || "Failed to load";
  } finally {
    loading.value = false;
  }
}

function selectAndShow(i) {
  selectedIndex.value = i;
  detailBatch.value = filteredBatches.value[i];
}

function closeDetail() {
  detailBatch.value = null;
  selectedIndex.value = -1;
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

onMounted(() => {
  onRangePreset();
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