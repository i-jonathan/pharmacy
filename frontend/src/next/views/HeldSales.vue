<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Held Sales</h1>
      <p class="text-sm text-muted-foreground mt-1">Resume or void temporarily paused transactions</p>
    </div>

    <!-- Empty State -->
    <div class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
      <PauseCircle :size="48" class="text-muted-foreground/40 mb-4" />
      <h3 class="text-lg font-semibold text-foreground mb-1">No held sales</h3>
      <p class="text-sm text-muted-foreground max-w-sm">
        Sales that are paused from the POS will appear here. You can resume or void them later.
      </p>
      <Button variant="outline" class="mt-4" @click="$router.push('/pos')">
        <ShoppingCart :size="16" class="mr-2" />
        Go to POS
      </Button>
    </div>
  </div>
</template>

<script setup>
import { PauseCircle, ShoppingCart } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
</script>