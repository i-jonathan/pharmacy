<template>
  <div class="relative h-64">
    <Doughnut :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
  Legend
} from 'chart.js'
import { Doughnut } from 'vue-chartjs'

ChartJS.register(ArcElement, Tooltip, Legend)

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
      position: 'bottom',
      labels: {
        padding: 15,
        usePointStyle: true
      }
    },
    tooltip: {
      callbacks: {
        label: function(context) {
          return context.label + ': ' + context.parsed + '%'
        }
      }
    }
  },
  cutout: '60%'
}

const updateChartData = () => {
  const colors = [
    'rgb(59, 130, 246)',   // blue
    'rgb(34, 197, 94)',    // green
    'rgb(168, 85, 247)',   // purple
    'rgb(251, 146, 60)',   // orange
    'rgb(107, 114, 128)'   // gray
  ]

  chartData.value = {
    labels: props.data.map(item => item.category),
    datasets: [
      {
        data: props.data.map(item => item.sales),
        backgroundColor: colors,
        borderWidth: 2,
        borderColor: '#ffffff'
      }
    ]
  }
}

watch(() => props.data, updateChartData, { immediate: true })

onMounted(() => {
  updateChartData()
})
</script>
