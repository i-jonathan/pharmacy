<template>
  <div class="grid h-[calc(100dvh-3.5rem)] min-h-0 grid-cols-1 grid-rows-[minmax(0,1fr)_auto] xl:grid-cols-[minmax(0,1.05fr)_minmax(28rem,0.95fr)] xl:grid-rows-1">
    <!-- Mobile keeps one workspace panel visible; desktop shows both columns. -->
    <div
      class="min-h-0 overflow-hidden p-3 sm:p-4 xl:col-start-1 xl:border-r xl:border-border"
      :class="mobilePane === 'products' ? 'block' : 'hidden xl:block'"
    >
      <ProductPanel
        @add-item="handleAddItem"
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
        :subtotal="pos.subtotal.value"
        :total-discount="pos.totalDiscount.value"
        :total="pos.total.value"
        :amount-paid="pos.amountPaid.value"
        :amount-owed="pos.amountOwed.value"
        :change="pos.change.value"
        :processing="processing"
        :action-error="actionError"
        :action-message="actionMessage"
        @remove="handleRemoveItem"
        @update-qty="pos.updateQty"
        @update-discount="pos.updateDiscount"
        @update-price="pos.updatePrice"
        @update-payment="pos.updatePayment"
        @hold="handleHold"
        @clear="pos.clearCart"
        @complete="handleComplete"
        @complete-and-print="handleCompleteAndPrint"
        @clear-message="clearActionMessage"
        @update:customer="pos.customer.value = $event"
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
      <div
        v-if="addedItemName"
        class="fixed inset-x-3 z-30 flex items-center gap-3 rounded-lg border border-primary/20 bg-card px-4 py-3 shadow-lg xl:hidden"
        :style="{ bottom: addedItemToastBottom }"
        role="status"
        aria-live="polite"
      >
        <span class="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-primary/10 text-primary">
          <Check :size="17" aria-hidden="true" />
        </span>
        <span class="min-w-0 flex-1 truncate text-sm font-medium">{{ addedItemName }} added to cart</span>
        <button type="button" class="shrink-0 text-sm font-semibold text-primary" @click="openCart">View cart</button>
      </div>
    </Transition>

  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { Check, PackageSearch, ShoppingCart } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { usePos } from "../composables/usePos.js";
import { useConfirm } from "../composables/useConfirm.js";
import ProductPanel from "./pos/ProductPanel.vue";
import CartPanel from "./pos/CartPanel.vue";

const pos = usePos();
const { confirm } = useConfirm();
const router = useRouter();
const mobilePane = ref("products");
const processing = ref(false);
const actionError = ref("");
const actionMessage = ref("");
const addedItemName = ref("");
const keyboardOffset = ref(0);
const addedItemToastBottom = computed(() => `calc(${keyboardOffset.value}px + 4.5rem + env(safe-area-inset-bottom))`);
let addedItemTimer = null;
let actionMessageTimer = null;

// Resolves true when the user may leave: after holding the sale, or after discarding the cart.
async function mayLeave() {
  const answer = await confirm({
    title: "Unfinished sale",
    message: "Leaving the POS will empty the current cart. Hold the sale to resume it later.",
    detail: `${pos.cartCount.value} items · ₦${pos.total.value.toLocaleString(undefined, { minimumFractionDigits: 2 })}`,
    confirmLabel: "Hold sale and continue",
    busyLabel: "Holding sale…",
    cancelLabel: "Stay on POS",
    alternate: { label: "Discard and continue", danger: true, run: async () => pos.clearCart() },
    run: async () => {
      try {
        await pos.holdCart();
      } catch (e) {
        throw new Error(e.message || "The sale could not be held. Please try again.");
      }
    },
  });
  return answer !== false;
}

const removeNavigationGuard = router.beforeEach((to, from) => {
  if (from.name === "pos" && pos.cart.length > 0) return mayLeave();
  return true;
});

async function onLogoutRequest(event) {
  if (pos.cart.length === 0) return;
  event.detail.intercept();
  if (await mayLeave()) event.detail.proceed();
}

function updateKeyboardOffset() {
  const viewport = window.visualViewport;
  keyboardOffset.value = viewport
    ? Math.max(0, window.innerHeight - viewport.height - viewport.offsetTop)
    : 0;
}

onMounted(() => {
  window.addEventListener("pharmacy:logout-request", onLogoutRequest);
  window.addEventListener("resize", updateKeyboardOffset);
  window.visualViewport?.addEventListener("resize", updateKeyboardOffset);
  window.visualViewport?.addEventListener("scroll", updateKeyboardOffset);
  updateKeyboardOffset();
});
onUnmounted(() => {
  window.removeEventListener("pharmacy:logout-request", onLogoutRequest);
  window.removeEventListener("resize", updateKeyboardOffset);
  window.visualViewport?.removeEventListener("resize", updateKeyboardOffset);
  window.visualViewport?.removeEventListener("scroll", updateKeyboardOffset);
  removeNavigationGuard();
  clearTimeout(addedItemTimer);
  clearTimeout(actionMessageTimer);
});

function handleAddItem(product, priceId, price) {
  pos.addItem(product, priceId, price);
  addedItemName.value = product.name;
  clearTimeout(addedItemTimer);
  addedItemTimer = setTimeout(() => {
    addedItemName.value = "";
  }, 3000);
}

function openCart() {
  mobilePane.value = "cart";
  addedItemName.value = "";
}

function clearActionMessage() {
  clearTimeout(actionMessageTimer);
  actionError.value = "";
  actionMessage.value = "";
}

async function runAction(action, successMessage) {
  if (processing.value) return;
  processing.value = true;
  clearTimeout(actionMessageTimer);
  actionError.value = "";
  actionMessage.value = "";
  try {
    await action();
    actionMessage.value = successMessage;
    actionMessageTimer = setTimeout(() => {
      actionMessage.value = "";
    }, 5000);
  } catch (e) {
    actionError.value = e.message || "The action could not be completed. Check your connection and try again.";
  } finally {
    processing.value = false;
  }
}

async function returnToSearch() {
  mobilePane.value = "products";
  await nextTick();
  pos.focusSearch();
}

function handleRemoveItem(index) {
  pos.removeItem(index);
  if (window.matchMedia("(min-width: 1280px)").matches) {
    pos.focusSearch();
  }
}

async function handleHold() {
  await runAction(() => pos.holdCart(), "Sale held. You can resume it from Held Sales.");
}

async function handleComplete() {
  await runAction(async () => {
    await pos.completeSale();
    pos.clearCart();
    await returnToSearch();
  }, "Sale completed.");
}

async function handleCompleteAndPrint() {
  await runAction(async () => {
    await pos.completeSale();
    pos.printReceipt();
    pos.clearCart();
    await returnToSearch();
  }, "Sale completed. Receipt sent to print.");
}
</script>
