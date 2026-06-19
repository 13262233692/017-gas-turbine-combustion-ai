<template>
  <div class="trend-chart" ref="chartRef"></div>
</template>

<script>
import { ref, watch, onMounted } from 'vue'

export default {
  name: 'TemperatureTrend',
  props: {
    history: { type: Array, default: () => [] },
  },
  setup(props) {
    const chartRef = ref(null)
    let chart = null

    const initChart = async () => {
      if (!chartRef.value) return
      const echarts = await import('echarts')
      chart = echarts.init(chartRef.value, 'dark')

      const option = {
        backgroundColor: 'transparent',
        tooltip: { trigger: 'axis' },
        legend: {
          data: ['最高温度', '平均温度', '最低温度'],
          textStyle: { color: '#8899aa', fontSize: 11 },
          top: 0,
        },
        grid: { top: 35, right: 15, bottom: 25, left: 55 },
        xAxis: {
          type: 'category',
          data: [],
          axisLine: { lineStyle: { color: '#334466' } },
          axisLabel: { color: '#667788', fontSize: 10 },
        },
        yAxis: {
          type: 'value',
          axisLine: { lineStyle: { color: '#334466' } },
          axisLabel: { color: '#8899aa', formatter: '{value}K' },
          splitLine: { lineStyle: { color: '#1a2744' } },
        },
        series: [
          {
            name: '最高温度',
            type: 'line',
            smooth: true,
            symbol: 'none',
            lineStyle: { color: '#ff5566', width: 2 },
            areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(255,85,102,0.3)' }, { offset: 1, color: 'rgba(255,85,102,0)' }] } },
            data: [],
          },
          {
            name: '平均温度',
            type: 'line',
            smooth: true,
            symbol: 'none',
            lineStyle: { color: '#00d4ff', width: 2 },
            data: [],
          },
          {
            name: '最低温度',
            type: 'line',
            smooth: true,
            symbol: 'none',
            lineStyle: { color: '#00ff88', width: 2 },
            areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(0,255,136,0)' }, { offset: 1, color: 'rgba(0,255,136,0.2)' }] } },
            data: [],
          },
        ],
      }
      chart.setOption(option)
    }

    const updateChart = () => {
      if (!chart || !props.history.length) return
      const times = props.history.map(h => {
        const d = new Date(h.time)
        return d.toLocaleTimeString('zh-CN', { minute: '2-digit', second: '2-digit' })
      })
      const maxData = props.history.map(h => h.max?.toFixed(1))
      const avgData = props.history.map(h => h.avg?.toFixed(1))
      const minData = props.history.map(h => h.min?.toFixed(1))

      chart.setOption({
        xAxis: { data: times },
        series: [
          { data: maxData },
          { data: avgData },
          { data: minData },
        ],
      })
    }

    onMounted(initChart)
    watch(() => props.history, updateChart, { deep: true })

    return { chartRef }
  },
}
</script>

<style scoped>
.trend-chart {
  width: 100%;
  height: 280px;
}
</style>
