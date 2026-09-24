<template>
  <div class="flex flex-col h-full border-l border-border bg-card p-3">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-border">
      <h2 class="text-lg font-semibold">Current Sale</h2>
      <div class="flex items-center gap-1.5">
        <router-link to="/held-sales">
          <Button variant="outline" size="sm">
            <History :size="13" class="mr-1" />
            Held
          </Button>
        </router-link>
        <Button variant="outline" size="sm" @click="$emit('hold')" :disabled="cart.length === 0">
          <Pause :size="13" class="mr-1" />
          Hold (F6)
        </Button>
        <Button variant="ghost" size="sm" class="text-red-600 dark:text-red-400" @click="$emit('clear')" :disabled="cart.length === 0">
          <Trash2 :size="13" class="mr-1" />
          Clear
        </Button>
      </div>
    </div>

    <!-- Customer -->
    <div class="px-4 py-2 border-border space-y-1.5">
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
      <div class="flex items-center gap-2">
        <Pencil :size="12" class="text-muted-foreground shrink-0" />
        <input
          :value="orderNote"
          @input="$emit('update:orderNote', $event.target.value)"
          class="flex-1 text-xs text-muted-foreground bg-transparent border border-border rounded-sm px-2 py-1.5 outline-none focus:ring-1 focus:ring-ring"
          placeholder="Add order note..."
        />
      </div>
    </div>

    <!-- Cart Items -->
    <div class="flex-1 overflow-auto">
      <div v-if="cart.length === 0" class="flex items-center justify-center h-full text-sm text-muted-foreground">
        Cart is empty. Search and add products.
      </div>

      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead>Item</TableHead>
            <TableHead class="w-16">Price</TableHead>
            <TableHead class="w-24 text-center">Qty</TableHead>
            <TableHead class="w-16">Disc.</TableHead>
            <TableHead class="w-20 text-right">Total</TableHead>
            <TableHead class="w-8"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(item, index) in cart" :key="index">
            <TableCell>
              <div class="text-sm font-medium">{{ item.name }}</div>
              <div v-if="item.manufacturer" class="text-xs text-muted-foreground">{{ item.manufacturer }}</div>
            </TableCell>
            <TableCell
              class="text-sm text-muted-foreground price-trigger"
              :class="{ 'cursor-pointer hover:text-foreground': hasPriceOptions(item) }"
              @click="hasPriceOptions(item) && togglePricePopover($event, index)"
            >
              <div>
                <div class="flex items-center gap-1">
                  <span>&#8358;{{ item.price.toLocaleString() }}</span>
                  <ChevronDown v-if="hasPriceOptions(item)" :size="10" class="text-muted-foreground" />
                </div>
                <div class="text-[10px] text-muted-foreground">{{ currentPriceName(item) }}</div>
              </div>
            </TableCell>
            <TableCell>
              <div class="inline-flex items-center border border-border rounded-sm">
                <button
                  class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent rounded-l-sm transition-colors"
                  @click="$emit('update-qty', index, item.qty - 1)"
                >
                  <Minus :size="12" />
                </button>
                <input
                  :value="item.qty"
                  @input="$emit('update-qty', index, Number($event.target.value) || 0)"
                  class="h-7 w-10 text-center text-sm bg-transparent border-x border-border outline-none"
                />
                <button
                  class="h-7 w-7 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent rounded-r-sm transition-colors"
                  @click="$emit('update-qty', index, item.qty + 1)"
                >
                  <Plus :size="12" />
                </button>
              </div>
            </TableCell>
            <TableCell>
              <input
                :value="item.discount || 0"
                @input="$emit('update-discount', index, Number($event.target.value) || 0)"
                class="w-14 text-center text-xs border border-border rounded px-1 py-0.5 bg-transparent"
                placeholder="0"
              />
            </TableCell>
            <TableCell class="text-right text-sm font-semibold">
              &#8358;{{ ((item.price * item.qty) - (item.discount || 0)).toLocaleString() }}
            </TableCell>
            <TableCell>
              <Button variant="ghost" size="icon" class="h-6 w-6 text-muted-foreground hover:text-destructive" @click="$emit('remove', index)">
                <X :size="12" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- Totals + Payment Methods (shared row) -->
    <div class="px-4 py-3 border-t border-border">
      <div class="flex gap-3">
        <!-- Totals -->
        <div class="w-1/2 space-y-3 bg-muted/50 rounded-sm px-4 py-3">
          <div class="flex justify-between text-xs">
            <span class="text-muted-foreground font-bo">Subtotal</span>
            <span class="font-bold">&#8358;{{ subtotal.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
          </div>
          <div class="flex justify-between text-xs">
            <span class="text-muted-foreground">Discount</span>
            <span>&#8358;{{ totalDiscount.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
          </div>
          <div class="flex justify-between font-bold text-sm pt-2 border-t border-border">
            <span>Total</span>
            <span>&#8358;{{ total.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
          </div>
        </div>

        <!-- Payment Methods -->
        <div class="w-1/2">
          <div class="text-xs font-semibold mb-2">Payment Methods</div>
          <div class="border border-border rounded-sm divide-y divide-border">
            <div v-for="method in paymentMethods" :key="method.key" class="flex items-center gap-2 px-2.5 py-1.5">
              <component :is="method.icon" :size="14" :class="method.color" class="shrink-0" />
              <span class="text-xs text-muted-foreground w-14">{{ method.label }}</span>
              <div class="flex items-center border border-border rounded-sm overflow-hidden flex-1">
                <span class="pl-1.5 pr-0.5 text-xs text-muted-foreground">&#8358;</span>
                <input
                  :value="payments[method.key] || ''"
                  @input="$emit('update-payment', method.key, Number($event.target.value) || 0)"
                  type="text"
                  inputmode="decimal"
                  class="w-full py-1 pr-1.5 text-xs bg-transparent outline-none"
                  placeholder="0"
                />
              </div>
              <button
                class="shrink-0 w-5 h-5 flex items-center justify-center text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-sm transition-colors"
                @click="$emit('update-payment', method.key, 0)"
              >
                <X :size="11" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Amount Owed / Paid / Change -->
    <div class="px-4 py-2 border-t border-border space-y-1">
      <div class="flex justify-between">
        <span class="text-muted-foreground text-sm">Amount Owed</span>
        <span :class="amountOwed > 0 ? 'text-red-600 dark:text-red-400 font-semibold' : 'text-emerald-600'">
          &#8358;{{ amountOwed.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}
        </span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground text-sm">Amount Paid</span>
        <span>&#8358;{{ amountPaid.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground text-sm">Change</span>
        <span class="font-bold text-emerald-600">&#8358;{{ change.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
      </div>
    </div>

    <!-- Complete Sale -->
    <div class="px-4 py-3 border-border flex items-center gap-2">
      <Button
        class="flex-1 disabled:opacity-50"
        size="lg"
        :disabled="cart.length === 0 || amountOwed > 0"
        @click="$emit('complete')"
      >
        <CircleCheck :size="16" class="mr-2" />
        Complete Sale (F5)
      </Button>
      <Button
        variant="outline"
        size="icon"
        class="h-11 w-11 shrink-0 disabled:opacity-50"
        :disabled="cart.length === 0 || amountOwed > 0"
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
import { Pause, Trash2, User, Minus, Plus, X, Pencil, CircleCheck, Printer, ChevronDown, Banknote, CreditCard, PiggyBank, History } from "lucide-vue-next";
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
  orderNote: { type: String, default: "" },
  subtotal: { type: Number, default: 0 },
  totalDiscount: { type: Number, default: 0 },
  total: { type: Number, default: 0 },
  amountPaid: { type: Number, default: 0 },
  amountOwed: { type: Number, default: 0 },
  change: { type: Number, default: 0 },
});

const paymentMethods = [
  { key: "Cash", label: "Cash", icon: Banknote, color: "text-emerald-500" },
  { key: "Card", label: "Card", icon: CreditCard, color: "text-blue-500" },
  { key: "Transfer", label: "Transfer", icon: PiggyBank, color: "text-amber-500" },
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
  "update:orderNote",
]);
</script>
