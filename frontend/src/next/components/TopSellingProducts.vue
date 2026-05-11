<template>
  <Card>
    <CardHeader>
      <CardTitle class="text-sm font-semibold uppercase tracking-wider">
        Top Selling Products
      </CardTitle>
      <CardDescription>Past 7 days</CardDescription>
    </CardHeader>
    <CardContent>
      <div v-if="products.length === 0" class="text-center py-8 text-muted-foreground text-sm">
        No sales data for this week
      </div>
      <template v-else>
        <div class="flex items-center gap-3 mb-2 px-1">
          <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider flex-1">Product</span>
          <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider w-14 text-center">Sold</span>
          <span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider w-24 text-right">Revenue</span>
        </div>
        <div class="space-y-1">
          <div
            v-for="(item, index) in products"
            :key="item.product_name"
            class="flex items-center gap-3 py-2"
          >
            <span class="w-6 h-6 rounded-full bg-muted flex items-center justify-center text-xs font-semibold text-muted-foreground shrink-0">
              {{ index + 1 }}
            </span>
            <span class="text-sm font-medium truncate flex-1">{{ item.product_name }}</span>
            <span class="text-sm text-muted-foreground w-14 text-center">{{ item.quantity }}</span>
            <span class="text-sm font-semibold w-24 text-right">{{ formatNaira(item.revenue_kobo) }}</span>
          </div>
        </div>
      </template>
    </CardContent>
  </Card>
</template>

<script setup>
import { computed } from "vue";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

const props = defineProps({
  data: { type: Array, default: () => [] },
});

const products = computed(() => props.data);

function formatNaira(kobo) {
  const naira = kobo / 100;
  return `₦${naira.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}
</script>
