<template>
  <Card>
    <CardHeader>
      <div class="flex items-center justify-between">
        <CardTitle class="text-sm font-semibold uppercase tracking-wider">
          Top Selling Products
        </CardTitle>
        <span class="text-xs text-muted-foreground">Past 7 days</span>
      </div>
    </CardHeader>
    <CardContent>
      <div v-if="products.length === 0" class="text-center py-12 text-muted-foreground text-sm">
        No sales data for this period
      </div>

      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead class="w-10">#</TableHead>
            <TableHead>Product</TableHead>
            <TableHead class="text-center">Sold</TableHead>
            <TableHead class="text-right">Revenue</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(item, index) in products" :key="item.product_name">
            <TableCell class="text-muted-foreground text-sm">{{ index + 1 }}</TableCell>
            <TableCell class="font-medium">{{ item.product_name }}</TableCell>
            <TableCell class="text-center text-muted-foreground">{{ item.quantity }}</TableCell>
            <TableCell class="text-right font-semibold">{{ formatNaira(item.revenue_kobo) }}</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </CardContent>
  </Card>
</template>

<script setup>
import { computed } from "vue";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const props = defineProps({
  data: { type: Array, default: () => [] },
});

const products = computed(() => props.data);

function formatNaira(kobo) {
  const naira = kobo / 100;
  return `₦${naira.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}
</script>
