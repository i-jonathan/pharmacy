<template>
  <div class="flex flex-col h-full border-l border-border bg-card p-3">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-border">
      <h2 class="text-lg font-semibold">Current Sale</h2>
      <div class="flex items-center gap-1.5">
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
    <div class="px-4 py-2 border-border">
      <div class="text-xs text-muted-foreground mb-1">Customer</div>
      <div class="flex items-center gap-2">
        <User :size="14" class="text-muted-foreground shrink-0" />
        <input
          :value="customer"
          @input="$emit('update:customer', $event.target.value)"
          class="flex-1 text-sm bg-transparent border-none outline-none"
          placeholder="Walk-in Customer"
        />
        <Button variant="outline" size="sm" class="text-xs h-7">+ New</Button>
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
              <div class="text-sm font-medium truncate max-w-[140px]">{{ item.name }}</div>
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              &#8358;{{ item.price.toLocaleString() }}
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

    <!-- Order Note + Totals -->
    <div class="px-4 py-2 border-t border-border">
      <div class="flex gap-3">
        <div class="w-1/2">
          <input
            :value="orderNote"
            @input="$emit('update:orderNote', $event.target.value)"
            class="w-full text-xs text-muted-foreground bg-transparent border border-border rounded-sm px-2 py-2 outline-none focus:ring-1 focus:ring-ring"
            placeholder="Add order note..."
          />
        </div>
        <div class="w-1/2 space-y-5 bg-muted/50 rounded-sm px-6 py-6">
          <div class="flex justify-between text-xs">
            <span class="text-muted-foreground font-bo">Subtotal</span>
            <span class="font-bold">&#8358;{{ subtotal.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
          </div>
          <div class="flex justify-between text-xs">
            <span class="text-muted-foreground">Discount</span>
            <span>&#8358;{{ totalDiscount.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
          </div>
          <div class="flex justify-between font-bold text-lg pt-2 border-t border-border">
            <span>Total</span>
            <span>&#8358;{{ total.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Payment Methods -->
    <div class="px-4 py-4 border-border">
      <div class="text-sm font-semibold mb-3">Payment Methods</div>
      <div class="border border-border rounded-sm divide-y divide-border">
        <div v-for="method in paymentMethods" :key="method.key" class="flex items-center gap-3 px-3 py-2">
          <component :is="method.icon" :size="16" :class="method.color" class="shrink-0" />
          <span class="text-sm text-muted-foreground w-16">{{ method.label }}</span>
          <div class="flex-1 flex items-center border border-border rounded-sm overflow-hidden">
            <span class="pl-2 pr-1 text-sm text-muted-foreground">&#8358;</span>
            <input
              :value="payments[method.key] || ''"
              @input="$emit('update-payment', method.key, Number($event.target.value) || 0)"
              type="text"
              inputmode="decimal"
              class="flex-1 py-1.5 pr-2 text-sm bg-transparent outline-none"
              placeholder="0.00"
            />
          </div>
          <button
            class="shrink-0 w-6 h-6 flex items-center justify-center text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-sm transition-colors"
            @click="$emit('update-payment', method.key, 0)"
          >
            <X :size="13" />
          </button>
        </div>
      </div>
    </div>

    <!-- Amount Paid + Change -->
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
  </div>
</template>

<script setup>
import { Pause, Trash2, User, Minus, Plus, X, Pencil, CircleCheck, Printer, Banknote, CreditCard, PiggyBank } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

defineProps({
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

defineEmits([
  "remove",
  "update-qty",
  "update-discount",
  "update-payment",
  "hold",
  "clear",
  "complete",
  "complete-and-print",
  "update:customer",
  "update:orderNote",
]);
</script>
