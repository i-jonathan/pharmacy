<template>
  <Card>
    <CardHeader>
      <div class="flex items-center justify-between">
        <CardTitle class="text-sm font-semibold uppercase tracking-wider">
          Low Stock Items
        </CardTitle>
        <Badge variant="destructive" class="text-xs">
          {{ items.length }} items
        </Badge>
      </div>
    </CardHeader>
    <CardContent>
      <div v-if="items.length === 0" class="text-center py-12 text-muted-foreground text-sm">
        All items are adequately stocked
      </div>

      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead>Medication</TableHead>
            <TableHead>Current Stock</TableHead>
            <TableHead>Reorder Level</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="item in items" :key="item.id">
            <TableCell>
              <div class="text-sm font-medium">{{ item.product_name }}</div>
              <div v-if="item.manufacturer" class="text-xs text-muted-foreground">{{ item.manufacturer }}</div>
            </TableCell>
            <TableCell class="font-semibold">{{ item.current_stock }}</TableCell>
            <TableCell class="text-muted-foreground">{{ item.reorder_level }}</TableCell>
            <TableCell>
              <Badge variant="destructive">Low Stock</Badge>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </CardContent>
  </Card>
</template>

<script setup>
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

defineProps({
  items: { type: Array, default: () => [] },
});
</script>
