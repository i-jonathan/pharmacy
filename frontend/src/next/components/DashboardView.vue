<template>
  <div class="p-6 lg:p-8 space-y-6">
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p class="text-sm text-muted-foreground mt-1">
          Overview of your pharmacy operations
        </p>
      </div>
      <DateFilterBar v-model="dateFilter" @update:model-value="onFilterChange" />
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-24">
      <div class="text-muted-foreground">Loading dashboard data...</div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="flex flex-col items-center justify-center py-24">
      <div class="text-destructive mb-4">{{ error }}</div>
      <Button variant="outline" @click="fetchDashboard(dateFilter)">Retry</Button>
    </div>

    <!-- Dashboard Content -->
    <template v-else>
      <!-- KPI Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <MetricCard
          title="Total Sales"
          :formatted-value="formatNaira(getKPIValue('today_sales', 0))"
          :trend="getKPIValue('sales_trend', null)"
          :icon="DollarSign"
          accent="indigo"
        />
        <MetricCard
          title="Total Orders"
          :formatted-value="getKPIValue('today_transactions', 0).toLocaleString()"
          :trend="getKPIValue('transaction_trend', null)"
          :icon="ShoppingCart"
          accent="emerald"
        />
        <MetricCard
          title="Total Products"
          :formatted-value="getKPIValue('total_inventory', 0).toLocaleString()"
          subtitle="Products in inventory"
          :icon="Package"
          accent="sky"
        />
        <MetricCard
          title="Low Stock Items"
          :formatted-value="getKPIValue('low_stock_count', 0).toLocaleString()"
          subtitle="Need reordering"
          :icon="AlertTriangle"
          accent="amber"
        />
        <MetricCard
          title="Expiring Soon"
          :formatted-value="getKPIValue('expiring_count', 0).toLocaleString()"
          subtitle="Within 90 days"
          :icon="Calendar"
          accent="rose"
        />
      </div>

      <!-- Charts Row -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <SalesTrendChart />
        <TopSellingProducts :data="dashboardData?.top_selling_products ?? []" />
      </div>

      <!-- Expiry + Low Stock Row -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ExpiryAlerts :data="dashboardData?.expiry_by_category ?? []" />
        <LowStockTable :items="dashboardData?.low_stock_items ?? []" :limit="5" @view-all="goToLowStock" />
      </div>
    </template>
  </div>
</template>

<script setup>
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import { DollarSign, ShoppingCart, Package, AlertTriangle, Calendar } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { useDashboard } from "../composables/useDashboard.js";
import { shared } from "../store.js";
import MetricCard from "./MetricCard.vue";
import SalesTrendChart from "./SalesTrendChart.vue";
import TopSellingProducts from "./TopSellingProducts.vue";
import ExpiryAlerts from "./ExpiryAlerts.vue";
import LowStockTable from "./LowStockTable.vue";
import DateFilterBar from "./DateFilterBar.vue";

const router = useRouter();
const { loading, error, dashboardData, dateFilter, fetchDashboard, getKPIValue } = useDashboard();

function formatNaira(kobo) {
  const naira = kobo / 100;
  return `₦${naira.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function onFilterChange(filter) {
  fetchDashboard(filter);
}

function goToLowStock() {
  shared.lowStockItems = dashboardData.value?.low_stock_items ?? [];
  router.push({ name: "low-stock" });
}

onMounted(() => {
  fetchDashboard(dateFilter.value);
});
</script>
