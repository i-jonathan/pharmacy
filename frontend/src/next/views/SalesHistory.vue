<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Sales History</h1>
      <p class="text-sm text-muted-foreground mt-1">Browse and search past transactions</p>
    </div>

    <!-- Filters -->
    <div class="flex flex-col sm:flex-row gap-3 mb-6">
      <div class="relative flex-1">
        <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by receipt # or cashier..."
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
          class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring w-40"
        />
        <input
          v-model="dateEnd"
          type="date"
          class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring w-40"
        />
        <Button variant="outline" size="icon" :disabled="loading" @click="fetchSales">
          <RotateCw :size="16" :class="{ 'animate-spin': loading }" />
        </Button>
      </div>
    </div>

    <!-- Summary -->
    <div v-if="!loading && sales.length" class="mb-4 text-sm text-muted-foreground">
      {{ sales.length }} sale{{ sales.length !== 1 ? 's' : '' }}
      <span v-if="hasTotalView"> &middot; Total: &#8358;{{ totalAmount.toLocaleString() }}</span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="rounded-lg border border-border bg-card p-12 flex items-center justify-center">
      <div class="flex items-center gap-3 text-muted-foreground">
        <RotateCw :size="20" class="animate-spin" />
        <span class="text-sm">Loading sales...</span>
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
      <AlertCircle :size="48" class="text-destructive/40 mb-4" />
      <h3 class="text-lg font-semibold text-foreground mb-1">Failed to load sales</h3>
      <p class="text-sm text-muted-foreground mb-4">{{ error }}</p>
      <Button variant="outline" @click="fetchSales">Retry</Button>
    </div>

    <!-- Empty -->
    <div v-else-if="sales.length === 0" class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
      <History :size="48" class="text-muted-foreground/40 mb-4" />
      <h3 class="text-lg font-semibold text-foreground mb-1">No sales found</h3>
      <p class="text-sm text-muted-foreground max-w-sm">
        {{ searchQuery ? 'No sales match your search criteria.' : 'No sales recorded in this date range.' }}
      </p>
    </div>

    <!-- Sales Table -->
    <div v-else class="rounded-lg border border-border bg-card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border bg-muted/30">
              <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-24">Receipt</th>
              <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Date</th>
              <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-16">Items</th>
              <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-28">Total</th>
              <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-32">Cashier</th>
              <th class="w-10 px-4 py-3"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/50">
            <tr
              v-for="sale in filteredSales"
              :key="sale.id"
              class="hover:bg-muted/20 transition-colors cursor-pointer"
              @click="openDetail(sale)"
            >
              <td class="px-4 py-3 font-mono text-xs font-medium text-foreground">{{ sale.receipt_number }}</td>
              <td class="px-4 py-3 text-muted-foreground text-xs">{{ formatDate(sale.created_at) }}</td>
              <td class="px-4 py-3 text-center text-muted-foreground">{{ sale.items?.length || 0 }}</td>
              <td class="px-4 py-3 text-right font-semibold text-foreground">&#8358;{{ Number(sale.total).toLocaleString() }}</td>
              <td class="px-4 py-3 text-muted-foreground text-xs">{{ sale.cashier }}</td>
              <td class="px-4 py-3">
                <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground">
                  <ChevronRight :size="14" />
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Detail Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="detailSale"
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
          @click.self="closeDetail"
        >
          <div class="absolute inset-0 bg-black/40" />
          <div class="relative bg-background rounded-xl border border-border shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <!-- Modal Header -->
            <div class="sticky top-0 bg-background border-b border-border px-6 py-4 flex items-center justify-between z-10">
              <div>
                <h2 class="text-lg font-semibold text-foreground">Sale Details</h2>
                <p class="text-xs text-muted-foreground font-mono mt-0.5">{{ detailSale.receipt_number }}</p>
              </div>
              <Button variant="ghost" size="icon" @click="closeDetail">
                <X :size="18" />
              </Button>
            </div>

            <div class="p-6 space-y-6">
              <!-- Sale Meta -->
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
                <div>
                  <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">Date</div>
                  <div class="text-sm text-foreground">{{ formatDate(detailSale.created_at) }}</div>
                </div>
                <div>
                  <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">Cashier</div>
                  <div class="text-sm text-foreground">{{ detailSale.cashier }}</div>
                </div>
                <div>
                  <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">Subtotal</div>
                  <div class="text-sm text-foreground">&#8358;{{ Number(detailSale.subtotal).toLocaleString() }}</div>
                </div>
                <div>
                  <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">Total</div>
                  <div class="text-lg font-bold text-foreground">&#8358;{{ Number(detailSale.total).toLocaleString() }}</div>
                </div>
              </div>

              <!-- Items Table -->
              <div>
                <h3 class="text-sm font-semibold text-foreground mb-3">Items</h3>
                <div class="border border-border rounded-lg overflow-hidden">
                  <table class="w-full text-sm">
                    <thead>
                      <tr class="bg-muted/30 border-b border-border">
                        <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5">Item</th>
                        <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5 w-16">Qty</th>
                        <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5 w-24">Unit Price</th>
                        <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5 w-20">Disc.</th>
                        <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5 w-24">Total</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-border/50">
                      <tr v-for="item in detailSale.items" :key="item.id">
                        <td class="px-4 py-2.5">
                          <div class="text-sm font-medium text-foreground">{{ item.product_name }}</div>
                          <div v-if="item.manufacturer" class="text-xs text-muted-foreground">{{ item.manufacturer }}</div>
                        </td>
                        <td class="px-4 py-2.5 text-center text-muted-foreground">{{ item.quantity }}</td>
                        <td class="px-4 py-2.5 text-right text-muted-foreground">&#8358;{{ Number(item.unit_price).toLocaleString() }}</td>
                        <td class="px-4 py-2.5 text-right text-muted-foreground">{{ item.discount ? '&#8358;' + Number(item.discount).toLocaleString() : '&mdash;' }}</td>
                        <td class="px-4 py-2.5 text-right font-medium text-foreground">&#8358;{{ ((item.unit_price * item.quantity) - (item.discount || 0)).toLocaleString() }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- Payments -->
              <div v-if="detailSale.payments?.length">
                <h3 class="text-sm font-semibold text-foreground mb-3">Payments</h3>
                <div class="border border-border rounded-lg overflow-hidden">
                  <table class="w-full text-sm">
                    <thead>
                      <tr class="bg-muted/30 border-b border-border">
                        <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5">Method</th>
                        <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5">Amount</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-border/50">
                      <tr v-for="pm in detailSale.payments" :key="pm.method_name">
                        <td class="px-4 py-2.5 text-sm text-foreground capitalize">{{ pm.method_name }}</td>
                        <td class="px-4 py-2.5 text-right text-sm font-medium text-foreground">&#8358;{{ Number(pm.amount).toLocaleString() }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <div class="flex justify-end gap-6 mt-2 text-sm px-4">
                  <span class="text-muted-foreground">Total Paid: <strong class="text-foreground">&#8358;{{ totalPaid.toLocaleString() }}</strong></span>
                  <span class="text-muted-foreground">Change: <strong class="text-emerald-600">&#8358;{{ change.toLocaleString() }}</strong></span>
                </div>
              </div>

              <!-- Returns -->
              <div v-if="detailSale.returns?.length">
                <h3 class="text-sm font-semibold text-foreground mb-3">Returned Items</h3>
                <div class="border border-border rounded-lg overflow-hidden">
                  <table class="w-full text-sm">
                    <thead>
                      <tr class="bg-muted/30 border-b border-border">
                        <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5">Item</th>
                        <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5 w-16">Qty</th>
                        <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5 w-24">Unit Price</th>
                        <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2.5 w-24">Refund</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-border/50">
                      <tr v-for="ret in detailSale.returns" :key="ret.product_name + ret.quantity">
                        <td class="px-4 py-2.5">
                          <div class="text-sm font-medium text-foreground">{{ ret.product_name }}</div>
                          <div v-if="ret.manufacturer" class="text-xs text-muted-foreground">{{ ret.manufacturer }}</div>
                        </td>
                        <td class="px-4 py-2.5 text-center text-muted-foreground">{{ ret.quantity }}</td>
                        <td class="px-4 py-2.5 text-right text-muted-foreground">&#8358;{{ Number(ret.unit_price).toLocaleString() }}</td>
                        <td class="px-4 py-2.5 text-right font-medium text-foreground">&#8358;{{ (ret.unit_price * ret.quantity).toLocaleString() }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { Search, RotateCw, AlertCircle, History, ChevronRight, X } from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const API = "";

// State
const sales = ref([]);
const loading = ref(false);
const error = ref(null);
const searchQuery = ref("");
const rangePreset = ref("today");
const dateStart = ref("");
const dateEnd = ref("");
const detailSale = ref(null);
const hasTotalView = ref(false);

const totalAmount = computed(() => {
  return sales.value.reduce((sum, s) => sum + Number(s.total), 0);
});

const totalPaid = computed(() => {
  if (!detailSale.value?.payments) return 0;
  return detailSale.value.payments.reduce((sum, p) => sum + Number(p.amount), 0);
});

const change = computed(() => {
  if (!detailSale.value) return 0;
  return Math.max(0, totalPaid.value - Number(detailSale.value.total));
});

const filteredSales = computed(() => {
  if (!searchQuery.value) return sales.value;
  const q = searchQuery.value.toLowerCase();
  return sales.value.filter(
    (s) =>
      s.receipt_number?.toLowerCase().includes(q) ||
      s.cashier?.toLowerCase().includes(q)
  );
});

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
      const ys = `${yest.getFullYear()}-${String(yest.getMonth() + 1).padStart(2, "0")}-${String(yest.getDate()).padStart(2, "0")}`;
      return { start: ys, end: ys };
    }
    case "this-week": {
      const mon = new Date(now);
      mon.setDate(mon.getDate() - mon.getDay() + 1);
      return {
        start: `${mon.getFullYear()}-${String(mon.getMonth() + 1).padStart(2, "0")}-${String(mon.getDate()).padStart(2, "0")}`,
        end: today,
      };
    }
    case "last-week": {
      const lastMon = new Date(now);
      lastMon.setDate(lastMon.getDate() - lastMon.getDay() - 6);
      const lastSun = new Date(lastMon);
      lastSun.setDate(lastMon.getDate() + 6);
      return {
        start: `${lastMon.getFullYear()}-${String(lastMon.getMonth() + 1).padStart(2, "0")}-${String(lastMon.getDate()).padStart(2, "0")}`,
        end: `${lastSun.getFullYear()}-${String(lastSun.getMonth() + 1).padStart(2, "0")}-${String(lastSun.getDate()).padStart(2, "0")}`,
      };
    }
    case "this-month":
      return { start: `${y}-${m}-01`, end: today };
    case "last-month": {
      const lm = new Date(now.getFullYear(), now.getMonth() - 1, 1);
      const lme = new Date(now.getFullYear(), now.getMonth(), 0);
      return {
        start: `${lm.getFullYear()}-${String(lm.getMonth() + 1).padStart(2, "0")}-01`,
        end: `${lme.getFullYear()}-${String(lme.getMonth() + 1).padStart(2, "0")}-${String(lme.getDate()).padStart(2, "0")}`,
      };
    }
    default:
      return { start: "", end: "" };
  }
}

function onRangePreset() {
  const { start, end } = getDateRange(rangePreset.value);
  dateStart.value = start;
  dateEnd.value = end;
  if (start && end) fetchSales();
}

async function fetchSales() {
  loading.value = true;
  error.value = null;
  try {
    const params = new URLSearchParams();
    if (dateStart.value) params.set("start", dateStart.value);
    if (dateEnd.value) params.set("end", dateEnd.value);
    const url = `${API}/sales/filter?${params}`;
    const res = await fetch(url);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    const data = await res.json();
    sales.value = data.data || [];
    if (data.total !== undefined && data.total !== null) {
      hasTotalView.value = true;
    }
  } catch (e) {
    error.value = e.message || "Failed to load sales history";
  } finally {
    loading.value = false;
  }
}

function openDetail(sale) {
  detailSale.value = sale;
}

function closeDetail() {
  detailSale.value = null;
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
  // Load today's sales by default
  onRangePreset();
});
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>