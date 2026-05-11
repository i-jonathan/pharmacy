<template>
  <div class="p-6 lg:p-8 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <Button variant="outline" size="icon" @click="router.back()">
          <ArrowLeft :size="18" />
        </Button>
        <span class="text-sm text-muted-foreground">
          {{ items.length }} items expiring within 90 days
        </span>
      </div>
      <Button @click="exportPDF" :disabled="items.length === 0">
        <Download :size="16" class="mr-2" />
        Export PDF
      </Button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-24">
      <div class="text-muted-foreground">Loading expiring items...</div>
    </div>

    <!-- Table -->
    <Card v-else>
      <CardContent class="p-0">
        <div v-if="items.length === 0" class="text-center py-24 text-muted-foreground text-sm">
          No items expiring within 90 days
        </div>

        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Product</TableHead>
              <TableHead>Quantity</TableHead>
              <TableHead>Expiry Date</TableHead>
              <TableHead>Days Left</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="item in items" :key="item.id">
              <TableCell>
                <div class="text-sm font-medium">{{ item.product_name }}</div>
              </TableCell>
              <TableCell class="font-semibold">{{ item.quantity }}</TableCell>
              <TableCell class="text-muted-foreground">
                {{ formatDate(item.expiry_date) }}
              </TableCell>
              <TableCell>
                <Badge :variant="item.days_until_expiry <= 30 ? 'destructive' : 'secondary'">
                  {{ item.days_until_expiry }} days
                </Badge>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
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
const items = computed(() => shared.expiringItems);

const { fetchDashboard, dashboardData } = useDashboard();

onMounted(async () => {
  if (items.value.length === 0) {
    loading.value = true;
    await fetchDashboard("today");
    shared.expiringItems = dashboardData.value?.expiring_items ?? [];
    loading.value = false;
  }
});

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleDateString();
}

function exportPDF() {
  const doc = new jsPDF();
  doc.setFontSize(16);
  doc.text("Expiring Items", 14, 20);
  doc.setFontSize(10);
  doc.text(`${items.value.length} items expiring within 90 days`, 14, 28);
  doc.text(`Generated: ${new Date().toLocaleDateString()}`, 14, 34);

  autoTable(doc, {
    startY: 40,
    head: [["Product", "Quantity", "Expiry Date", "Days Left"]],
    body: items.value.map((item) => [
      item.product_name,
      item.quantity,
      formatDate(item.expiry_date),
      `${item.days_until_expiry} days`,
    ]),
    headStyles: { fillColor: [239, 68, 68] },
    styles: { fontSize: 9 },
    tableWidth: "auto",
    columnStyles: {
      1: { halign: "center" },
      2: { halign: "center" },
      3: { halign: "center" },
    },
  });

  doc.save("expiring-items.pdf");
}
</script>
