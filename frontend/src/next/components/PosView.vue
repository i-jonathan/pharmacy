<template>
  <div class="flex h-[calc(100vh-3.5rem)]">
    <!-- Left: Product Panel -->
    <div class="w-1/2 overflow-hidden p-4 border-r border-border">
      <ProductPanel
        @add-item="pos.addItem"
        @search-ref="pos.setSearchRef"
      />
    </div>

    <!-- Right: Cart Panel -->
    <div class="w-1/2 overflow-hidden">
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
        @remove="pos.removeItem"
        @update-qty="pos.updateQty"
        @update-discount="pos.updateDiscount"
        @update-payment="pos.updatePayment"
        @hold="handleHold"
        @clear="pos.clearCart"
        @complete="handleComplete"
        @complete-and-print="handleCompleteAndPrint"
        @update:customer="pos.customer.value = $event"
        @update:order-note="pos.orderNote.value = $event"
      />
    </div>
  </div>
</template>

<script setup>
import { usePos } from "../composables/usePos.js";
import ProductPanel from "./pos/ProductPanel.vue";
import CartPanel from "./pos/CartPanel.vue";

const pos = usePos();

async function handleHold() {
  try {
    await pos.holdCart();
  } catch (e) {
    console.error("Failed to hold sale:", e);
  }
}

async function handleComplete() {
  try {
    await pos.completeSale();
    pos.clearCart();
  } catch (e) {
    console.error("Failed to complete sale:", e);
  }
}

async function handleCompleteAndPrint() {
  try {
    await pos.completeSale();
    pos.printReceipt();
    pos.clearCart();
  } catch (e) {
    console.error("Failed to complete sale:", e);
  }
}
</script>
