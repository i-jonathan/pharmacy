<template>
  <Card>
    <CardHeader>
      <CardTitle class="text-sm font-semibold uppercase tracking-wider">
        Recent Transactions
      </CardTitle>
    </CardHeader>
    <CardContent>
      <div v-if="transactions.length === 0" class="text-center py-12 text-muted-foreground text-sm">
        No transactions for this period
      </div>

      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead>Receipt</TableHead>
            <TableHead>Items</TableHead>
            <TableHead>Total</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="txn in transactions" :key="txn.id">
            <TableCell>
              <div class="text-sm font-medium font-mono">{{ txn.receipt_number }}</div>
              <div class="text-xs text-muted-foreground">{{ formatTime(txn.created_at) }}</div>
            </TableCell>
            <TableCell class="text-muted-foreground">{{ txn.item_count }}</TableCell>
            <TableCell class="font-semibold">{{ formatNaira(txn.total) }}</TableCell>
            <TableCell>
              <Badge :variant="statusBadge(txn.status).variant" :class="['text-xs', statusBadge(txn.status).class]">{{ statusBadge(txn.status).label }}</Badge>
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
  transactions: { type: Array, default: () => [] },
});

function formatNaira(kobo) {
  const naira = kobo / 100;
  return `₦${naira.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatTime(dateStr) {
  const d = new Date(dateStr);
  return d.toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit", hour12: true });
}

function statusBadge(status) {
  const label = status.charAt(0) + status.slice(1).toLowerCase();
  if (status === "COMPLETED") {
    return { label, variant: "outline", class: "border-emerald-300 text-emerald-600 dark:border-emerald-700 dark:text-emerald-400" };
  }
  return { label, variant: "secondary", class: "" };
}
</script>
