<template>
  <div class="flex h-full flex-col border-l border-border bg-card p-3 sm:p-4">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-2 border-border">
      <div>
        <h2 class="text-base font-semibold tracking-tight">Current sale</h2>
        <p class="mt-0.5 text-xs text-muted-foreground">{{ cart.length }} {{ cart.length === 1 ? 'item' : 'items' }}</p>
      </div>
      <div class="flex items-center gap-1.5">
        <router-link to="/held-sales">
          <Button variant="outline" size="sm">
            <History :size="13" class="mr-1" />
            Held
          </Button>
        </router-link>
        <Button variant="outline" size="sm" @click="$emit('hold')" :disabled="cart.length === 0 || processing">
          <Pause :size="13" class="mr-1" />
          Hold (F6)
        </Button>
        <Button variant="ghost" size="sm" class="text-muted-foreground hover:text-destructive" @click="$emit('clear')" :disabled="cart.length === 0 || processing">
          <Trash2 :size="13" class="mr-1" />
          Clear
        </Button>
      </div>
    </div>

    <!-- Customer -->
    <div class="px-4 py-1.5 border-border">
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-2 flex-1">
          <User :size="14" class="text-muted-foreground shrink-0" />
          <input
            :value="customer"
            @input="$emit('update:customer', $event.target.value)"
            class="flex-1 text-sm bg-transparent border-none outline-none"
            placeholder="Walk-in Customer"
          />
        </div>
        <Button variant="outline" size="sm" class="text-xs h-7">+ New</Button>
      </div>
    </div>

    <!-- Cart Items -->
    <div class="min-h-0 flex-1 overflow-auto">
      <div v-if="cart.length === 0" class="flex h-full min-h-36 flex-col items-center justify-center px-5 text-center text-sm text-muted-foreground">
        <ShoppingCart :size="22" class="mb-2 text-muted-foreground/60" aria-hidden="true" />
        <p class="font-medium text-foreground">Your sale is ready</p>
        <p class="mt-1">Search or scan a product to add it here.</p>
      </div>

      <Table v-else class="text-xs">
        <TableHeader>
          <TableRow>
            <TableHead class="h-8 px-2 text-[10px]">Item</TableHead>
            <TableHead class="h-8 w-[4.5rem] px-1.5 text-[10px]">Price</TableHead>
            <TableHead class="h-8 w-[5rem] px-1 text-center text-[10px]">Qty</TableHead>
            <TableHead class="h-8 w-[3.5rem] px-1 text-[10px]">Disc.</TableHead>
            <TableHead class="h-8 w-[4.5rem] px-1.5 text-right text-[10px]">Total</TableHead>
            <TableHead class="h-8 w-7 px-0"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(item, index) in cart" :key="index">
            <TableCell class="max-w-0 px-2 py-1.5">
              <div class="truncate text-xs font-medium leading-4">{{ item.name }}</div>
              <div v-if="item.manufacturer" class="truncate text-[10px] leading-3 text-muted-foreground">{{ item.manufacturer }}</div>
            </TableCell>
            <TableCell
              class="price-trigger px-1.5 py-1.5 text-[11px] text-muted-foreground"
              :class="{ 'cursor-pointer hover:text-foreground': hasPriceOptions(item) }"
              @click="hasPriceOptions(item) && togglePricePopover($event, index)"
            >
              <div class="flex items-center gap-0.5">
                <span class="whitespace-nowrap">&#8358;{{ item.price.toLocaleString() }}</span>
                <ChevronDown v-if="hasPriceOptions(item)" :size="9" class="shrink-0" />
              </div>
              <div class="truncate text-[9px] leading-3">{{ currentPriceName(item) }}</div>
            </TableCell>
            <TableCell class="px-1 py-1.5">
              <div class="inline-flex items-center rounded-sm border border-border">
                <button
                  type="button"
                  aria-label="Decrease quantity"
                  class="flex h-6 w-5 items-center justify-center rounded-l-sm text-muted-foreground hover:bg-accent hover:text-foreground"
                  @click="$emit('update-qty', index, item.qty - 1)"
                ><Minus :size="11" /></button>
                <input
                  type="number"
                  min="1"
                  step="1"
                  inputmode="numeric"
                  :aria-label="`Quantity of ${item.name}`"
                  :value="item.qty"
                  @input="$emit('update-qty', index, Number($event.target.value) || 0)"
                  class="no-spinners h-6 w-7 border-x border-border bg-transparent text-center text-xs outline-none"
                />
                <button
                  type="button"
                  aria-label="Increase quantity"
                  class="flex h-6 w-5 items-center justify-center rounded-r-sm text-muted-foreground hover:bg-accent hover:text-foreground"
                  @click="$emit('update-qty', index, item.qty + 1)"
                ><Plus :size="11" /></button>
              </div>
            </TableCell>
            <TableCell class="px-1 py-1.5">
              <input
                :value="item.discount || 0"
                :aria-label="`Discount for ${item.name}`"
                @input="$emit('update-discount', index, Number($event.target.value) || 0)"
                class="w-10 rounded border border-border bg-transparent px-0.5 py-1 text-center text-[10px]"
                placeholder="0"
              />
            </TableCell>
            <TableCell class="whitespace-nowrap px-1.5 py-1.5 text-right text-xs font-semibold tabular-nums">
              &#8358;{{ ((item.price * item.qty) - (item.discount || 0)).toLocaleString() }}
            </TableCell>
            <TableCell class="px-0 py-1">
              <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-destructive" :aria-label="`Remove ${item.name}`" @click="$emit('remove', index)">
                <X :size="13" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- Payment and total -->
    <div class="space-y-3 border-t border-border px-4 py-3">
      <div class="flex items-end justify-between gap-3 rounded-md bg-muted/50 px-4 py-3">
        <div>
          <p class="text-xs font-medium text-muted-foreground">Total due</p>
          <p class="mt-1 text-xs text-muted-foreground">Subtotal ₦{{ subtotal.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}<span v-if="totalDiscount > 0"> · Discount ₦{{ totalDiscount.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span></p>
        </div>
        <strong class="whitespace-nowrap text-2xl font-semibold tabular-nums tracking-tight">₦{{ total.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</strong>
      </div>

      <div>
        <div class="mb-2 text-xs font-semibold text-foreground">Payment received</div>
        <div class="grid grid-cols-3 gap-2">
            <label v-for="method in paymentMethods" :key="method.key" class="min-w-0 rounded-md border border-border bg-background px-2 py-2">
              <component :is="method.icon" :size="14" :class="method.color" class="shrink-0" />
              <span class="mb-1 block text-xs font-medium text-muted-foreground">{{ method.label }}</span>
              <div class="flex items-center overflow-hidden rounded border border-input">
                <span class="pl-1.5 pr-0.5 text-xs text-muted-foreground">&#8358;</span>
                <input
                  :value="payments[method.key] || ''"
                  @input="$emit('update-payment', method.key, Number($event.target.value) || 0)"
                  type="text"
                  inputmode="decimal"
                  class="min-w-0 w-full bg-transparent py-1.5 pr-1 text-xs tabular-nums outline-none"
                  :aria-label="`${method.label} payment amount`"
                  placeholder="0"
                  :disabled="processing"
                />
              </div>
            </label>
          </div>
      </div>
    </div>

    <!-- Settlement summary -->
    <div class="grid grid-cols-3 gap-2 border-t border-border px-4 py-2.5 text-xs">
      <div>
        <span class="block text-muted-foreground">Still due</span>
        <span class="mt-1 block font-semibold tabular-nums" :class="amountOwed > 0 ? 'text-amber-700 dark:text-amber-300' : 'text-primary'">
          ₦{{ amountOwed.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}
        </span>
      </div>
      <div>
        <span class="block text-muted-foreground">Received</span>
        <span class="mt-1 block font-medium tabular-nums">₦{{ amountPaid.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
      </div>
      <div>
        <span class="block text-muted-foreground">Change</span>
        <span class="mt-1 block font-medium tabular-nums">₦{{ change.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
      </div>
    </div>

    <div v-if="actionError" class="mx-4 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive" role="alert">
      {{ actionError }}
      <button class="ml-2 font-medium underline underline-offset-2" @click="$emit('clear-message')">Dismiss</button>
    </div>
    <div v-else-if="actionMessage" class="mx-4 rounded-md border border-primary/20 bg-primary/5 px-3 py-2 text-sm text-primary" role="status" aria-live="polite">
      {{ actionMessage }}
      <button class="ml-2 font-medium underline underline-offset-2" @click="$emit('clear-message')">Dismiss</button>
    </div>

    <!-- Complete Sale -->
    <div class="flex items-center gap-2 px-4 py-3">
      <Button
        class="flex-1 disabled:opacity-50"
        size="lg"
        :disabled="cart.length === 0 || amountOwed > 0 || processing"
        @click="$emit('complete')"
      >
        <CircleCheck v-if="!processing" :size="16" class="mr-2" />
        <span v-else class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-r-transparent"></span>
        {{ processing ? 'Processing…' : 'Complete sale (F5)' }}
      </Button>
      <Button
        variant="outline"
        size="icon"
        class="h-11 w-11 shrink-0 disabled:opacity-50"
        :disabled="cart.length === 0 || amountOwed > 0 || processing"
        aria-label="Complete sale and print receipt"
        title="Complete sale and print receipt"
        @click="$emit('complete-and-print')"
      >
        <Printer :size="18" />
      </Button>
    </div>

    <Teleport to="body">
      <div
        v-if="pricePopover.index !== null"
        class="price-dropdown fixed z-60 w-44 rounded-sm border border-border bg-popover shadow-lg p-1"
        :style="{ top: pricePopover.y + 'px', left: pricePopover.x + 'px' }"
      >
        <div class="text-xs text-muted-foreground px-2 py-1.5 border-b border-border">Change price</div>
        <button
          v-for="opt in pricePopoverOptions"
          :key="opt.id"
          class="flex items-center justify-between w-full px-2 py-1.5 text-sm rounded-sm hover:bg-primary/15 transition-colors"
          :class="{ 'bg-primary/20 font-medium': opt.id === pricePopover.currentId }"
          @click="selectPriceOption(opt)"
        >
          <span>{{ opt.name || 'Base' }}</span>
          <span class="font-medium">&#8358;{{ opt.price.toLocaleString() }}</span>
        </button>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { reactive, computed, onMounted, onUnmounted } from "vue";
import { Pause, Trash2, User, Minus, Plus, X, CircleCheck, Printer, ChevronDown, Banknote, CreditCard, PiggyBank, History, ShoppingCart } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const props = defineProps({
  cart: { type: Array, required: true },
  payments: { type: Object, default: () => ({ Cash: 0, Card: 0, Transfer: 0 }) },
  customer: { type: String, default: "Walk-in Customer" },
  subtotal: { type: Number, default: 0 },
  totalDiscount: { type: Number, default: 0 },
  total: { type: Number, default: 0 },
  amountPaid: { type: Number, default: 0 },
  amountOwed: { type: Number, default: 0 },
  change: { type: Number, default: 0 },
  processing: { type: Boolean, default: false },
  actionError: { type: String, default: "" },
  actionMessage: { type: String, default: "" },
});

const paymentMethods = [
  { key: "Cash", label: "Cash", icon: Banknote, color: "text-primary" },
  { key: "Card", label: "Card", icon: CreditCard, color: "text-primary" },
  { key: "Transfer", label: "Transfer", icon: PiggyBank, color: "text-primary" },
];

const pricePopover = reactive({ index: null, x: 0, y: 0, currentId: 0 });

const pricePopoverOptions = computed(() => {
  if (pricePopover.index === null) return [];
  const item = props.cart[pricePopover.index];
  return item?.priceOptions || [];
});

function currentPriceName(item) {
  const opt = item.priceOptions?.find((o) => o.id === item.priceId);
  return opt?.name || "Base";
}

function hasPriceOptions(item) {
  return item.priceOptions && item.priceOptions.length > 1;
}

function togglePricePopover(event, index) {
  if (pricePopover.index === index) {
    pricePopover.index = null;
    return;
  }
  const rect = event.currentTarget.getBoundingClientRect();
  pricePopover.index = index;
  pricePopover.currentId = props.cart[index].priceId;
  pricePopover.x = rect.right - 190;
  pricePopover.y = Math.min(rect.bottom + 4, window.innerHeight - 250);
}

function selectPriceOption(opt) {
  emit("update-price", pricePopover.index, opt.id, opt.price);
  pricePopover.index = null;
}

function onDocumentClick(e) {
  if (pricePopover.index === null) return;
  // Don't close if clicking the trigger cell (it toggles the popover itself)
  if (e.target.closest(".price-trigger")) return;
  const el = document.querySelector(".price-dropdown");
  if (el && !el.contains(e.target)) {
    pricePopover.index = null;
  }
}

onMounted(() => document.addEventListener("click", onDocumentClick));
onUnmounted(() => document.removeEventListener("click", onDocumentClick));

const emit = defineEmits([
  "clear-message",
  "remove",
  "update-qty",
  "update-discount",
  "update-price",
  "update-payment",
  "hold",
  "clear",
  "complete",
  "complete-and-print",
  "update:customer",
]);
</script>
