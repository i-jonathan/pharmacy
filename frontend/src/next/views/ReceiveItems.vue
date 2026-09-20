<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Receive Items</h1>
      <p class="text-sm text-muted-foreground mt-1">Record incoming inventory from suppliers</p>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">Pending</div>
        <div class="text-2xl font-bold text-foreground">0</div>
      </div>
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">Received Today</div>
        <div class="text-2xl font-bold text-foreground">0</div>
      </div>
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">This Month</div>
        <div class="text-2xl font-bold text-foreground">0</div>
      </div>
    </div>

    <!-- Empty State -->
    <div class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
      <Truck :size="48" class="text-muted-foreground/40 mb-4" />
      <h3 class="text-lg font-semibold text-foreground mb-1">No pending receipts</h3>
      <p class="text-sm text-muted-foreground max-w-sm">
        Create a new receiving record when inventory arrives from your suppliers.
      </p>
      <Button class="mt-4">
        <Plus :size="16" class="mr-2" />
        New Receipt
      </Button>
    </div>
  </div>
</template>

<script setup>
import { Truck, Plus } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
</script>