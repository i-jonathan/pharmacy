<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Products</h1>
      <p class="text-sm text-muted-foreground mt-1">Manage your pharmacy inventory items</p>
    </div>

    <!-- Toolbar -->
    <div class="flex flex-col sm:flex-row gap-3 mb-6">
      <div class="relative flex-1">
        <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search by product name, manufacturer, or SKU..."
          class="w-full pl-9 pr-4 py-2 text-sm border border-border rounded-md bg-background text-foreground placeholder:text-muted-foreground outline-none focus:ring-1 focus:ring-ring"
        />
      </div>
      <div class="flex gap-2">
        <select class="px-3 py-2 text-sm border border-border rounded-md bg-background text-foreground outline-none focus:ring-1 focus:ring-ring">
          <option>All Categories</option>
          <option>Tablets</option>
          <option>Capsules</option>
          <option>Syrups</option>
          <option>Injections</option>
          <option>Topicals</option>
        </select>
        <Button variant="outline">
          <Filter :size="16" class="mr-2" />
          Filters
        </Button>
        <Button>
          <Plus :size="16" class="mr-2" />
          Add Product
        </Button>
      </div>
    </div>

    <!-- Table placeholder -->
    <div class="rounded-lg border border-border bg-card overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="border-b border-border bg-muted/30">
            <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Name</th>
            <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">SKU</th>
            <th class="text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Category</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Price</th>
            <th class="text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-3">Stock</th>
          </tr>
        </thead>
        <tbody>
          <tr class="border-b border-border/50 hover:bg-muted/20 transition-colors" v-for="i in 5" :key="i">
            <td class="px-4 py-3 text-sm text-foreground">Product {{ i }}</td>
            <td class="px-4 py-3 text-sm text-muted-foreground">SKU-{{ String(i).padStart(4, '0') }}</td>
            <td class="px-4 py-3 text-sm text-muted-foreground">Category</td>
            <td class="px-4 py-3 text-sm text-right font-medium">&#8358;0.00</td>
            <td class="px-4 py-3 text-sm text-right">
              <span class="bg-muted text-muted-foreground px-2 py-0.5 rounded-full text-xs font-medium">0</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { Search, Filter, Plus } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
</script>