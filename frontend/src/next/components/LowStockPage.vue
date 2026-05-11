<template>
  <div class="p-6 lg:p-8 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <Button variant="outline" size="icon" @click="router.back()">
          <ArrowLeft :size="18" />
        </Button>
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Low Stock Items</h1>
          <p class="text-sm text-muted-foreground mt-1">
            {{ items.length }} items need reordering
          </p>
        </div>
      </div>
      <Button @click="exportPDF" :disabled="items.length === 0">
        <Download :size="16" class="mr-2" />
        Export PDF
      </Button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-24">
      <div class="text-muted-foreground">Loading low stock items...</div>
    </div>

    <!-- Table -->
    <Card v-else>
      <CardContent class="p-0">
        <div v-if="items.length === 0" class="text-center py-24 text-muted-foreground text-sm">
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { ArrowLeft, Download } from "lucide-vue-next";
import { useRouter } from "vue-router";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { shared } from "../store.js";
import { useDashboard } from "../composables/useDashboard.js";
import { jsPDF } from "jspdf";
import autoTable from "jspdf-autotable";

const router = useRouter();
const loading = ref(false);
const items = computed(() => shared.lowStockItems);

const { fetchDashboard, dashboardData } = useDashboard();

onMounted(async () => {
  if (items.value.length === 0) {
    loading.value = true;
    await fetchDashboard("today");
    shared.lowStockItems = dashboardData.value?.low_stock_items ?? [];
    loading.value = false;
  }
});

function exportPDF() {
  const doc = new jsPDF();
  doc.setFontSize(16);
  doc.text("Low Stock Items", 14, 20);
  doc.setFontSize(10);
  doc.text(`${items.value.length} items need reordering`, 14, 28);
  doc.text(`Generated: ${new Date().toLocaleDateString()}`, 14, 34);

  autoTable(doc, {
    startY: 40,
    head: [["Medication", "Manufacturer", "Current Stock", "Reorder Level"]],
    body: items.value.map((item) => [
      item.product_name,
      item.manufacturer || "-",
      item.current_stock,
      item.reorder_level,
    ]),
    headStyles: { fillColor: [239, 68, 68] },
    styles: { fontSize: 9 },
    tableWidth: "auto",
    columnStyles: {
      2: { halign: "center" },
      3: { halign: "center" },
    },
  });

  doc.save("low-stock-items.pdf");
}
</script>
