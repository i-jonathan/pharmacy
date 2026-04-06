<template>
  <div class="relative h-64">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const props = defineProps({
  data: {
    type: Array,
    required: true
  }
})

const chartData = ref({
  labels: [],
  datasets: []
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: function(context) {
          return 'Sales: ₦' + context.parsed.y.toLocaleString()
        }
      }
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      ticks: {
        callback: function(value) {
          return '₦' + (value / 1000) + 'k'
        }
      }
    }
  },
  elements: {
    line: {
      tension: 0.4,
      borderWidth: 3
    },
    point: {
      radius: 5,
      hoverRadius: 7
    }
  }
}

const updateChartData = () => {
  chartData.value = {
    labels: props.data.map(item => item.day),
    datasets: [
      {
        label: 'Sales',
        data: props.data.map(item => item.sales),
        borderColor: 'rgb(59, 130, 246)',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        fill: true
      }
    ]
  }
}

watch(() => props.data, updateChartData, { immediate: true })

onMounted(() => {
  updateChartData()
})
</script>
