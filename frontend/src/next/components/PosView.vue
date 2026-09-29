<template>
  <div class="grid h-[calc(100dvh-3.5rem)] min-h-0 grid-cols-1 grid-rows-[minmax(0,1fr)_auto] xl:grid-cols-[minmax(0,1.05fr)_minmax(28rem,0.95fr)] xl:grid-rows-1">
    <!-- Mobile keeps one workspace panel visible; desktop shows both columns. -->
    <div
      class="min-h-0 overflow-hidden p-3 sm:p-4 xl:col-start-1 xl:border-r xl:border-border"
      :class="mobilePane === 'products' ? 'block' : 'hidden xl:block'"
    >
      <ProductPanel
        @add-item="pos.addItem"
        @search-ref="pos.setSearchRef"
      />
    </div>

    <div
      class="min-h-0 min-w-0 overflow-hidden xl:col-start-2"
      :class="mobilePane === 'cart' ? 'block' : 'hidden xl:block'"
    >
      <CartPanel
        :cart="pos.cart"
        :payments="pos.payments"
        :customer="pos.customer.value"
        :order-note="pos.orderNote.value"
        :subtotal="pos.subtotal.value"
        :total-discount="pos.totalDiscount.value"
        :total="pos.total.value"
        :amount-paid="pos.amountPaid.value"
        :amount-owed="pos.amountOwed.value"
        :change="pos.change.value"
        :processing="processing"
        :action-error="actionError"
        :action-message="actionMessage"
        @remove="pos.removeItem"
        @update-qty="pos.updateQty"
        @update-discount="pos.updateDiscount"
        @update-price="pos.updatePrice"
        @update-payment="pos.updatePayment"
        @hold="handleHold"
        @clear="pos.clearCart"
        @complete="handleComplete"
        @complete-and-print="handleCompleteAndPrint"
        @clear-message="actionError = ''; actionMessage = ''"
        @update:customer="pos.customer.value = $event"
        @update:order-note="pos.orderNote.value = $event"
      />
    </div>

    <nav class="z-20 grid grid-cols-2 border-t border-border bg-background p-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] shadow-[0_-6px_20px_-16px_rgba(0,0,0,0.35)] xl:hidden" aria-label="Point of sale sections">
      <button
        type="button"
        class="flex min-h-12 items-center justify-center gap-2 rounded-md px-3 text-sm font-medium transition-colors"
        :class="mobilePane === 'products' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
        :aria-current="mobilePane === 'products' ? 'page' : undefined"
        @click="mobilePane = 'products'"
      >
        <PackageSearch :size="17" aria-hidden="true" />
        <span>Products</span>
      </button>
      <button
        type="button"
        class="flex min-h-12 flex-col items-center justify-center gap-0.5 rounded-md px-2 py-1 text-xs font-medium transition-colors sm:flex-row sm:gap-2 sm:text-sm"
        :class="mobilePane === 'cart' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
        :aria-current="mobilePane === 'cart' ? 'page' : undefined"
        @click="mobilePane = 'cart'"
      >
        <span class="flex items-center gap-1.5">
          <ShoppingCart :size="17" aria-hidden="true" />
          <span>Cart</span>
          <span class="rounded px-1.5 py-0.5 text-[11px] tabular-nums" :class="mobilePane === 'cart' ? 'bg-white/15' : 'bg-muted'">{{ pos.cartCount.value }}</span>
        </span>
        <span class="tabular-nums sm:ml-1 sm:border-l sm:border-current/20 sm:pl-2">₦{{ pos.total.value.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
      </button>
    </nav>

    <Transition name="fade">
      <div v-if="showLeaveWarning" class="fixed inset-0 z-[70] flex items-center justify-center bg-neutral-950/55 p-4 backdrop-blur-sm" role="presentation" @click.self="stayOnPage">
        <section class="w-full max-w-md rounded-xl border border-border bg-card p-5 shadow-2xl sm:p-6" role="alertdialog" aria-modal="true" aria-labelledby="leave-sale-title" aria-describedby="leave-sale-description">
          <div class="mb-4 flex items-start gap-3">
            <span class="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-amber-100 dark:bg-amber-900/30">
              <AlertTriangle :size="20" class="text-amber-700 dark:text-amber-300" aria-hidden="true" />
            </span>
            <div>
              <h2 id="leave-sale-title" class="text-base font-semibold">Unfinished sale</h2>
              <p class="mt-1 text-sm leading-5 text-muted-foreground">Leaving the POS will empty the current cart. Hold the sale to resume it later.</p>
            </div>
          </div>
          <p id="leave-sale-description" class="mb-5 text-sm text-foreground">{{ pos.cartCount.value }} items · ₦{{ pos.total.value.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</p>
          <p v-if="leaveError" class="mb-4 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive" role="alert">{{ leaveError }}</p>
          <div class="flex flex-col-reverse gap-2 sm:flex-row sm:flex-wrap sm:justify-end">
            <Button variant="ghost" :disabled="leaving" @click="discardAndContinue">Discard and continue</Button>
            <Button variant="outline" :disabled="leaving" @click="stayOnPage">Stay on POS</Button>
            <Button :disabled="leaving" @click="holdAndContinue">
              <Pause v-if="!leaving" :size="15" class="mr-2" aria-hidden="true" />
              <span v-else class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-r-transparent"></span>
              {{ leaving ? 'Holding sale…' : 'Hold sale and continue' }}
            </Button>
          </div>
        </section>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { AlertTriangle, PackageSearch, Pause, ShoppingCart } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { usePos } from "../composables/usePos.js";
import ProductPanel from "./pos/ProductPanel.vue";
import CartPanel from "./pos/CartPanel.vue";

const pos = usePos();
const router = useRouter();
const mobilePane = ref("products");
const processing = ref(false);
const actionError = ref("");
const actionMessage = ref("");
const showLeaveWarning = ref(false);
const leaveError = ref("");
const leaving = ref(false);
const pendingRoute = ref(null);
const pendingLogout = ref(null);
let allowNavigationOnce = false;

const removeNavigationGuard = router.beforeEach((to, from) => {
  if (allowNavigationOnce) {
    allowNavigationOnce = false;
    return true;
  }
  if (from.name === "pos" && pos.cart.length > 0) {
    pendingRoute.value = to;
    pendingLogout.value = null;
    leaveError.value = "";
    showLeaveWarning.value = true;
    return false;
  }
  return true;
});

function onLogoutRequest(event) {
  if (pos.cart.length === 0) return;
  event.detail.intercept();
  pendingRoute.value = null;
  pendingLogout.value = event.detail.proceed;
  leaveError.value = "";
  showLeaveWarning.value = true;
}

function stayOnPage() {
  if (leaving.value) return;
  showLeaveWarning.value = false;
  pendingRoute.value = null;
  pendingLogout.value = null;
  leaveError.value = "";
}

function continueLeaving() {
  const logout = pendingLogout.value;
  const target = pendingRoute.value;
  showLeaveWarning.value = false;
  pendingLogout.value = null;
  pendingRoute.value = null;
  if (logout) {
    logout();
  } else if (target) {
    allowNavigationOnce = true;
    router.push(target);
  }
}

function discardAndContinue() {
  if (leaving.value) return;
  pos.clearCart();
  continueLeaving();
}

async function holdAndContinue() {
  if (leaving.value) return;
  leaving.value = true;
  leaveError.value = "";
  try {
    await pos.holdCart();
    continueLeaving();
  } catch (e) {
    leaveError.value = e.message || "The sale could not be held. Please try again.";
  } finally {
    leaving.value = false;
  }
}

onMounted(() => window.addEventListener("pharmacy:logout-request", onLogoutRequest));
onUnmounted(() => {
  window.removeEventListener("pharmacy:logout-request", onLogoutRequest);
  removeNavigationGuard();
});

async function runAction(action, successMessage) {
  if (processing.value) return;
  processing.value = true;
  actionError.value = "";
  actionMessage.value = "";
  try {
    await action();
    actionMessage.value = successMessage;
  } catch (e) {
    actionError.value = e.message || "The action could not be completed. Check your connection and try again.";
  } finally {
    processing.value = false;
  }
}

async function handleHold() {
  await runAction(() => pos.holdCart(), "Sale held. You can resume it from Held Sales.");
}

async function handleComplete() {
  await runAction(async () => {
    await pos.completeSale();
    pos.clearCart();
  }, "Sale completed.");
}

async function handleCompleteAndPrint() {
  await runAction(async () => {
    await pos.completeSale();
    pos.printReceipt();
    pos.clearCart();
  }, "Sale completed. Receipt sent to print.");
}
</script>
