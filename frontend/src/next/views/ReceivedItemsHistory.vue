<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Received Items History</h1>
      <p class="text-sm text-muted-foreground mt-1">View all completed inventory receipts</p>
    </div>

    <!-- Filters -->
    <div class="flex flex-col sm:flex-row gap-3 mb-6">
      <div class="relative flex-1">
        <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search by receipt #, supplier, or item..."
          class="w-full pl-9 pr-4 py-2 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
        />
      </div>
      <div class="flex gap-2">
        <select class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring">
          <option>Last 7 Days</option>
          <option>Last 30 Days</option>
          <option>Last 90 Days</option>
          <option>All Time</option>
        </select>
        <Button variant="outline" size="icon">
          <CalendarRange :size="16" />
        </Button>
      </div>
    </div>

    <!-- Empty State -->
    <div class="rounded-lg border border-border bg-card p-12 flex flex-col items-center justify-center text-center">
      <ClipboardList :size="48" class="text-muted-foreground/40 mb-4" />
      <h3 class="text-lg font-semibold text-foreground mb-1">No receipt history</h3>
      <p class="text-sm text-muted-foreground max-w-sm">
        Completed inventory receipts will appear here once you start receiving stock.
      </p>
    </div>
  </div>
</template>

<script setup>
import { Search, CalendarRange, ClipboardList } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
</script>