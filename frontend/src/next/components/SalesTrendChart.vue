<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between pb-2">
      <CardTitle class="text-sm font-semibold uppercase tracking-wider">
        Sales Overview
      </CardTitle>
      <Select v-model="selectedPeriod" @update:model-value="onPeriodChange">
        <SelectTrigger class="w-36 h-8 text-xs">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="this_week">This Week</SelectItem>
          <SelectItem value="this_month">This Month</SelectItem>
          <SelectItem value="last_week">Last Week</SelectItem>
          <SelectItem value="last_month">Last Month</SelectItem>
        </SelectContent>
      </Select>
    </CardHeader>
    <CardContent>
      <div class="relative h-64">
        <Line v-if="hasData" :data="chartData" :options="chartOptions" />
        <div v-else class="flex items-center justify-center h-full text-muted-foreground text-sm">
          No sales data available for this period
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<script setup>
import { ref, computed, watch, onMounted } from "vue";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from "chart.js";
import { Line } from "vue-chartjs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler);

const emit = defineEmits(["period-change"]);
const selectedPeriod = ref("this_week");
const trendData = ref([]);

const hasData = computed(() => trendData.value.some((d) => d.sales > 0));

const chartData = computed(() => ({
  labels: trendData.value.map((d) => d.date),
  datasets: [
    {
      label: "Sales",
      data: trendData.value.map((d) => d.sales / 100),
      borderColor: "rgb(99, 102, 241)",
      backgroundColor: "rgba(99, 102, 241, 0.08)",
      pointBackgroundColor: "rgb(99, 102, 241)",
      pointBorderColor: "hsl(var(--card))",
      pointBorderWidth: 2,
      pointRadius: 3,
      pointHoverRadius: 5,
      tension: 0.35,
      fill: true,
      borderWidth: 2,
    },
  ],
}));

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: "index",
  },
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
      ticks: { color: "#94a3b8", font: { size: 11 }, maxRotation: 0 },
    },
    y: {
      beginAtZero: true,
      grid: { color: "rgba(148, 163, 184, 0.1)" },
      ticks: {
        color: "#94a3b8",
        font: { size: 11 },
        maxTicksLimit: 5,
        callback: (v) => {
            if (v >= 1000) {
              const k = v / 1000;
              return `₦${k % 1 === 0 ? k.toFixed(0) : k.toFixed(1)}k`;
            }
            return `₦${v}`;
          },
      },
    },
  },
};

function getPeriodRange(period) {
  const now = new Date();
  const fmt = (d) => d.toISOString().slice(0, 10);
  switch (period) {
    case "this_week": {
      const start = new Date(now);
      start.setDate(now.getDate() - now.getDay() + 1); // Monday
      return { start_date: fmt(start), end_date: fmt(now) };
    }
    case "this_month": {
      const start = new Date(now.getFullYear(), now.getMonth(), 1);
      return { start_date: fmt(start), end_date: fmt(now) };
    }
    case "last_week": {
      const lastMonday = new Date(now);
      lastMonday.setDate(now.getDate() - now.getDay() - 6);
      const lastSunday = new Date(lastMonday);
      lastSunday.setDate(lastMonday.getDate() + 6);
      return { start_date: fmt(lastMonday), end_date: fmt(lastSunday) };
    }
    case "last_month": {
      const start = new Date(now.getFullYear(), now.getMonth() - 1, 1);
      const end = new Date(now.getFullYear(), now.getMonth(), 0);
      return { start_date: fmt(start), end_date: fmt(end) };
    }
    default:
      return { start_date: fmt(now), end_date: fmt(now) };
  }
}

async function fetchTrend(period) {
  try {
    const range = getPeriodRange(period);
    const params = new URLSearchParams(range);
    const res = await fetch(`/api/dashboard?${params}`);
    if (res.ok) {
      const data = await res.json();
      trendData.value = data.sales_trend ?? [];
    }
  } catch (e) {
    console.error("Failed to fetch sales trend:", e);
  }
}

function onPeriodChange(period) {
  selectedPeriod.value = period;
  fetchTrend(period);
}

onMounted(() => {
  fetchTrend(selectedPeriod.value);
});
</script>
