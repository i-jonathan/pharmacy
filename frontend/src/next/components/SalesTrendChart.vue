<template>
  <Card>
    <CardHeader>
      <CardTitle class="text-sm font-semibold uppercase tracking-wider">
        Sales Trend
      </CardTitle>
    </CardHeader>
    <CardContent>
      <div class="relative h-64">
        <Bar v-if="hasData" :data="chartData" :options="chartOptions" />
        <div v-else class="flex items-center justify-center h-full text-muted-foreground text-sm">
          No sales data available for this period
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<script setup>
import { computed } from "vue";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from "chart.js";
import { Bar } from "vue-chartjs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend);

const props = defineProps({
  data: { type: Array, default: () => [] },
});

const hasData = computed(() => props.data.some((d) => d.sales > 0));

const chartData = computed(() => ({
  labels: props.data.map((d) => d.day),
  datasets: [
    {
      label: "Sales",
      data: props.data.map((d) => d.sales / 100),
      backgroundColor: "rgba(99, 102, 241, 0.85)",
      borderRadius: 8,
      borderSkipped: false,
    },
  ],
}));

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx) => {
          const naira = ctx.parsed.y;
          return `Sales: ₦${naira.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
        },
      },
    },
  },
  scales: {
    x: {
      grid: { display: false },
      ticks: { color: "#94a3b8", font: { size: 12 } },
    },
    y: {
      beginAtZero: true,
      grid: { color: "rgba(148, 163, 184, 0.12)" },
      ticks: {
        color: "#94a3b8",
        font: { size: 12 },
        callback: (v) => `₦${v.toLocaleString()}`,
      },
    },
  },
};
</script>
