<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Held Receive Items</h1>
      <p class="text-sm text-muted-foreground mt-1">Draft and incomplete inventory receipts</p>
    </div>

    <!-- Empty State -->
    <div class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
      <PauseCircle :size="48" class="text-muted-foreground/40 mb-4" />
      <h3 class="text-lg font-semibold text-foreground mb-1">No held receipts</h3>
      <p class="text-sm text-muted-foreground max-w-sm">
        Incomplete inventory receipts will appear here so you can resume or discard them later.
      </p>
      <Button variant="outline" class="mt-4" @click="$router.push('/receive-items')">
        <Truck :size="16" class="mr-2" />
        Go to Receive Items
      </Button>
    </div>
  </div>
</template>

<script setup>
import { PauseCircle, Truck } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
</script>