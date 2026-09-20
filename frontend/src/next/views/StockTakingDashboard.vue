<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Stock Taking</h1>
      <p class="text-sm text-muted-foreground mt-1">Conduct and manage inventory counts</p>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="flex items-center gap-2 mb-1">
          <ClipboardCheck :size="16" class="text-emerald-500" />
          <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Completed</span>
        </div>
        <div class="text-2xl font-bold text-foreground">0</div>
      </div>
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="flex items-center gap-2 mb-1">
          <Clock :size="16" class="text-amber-500" />
          <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">In Progress</span>
        </div>
        <div class="text-2xl font-bold text-foreground">0</div>
      </div>
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="flex items-center gap-2 mb-1">
          <AlertTriangle :size="16" class="text-red-500" />
          <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Discrepancies</span>
        </div>
        <div class="text-2xl font-bold text-foreground">0</div>
      </div>
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="flex items-center gap-2 mb-1">
          <Package :size="16" class="text-primary" />
          <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Items Counted</span>
        </div>
        <div class="text-2xl font-bold text-foreground">0</div>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex gap-3 mb-6">
      <Button>
        <Plus :size="16" class="mr-2" />
        New Stock Count
      </Button>
      <Button variant="outline">
        <History :size="16" class="mr-2" />
        View History
      </Button>
    </div>

    <!-- Empty State -->
    <div class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
      <ClipboardCheck :size="48" class="text-muted-foreground/40 mb-4" />
      <h3 class="text-lg font-semibold text-foreground mb-1">No stock counts yet</h3>
      <p class="text-sm text-muted-foreground max-w-sm">
        Start a new stock count to track inventory levels and identify discrepancies.
      </p>
    </div>
  </div>
</template>

<script setup>
import { ClipboardCheck, Clock, AlertTriangle, Package, Plus, History } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
</script>