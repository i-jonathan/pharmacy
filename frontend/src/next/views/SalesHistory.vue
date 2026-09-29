<template>
  <div class="flex min-h-[calc(100dvh-3.5rem)] flex-col lg:h-[calc(100dvh-3.5rem)] lg:min-h-0 lg:flex-row">
    <!-- Main table area -->
    <div class="flex-1 flex flex-col overflow-hidden" :class="{ 'border-r border-border': detailSale }">
      <!-- Header -->
      <div class="px-4 pb-0 pt-5 sm:px-6">
        <h1 class="text-2xl font-bold text-foreground">Sales History</h1>
        <p class="text-sm text-muted-foreground mt-1">Browse and search past transactions</p>
      </div>

      <!-- Filters -->
      <div class="flex flex-col gap-3 px-4 pb-4 pt-4 sm:flex-row sm:px-6">
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
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring w-36"
          />
          <input
            v-model="dateEnd"
            type="date"
            class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring w-36"
          />
          <Button variant="outline" size="icon" :disabled="loading" @click="fetchSales">
            <RotateCw :size="16" :class="{ 'animate-spin': loading }" />
          </Button>
        </div>
      </div>

      <!-- Summary bar -->
      <div class="flex items-center justify-between px-4 pb-3 text-sm sm:px-6">
        <span class="text-muted-foreground">
          <template v-if="!loading">{{ totalItems }} sale{{ totalItems !== 1 ? 's' : '' }}</template>
        </span>
        <span v-if="hasTotalView && !loading" class="font-semibold text-foreground">
          Total: &#8358;{{ periodTotal.toLocaleString() }}
        </span>
      </div>

      <!-- Scrollable table area -->
      <div class="min-h-0 flex-1 overflow-y-auto px-4 pb-4 sm:px-6" ref="tableContainerRef" @keydown="onTableKeydown">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-24 text-muted-foreground">
          <RotateCw :size="20" class="animate-spin mr-3" />
          <span class="text-sm">Loading sales...</span>
        </div>

        <!-- Error -->
        <div v-else-if="error" class="flex flex-col items-center justify-center py-24 text-center">
          <AlertCircle :size="40" class="text-destructive/40 mb-3" />
          <p class="text-sm text-muted-foreground mb-3">{{ error }}</p>
          <Button variant="outline" size="sm" @click="fetchSales">Retry</Button>
        </div>

        <!-- Empty -->
        <div v-else-if="paginatedSales.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
          <History :size="40" class="text-muted-foreground/40 mb-3" />
          <h3 class="text-base font-semibold text-foreground mb-1">No sales found</h3>
          <p class="text-sm text-muted-foreground max-w-sm">
            {{ searchQuery ? 'No sales match your search.' : 'No sales in this date range.' }}
          </p>
        </div>

        <!-- Sales Table -->
        <div v-else class="border border-border rounded-lg overflow-hidden">
          <table class="w-full text-sm" role="grid">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-28">Receipt</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Date</th>
                <th class="text-center text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-16">Items</th>
                <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-28">Total</th>
                <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3 w-28">Cashier</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr
                v-for="(sale, i) in paginatedSales"
                :key="sale.id"
                :ref="(el) => rowRefs[i] = el"
                class="cursor-pointer transition-colors outline-none"
                :class="rowClass(i)"
                :data-index="i"
                tabindex="0"
                @click="selectAndShow(i)"
                @keydown.enter.prevent="selectAndShow(i)"
                @keydown.space.prevent="selectAndShow(i)"
              >
                <td class="px-4 py-3 font-mono text-xs font-medium text-foreground">{{ sale.receipt_number }}</td>
                <td class="px-4 py-3 text-muted-foreground text-xs whitespace-nowrap">{{ formatDate(sale.created_at) }}</td>
                <td class="px-4 py-3 text-center text-muted-foreground">{{ sale.items?.length || 0 }}</td>
                <td class="px-4 py-3 text-right font-semibold text-foreground">&#8358;{{ Number(sale.total).toLocaleString() }}</td>
                <td class="px-4 py-3 text-muted-foreground text-xs">{{ sale.cashier }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="flex items-center justify-between pt-4">
          <div v-if="totalPages > 1" class="flex items-center gap-4">
            <span class="text-xs text-muted-foreground">Page {{ currentPage }} of {{ totalPages }}</span>
            <div class="flex gap-1">
              <Button variant="outline" size="sm" :disabled="currentPage <= 1" @click="goPage(currentPage - 1)">
                <ChevronLeft :size="14" />
              </Button>
              <Button
                v-for="p in visiblePages"
                :key="p"
                variant="outline"
                size="sm"
                :class="{ 'bg-primary/10 text-primary border-primary/30': p === currentPage }"
                @click="goPage(p)"
              >{{ p }}</Button>
              <Button variant="outline" size="sm" :disabled="currentPage >= totalPages" @click="goPage(currentPage + 1)">
                <ChevronRight :size="14" />
              </Button>
            </div>
          </div>
          <div class="flex-1"></div>
          <select
            v-model.number="perPage"
            class="text-xs border border-border rounded bg-background text-foreground outline-none px-2 py-1"
            @change="goPage(1)"
          >
            <option :value="10">10 / page</option>
            <option :value="20">20 / page</option>
            <option :value="50">50 / page</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Detail Sidebar (CartPanel-style layout) -->
    <Transition name="slide-panel">
      <div
        v-if="detailSale"
        class="flex w-full max-w-full flex-col border-t border-border bg-card lg:h-full lg:w-[40%] lg:min-w-[360px] lg:border-l lg:border-t-0 lg:flex-shrink-0"
      >
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-border">
          <div>
            <h2 class="text-sm font-semibold">Sale Details</h2>
            <p class="text-xs text-muted-foreground font-mono">{{ detailSale.receipt_number }}</p>
          </div>
          <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" @click="closeDetail">
            <X :size="15" />
          </Button>
        </div>

        <!-- Customer / Sale Info -->
        <div class="px-4 py-3 border-b border-border">
          <div class="flex items-center gap-2 text-xs text-muted-foreground mb-1">
            <span class="font-medium text-foreground">{{ detailSale.cashier }}</span>
            <span>&middot;</span>
            <span>{{ formatDate(detailSale.created_at) }}</span>
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
                  <TableHead class="w-16 text-[10px] text-right">Price</TableHead>
                  <TableHead class="w-14 text-[10px] text-right">Total</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in detailSale.items" :key="item.id">
                  <TableCell class="py-1.5">
                    <div class="text-xs font-medium">{{ item.product_name }}</div>
                    <div v-if="item.manufacturer" class="text-[10px] text-muted-foreground">{{ item.manufacturer }}</div>
                  </TableCell>
                  <TableCell class="py-1.5 text-center text-xs text-muted-foreground">{{ item.quantity }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs text-muted-foreground">&#8358;{{ Number(item.unit_price).toLocaleString() }}</TableCell>
                  <TableCell class="py-1.5 text-right text-xs font-medium">&#8358;{{ lineTotal(item).toLocaleString() }}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>

          <!-- Totals + Payments side by side -->
          <div class="px-4 py-3 border-t border-border flex gap-3">
            <div class="flex-1 min-w-0">
              <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Summary</div>
              <div class="space-y-2 bg-muted/50 rounded-sm px-3 py-3">
                <div class="flex justify-between text-xs">
                  <span class="text-muted-foreground">Subtotal</span>
                  <span class="font-medium">&#8358;{{ Number(detailSale.subtotal).toLocaleString() }}</span>
                </div>
                <div v-if="Number(detailSale.discount) > 0" class="flex justify-between text-xs">
                  <span class="text-muted-foreground">Discount</span>
                  <span class="text-red-600 dark:text-red-400">-&#8358;{{ Number(detailSale.discount).toLocaleString() }}</span>
                </div>
                <div class="flex justify-between font-bold text-sm pt-2 border-t border-border">
                  <span>Total</span>
                  <span>&#8358;{{ Number(detailSale.total).toLocaleString() }}</span>
                </div>
              </div>
            </div>
            <div v-if="detailSale.payments?.length" class="flex-1 min-w-0">
              <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Payments</div>
              <div class="border border-border rounded-sm divide-y divide-border">
                <div v-for="pm in detailSale.payments" :key="pm.method_name" class="flex items-center justify-between px-3 py-2">
                  <span class="text-xs capitalize text-muted-foreground">{{ pm.method_name }}</span>
                  <span class="text-xs font-medium">&#8358;{{ Number(pm.amount).toLocaleString() }}</span>
                </div>
                <div class="flex items-center justify-between px-3 py-2 border-t border-border bg-muted/20 font-medium text-xs">
                  <span>Total Paid</span>
                  <span>&#8358;{{ totalPaid.toLocaleString() }}</span>
                </div>
                <div class="flex items-center justify-between px-3 py-2 text-xs">
                  <span class="text-muted-foreground">Change</span>
                  <span class="font-bold text-emerald-600">&#8358;{{ change.toLocaleString() }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Returns -->
          <div v-if="detailSale.returns?.length" class="px-4 py-3 border-t border-border">
            <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Returned Items</div>
            <div class="border border-border rounded-sm overflow-hidden">
              <table class="w-full text-[11px]">
                <thead>
                  <tr class="bg-muted/30 border-b border-border">
                    <th class="text-left font-semibold text-muted-foreground uppercase px-2 py-1.5">Item</th>
                    <th class="text-center font-semibold text-muted-foreground uppercase px-2 py-1.5 w-10">Qty</th>
                    <th class="text-right font-semibold text-muted-foreground uppercase px-2 py-1.5 w-14">Refund</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-border/50">
<<<<<<< Updated upstream
                  <tr v-for="ret in detailSale.returns" :key="ret.product_name + ret.quantity">
=======
                  <tr v-for="ret in detailSale.returns" :key="ret.sale_item_id">
>>>>>>> Stashed changes
                    <td class="px-2 py-1.5">
                      <div class="text-xs font-medium">{{ ret.product_name }}</div>
                      <div v-if="ret.manufacturer" class="text-[10px] text-muted-foreground">{{ ret.manufacturer }}</div>
                    </td>
                    <td class="px-2 py-1.5 text-center text-xs text-muted-foreground">{{ ret.quantity }}</td>
<<<<<<< Updated upstream
                    <td class="px-2 py-1.5 text-right text-xs font-medium">&#8358;{{ (ret.unit_price * ret.quantity).toLocaleString() }}</td>
=======
                    <td class="px-2 py-1.5 text-right text-xs font-medium">&#8358;{{ returnedRefund(ret).toLocaleString() }}</td>
>>>>>>> Stashed changes
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
<<<<<<< Updated upstream
        </div>
      </div>
    </Transition>
=======

          <div class="border-t border-border px-4 py-4">
            <Button
              v-if="hasReturnableItems"
              class="w-full sm:w-auto"
              variant="outline"
              @click="openReturnDialog"
            >
              <Undo2 :size="15" class="mr-2" />
              Return items
            </Button>
            <p v-else class="text-xs text-muted-foreground">All items on this sale have been returned.</p>
          </div>
        </div>
      </div>
    </Transition>

    <div
      v-if="returnDialogOpen && detailSale"
      class="fixed inset-0 z-[70] flex items-end justify-center bg-neutral-950/50 p-0 backdrop-blur-[2px] sm:items-center sm:p-5"
      @click.self="closeReturnDialog"
      @keydown.esc="closeReturnDialog"
    >
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="return-dialog-title"
        class="flex max-h-[92dvh] w-full max-w-3xl flex-col overflow-hidden rounded-t-2xl border border-border bg-card shadow-2xl sm:rounded-2xl"
      >
        <header class="flex items-center justify-between border-b border-border px-5 py-4 sm:px-6">
          <div>
            <h2 id="return-dialog-title" class="text-base font-semibold text-foreground">Return sold items</h2>
            <p class="mt-0.5 text-xs text-muted-foreground">Receipt {{ detailSale.receipt_number }}</p>
          </div>
          <Button variant="ghost" size="icon" aria-label="Close return dialog" :disabled="returnSubmitting" @click="closeReturnDialog">
            <X :size="16" />
          </Button>
        </header>

        <div class="min-h-0 flex-1 overflow-y-auto p-4 sm:p-6">
          <p class="mb-4 rounded-lg bg-muted/50 px-3 py-2.5 text-xs leading-5 text-muted-foreground">
            Enter the quantity being returned. The refund preview includes the original line discount. Returned quantities cannot be returned again.
          </p>
          <div class="overflow-x-auto rounded-lg border border-border">
            <table class="w-full min-w-[560px] text-sm">
              <thead class="bg-muted/40">
                <tr>
                  <th class="px-3 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Item</th>
                  <th class="w-20 px-3 py-2.5 text-right text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Sold</th>
                  <th class="w-24 px-3 py-2.5 text-right text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Returned</th>
                  <th class="w-32 px-3 py-2.5 text-center text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Return qty</th>
                  <th class="w-28 px-3 py-2.5 text-right text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Refund</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-border">
                <tr v-for="item in detailSale.items" :key="item.id">
                  <td class="px-3 py-3">
                    <p class="font-medium text-foreground">{{ item.product_name }}</p>
                    <p v-if="item.manufacturer" class="mt-0.5 text-xs text-muted-foreground">{{ item.manufacturer }}</p>
                  </td>
                  <td class="px-3 py-3 text-right text-muted-foreground">{{ item.quantity }}</td>
                  <td class="px-3 py-3 text-right text-muted-foreground">{{ item.returned_quantity || 0 }}</td>
                  <td class="px-3 py-2 text-center">
                    <input
                      v-model.number="returnQuantities[item.id]"
                      type="number"
                      min="0"
                      :max="Math.max(0, item.quantity - (item.returned_quantity || 0))"
                      :disabled="Math.max(0, item.quantity - (item.returned_quantity || 0)) === 0 || returnSubmitting"
                      :aria-label="`Quantity to return for ${item.product_name}`"
                      class="no-spinners h-9 w-20 rounded-md border border-border bg-background px-2 text-right text-sm text-foreground outline-none focus:ring-2 focus:ring-ring disabled:opacity-50"
                    />
                  </td>
                  <td class="px-3 py-3 text-right font-medium text-foreground">₦{{ returnRefundAmount(item, Number(returnQuantities[item.id] || 0)).toLocaleString() }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <label class="mt-4 block">
            <span class="text-xs font-medium text-foreground">Return note <span class="font-normal text-muted-foreground">(optional)</span></span>
            <textarea
              v-model="returnNotes"
              rows="2"
              maxlength="500"
              placeholder="Add a reason or note for the transaction"
              class="mt-1.5 w-full resize-y rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none placeholder:text-muted-foreground focus:ring-2 focus:ring-ring"
            ></textarea>
          </label>

          <p v-if="returnError" role="alert" class="mt-3 rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{{ returnError }}</p>
        </div>

        <footer class="flex flex-col gap-3 border-t border-border bg-muted/20 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-6">
          <div>
            <p class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Refund total</p>
            <p class="text-xl font-bold text-foreground">₦{{ returnRefundTotal.toLocaleString() }}</p>
          </div>
          <div class="flex flex-col-reverse gap-2 sm:flex-row">
            <Button variant="outline" :disabled="returnSubmitting" @click="closeReturnDialog">Cancel</Button>
            <Button :disabled="!hasSelectedReturnItems || returnSubmitting" @click="submitReturn">
              <RotateCw v-if="returnSubmitting" :size="15" class="mr-2 animate-spin" />
              {{ returnSubmitting ? 'Processing return…' : 'Confirm return' }}
            </Button>
          </div>
        </footer>
      </section>
    </div>

    <div v-if="toast" class="fixed bottom-5 right-5 z-[80] rounded-xl border border-border bg-card px-4 py-3 text-sm font-medium text-foreground shadow-xl" role="status">
      {{ toast }}
    </div>
>>>>>>> Stashed changes
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from "vue";
<<<<<<< Updated upstream
import { Search, RotateCw, AlertCircle, History, ChevronLeft, ChevronRight, X } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
=======
import { Search, RotateCw, AlertCircle, History, ChevronLeft, ChevronRight, X, Undo2 } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { csrfFetch } from "../lib/csrf.js";
>>>>>>> Stashed changes
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
const sales = ref([]);
const loading = ref(false);
const error = ref(null);
const searchQuery = ref("");
const rangePreset = ref("today");
const dateStart = ref("");
const dateEnd = ref("");
const detailSale = ref(null);
const selectedIndex = ref(-1);
const hasTotalView = ref(false);
const currentPage = ref(1);
const perPage = ref(20);

const rowRefs = ref([]);
const tableContainerRef = ref(null);
<<<<<<< Updated upstream
=======
const returnDialogOpen = ref(false);
const returnSubmitting = ref(false);
const returnQuantities = ref({});
const returnNotes = ref("");
const returnError = ref("");
const toast = ref("");
let toastTimer = null;
>>>>>>> Stashed changes

const totalCount = ref(0);
const totalItems = computed(() => totalCount.value);
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / perPage.value)));
<<<<<<< Updated upstream
=======
const hasReturnableItems = computed(() => detailSale.value?.items?.some((item) => Number(item.returned_quantity || 0) < Number(item.quantity)) ?? false);
const selectedReturnItems = computed(() => detailSale.value?.items?.filter((item) => Number(returnQuantities.value[item.id] || 0) > 0) ?? []);
const hasSelectedReturnItems = computed(() => selectedReturnItems.value.length > 0 && selectedReturnItems.value.every((item) => {
  const quantity = Number(returnQuantities.value[item.id]);
  return Number.isInteger(quantity) && quantity > 0 && quantity <= Number(item.quantity) - Number(item.returned_quantity || 0);
}));
const returnRefundTotal = computed(() => selectedReturnItems.value.reduce((sum, item) => sum + returnRefundAmount(item, Number(returnQuantities.value[item.id] || 0)), 0));
>>>>>>> Stashed changes

// Backend returns already-paginated data; client-side search filters within current page
const filteredSales = computed(() => {
  if (!searchQuery.value) return sales.value;
  const q = searchQuery.value.toLowerCase();
  return sales.value.filter(
    (s) =>
      s.receipt_number?.toLowerCase().includes(q) ||
      s.cashier?.toLowerCase().includes(q)
  );
});

const paginatedSales = computed(() => filteredSales.value);

const visiblePages = computed(() => {
  const total = totalPages.value;
  const cur = currentPage.value;
  if (total <= 5) return Array.from({ length: total }, (_, i) => i + 1);
  if (cur <= 3) return [1, 2, 3, 4, 5];
  if (cur >= total - 2) return [total - 4, total - 3, total - 2, total - 1, total];
  return [cur - 2, cur - 1, cur, cur + 1, cur + 2];
});

const periodTotal = ref(0);

const totalPaid = computed(() => {
  if (!detailSale.value?.payments) return 0;
  return detailSale.value.payments.reduce((sum, p) => sum + Number(p.amount), 0);
});

const change = computed(() => {
  if (!detailSale.value) return 0;
  return Math.max(0, totalPaid.value - Number(detailSale.value.total));
});

function lineTotal(item) {
  return Math.max(0, (item.unit_price * item.quantity) - (item.discount || 0));
}

<<<<<<< Updated upstream
=======
function returnRefundAmount(item, quantity, previouslyReturned = Number(item.returned_quantity || 0)) {
  if (!item || item.quantity <= 0 || quantity <= 0) return 0;
  const lineTotalKobo = Math.max(0, Math.round((Number(item.unit_price) * Number(item.quantity) - Number(item.discount || 0)) * 100));
  const soldQuantity = Number(item.quantity);
  const before = Math.floor(lineTotalKobo * previouslyReturned / soldQuantity);
  const after = Math.floor(lineTotalKobo * (previouslyReturned + quantity) / soldQuantity);
  return (after - before) / 100;
}

function returnedRefund(returnedItem) {
  const item = detailSale.value?.items?.find((saleItem) => saleItem.id === returnedItem.sale_item_id);
  return item ? returnRefundAmount(item, Number(returnedItem.quantity), 0) : 0;
}

>>>>>>> Stashed changes
function rowClass(i) {
  return [
    "hover:bg-muted/20 focus:ring-1 focus:ring-ring focus:ring-inset",
    selectedIndex.value === i
      ? "bg-primary/5 ring-1 ring-inset ring-primary/20 font-medium"
      : "",
  ];
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
  fetchSales();
}

async function fetchSales() {
  loading.value = true;
  error.value = null;
  detailSale.value = null;
  selectedIndex.value = -1;
  currentPage.value = 1;
  try {
    const params = new URLSearchParams();
    if (dateStart.value) params.set("start", dateStart.value);
    if (dateEnd.value) params.set("end", dateEnd.value);
    params.set("page", String(currentPage.value));
    params.set("per_page", String(perPage.value));
    const res = await fetch(`${API}/sales/filter?${params}`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    const data = await res.json();
    sales.value = data.data || [];
    totalCount.value = data.total_count || 0;
    if (data.total !== undefined && data.total !== null) {
      hasTotalView.value = true;
      periodTotal.value = data.total;
    }
  } catch (e) {
    error.value = e.message || "Failed to load";
  } finally {
    loading.value = false;
  }
}

async function goPage(p) {
  if (p < 1 || p > totalPages.value) return;
  currentPage.value = p;
  detailSale.value = null;
  selectedIndex.value = -1;
  loading.value = true;
  try {
    const params = new URLSearchParams();
    if (dateStart.value) params.set("start", dateStart.value);
    if (dateEnd.value) params.set("end", dateEnd.value);
    params.set("page", String(p));
    params.set("per_page", String(perPage.value));
    const res = await fetch(`${API}/sales/filter?${params}`);
    if (!res.ok) throw new Error(`Server error (${res.status})`);
    const data = await res.json();
    sales.value = data.data || [];
    totalCount.value = data.total_count || 0;
    if (data.total !== undefined && data.total !== null) {
      hasTotalView.value = true;
      periodTotal.value = data.total;
    }
  } catch (e) {
    error.value = e.message || "Failed to load page";
  } finally {
    loading.value = false;
  }
}

function select(i) {
  if (i < 0 || i >= paginatedSales.value.length) return;
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
  const sale = filteredSales.value[i];
  if (!sale) return;
  detailSale.value = sale;
}

function closeDetail() {
  detailSale.value = null;
}

<<<<<<< Updated upstream
=======
function openReturnDialog() {
  if (!detailSale.value) return;
  returnQuantities.value = Object.fromEntries(
    detailSale.value.items.map((item) => [item.id, 0]),
  );
  returnNotes.value = "";
  returnError.value = "";
  returnDialogOpen.value = true;
}

function closeReturnDialog() {
  if (returnSubmitting.value) return;
  returnDialogOpen.value = false;
  returnError.value = "";
}

async function submitReturn() {
  if (!detailSale.value || !hasSelectedReturnItems.value || returnSubmitting.value) return;

  returnSubmitting.value = true;
  returnError.value = "";
  const saleID = detailSale.value.id;
  const page = currentPage.value;
  const payload = {
    sale_id: saleID,
    notes: returnNotes.value.trim(),
    return_items: selectedReturnItems.value.map((item) => ({
      sale_item_id: item.id,
      quantity: Number(returnQuantities.value[item.id]),
    })),
  };

  try {
    const response = await csrfFetch(`${API}/sales/returns`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const body = await response.json().catch(() => ({}));
      throw new Error(body.error || `Could not process return (${response.status})`);
    }

    returnDialogOpen.value = false;
    await goPage(page);
    const refreshedIndex = sales.value.findIndex((sale) => sale.id === saleID);
    detailSale.value = refreshedIndex >= 0 ? sales.value[refreshedIndex] : null;
    selectedIndex.value = refreshedIndex;
    toast.value = `Return processed · ₦${returnRefundTotal.value.toLocaleString()}`;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => { toast.value = ""; }, 3500);
  } catch (e) {
    returnError.value = e.message || "Failed to process return";
  } finally {
    returnSubmitting.value = false;
  }
}

>>>>>>> Stashed changes
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

// --- Keyboard navigation ---
function onTableKeydown(e) {
  if (e.key === "ArrowDown") {
    e.preventDefault();
    const next = Math.min(selectedIndex.value + 1, paginatedSales.value.length - 1);
    if (next >= 0) select(next);
    if (detailSale.value) openDetailFor(next);
  }
  if (e.key === "ArrowUp") {
    e.preventDefault();
    const prev = Math.max(selectedIndex.value - 1, 0);
    if (prev >= 0) select(prev);
    if (detailSale.value) openDetailFor(prev);
  }
  if (e.key === "Enter" || e.key === " ") {
    e.preventDefault();
    if (selectedIndex.value >= 0) {
      if (detailSale.value && filteredSales.value[selectedIndex.value]?.id === detailSale.value.id) {
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
  onRangePreset();
  // Global keyboard listener for when table doesn't have focus
  document.addEventListener("keydown", (e) => {
    if (e.key === "ArrowDown" || e.key === "ArrowUp" || e.key === "Enter" || e.key === "Escape") {
      // Only intercept when this page is visible
      // The component's root will handle it via tabindex
    }
  });
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
<<<<<<< Updated upstream
</style>
=======

input.no-spinners::-webkit-outer-spin-button,
input.no-spinners::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

input.no-spinners[type="number"] {
  -moz-appearance: textfield;
}
</style>
>>>>>>> Stashed changes
