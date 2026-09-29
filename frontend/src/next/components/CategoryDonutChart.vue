<template>
  <Card>
    <CardHeader>
      <CardTitle class="text-sm font-semibold uppercase tracking-wider">
        Sales by Category
      </CardTitle>
    </CardHeader>
    <CardContent>
      <div class="relative h-64">
        <Doughnut v-if="hasData" :data="chartData" :options="chartOptions" />
        <div v-else class="flex items-center justify-center h-full text-muted-foreground text-sm">
          No category data available
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<script setup>
import { computed } from "vue";
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from "chart.js";
import { Doughnut } from "vue-chartjs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

ChartJS.register(ArcElement, Tooltip, Legend);

const props = defineProps({
  data: { type: Array, default: () => [] },
});

const hasData = computed(() => props.data.some((d) => d.value > 0));

const colors = [
  "rgb(99, 102, 241)",   // indigo
  "rgb(16, 185, 129)",   // emerald
  "rgb(245, 158, 11)",   // amber
  "rgb(239, 68, 68)",    // red
  "rgb(14, 165, 233)",   // sky
  "rgb(168, 85, 247)",   // purple
  "rgb(251, 146, 60)",   // orange
  "rgb(148, 163, 184)",  // slate
];

const chartData = computed(() => ({
  labels: props.data.map((d) => d.name),
  datasets: [
    {
      data: props.data.map((d) => d.value),
      backgroundColor: colors.slice(0, props.data.length),
      borderColor: "hsl(var(--card))",
      borderWidth: 2,
      hoverBorderWidth: 3,
    },
  ],
}));

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: "65%",
  plugins: {
    legend: {
      position: "right",
      labels: {
        padding: 16,
        usePointStyle: true,
        pointStyleWidth: 10,
        font: { size: 12 },
        color: "#94a3b8",
      },
    },
    tooltip: {
      callbacks: {
        label: (ctx) => `${ctx.label}: ${ctx.parsed}%`,
      },
    },
  },
};
</script>
