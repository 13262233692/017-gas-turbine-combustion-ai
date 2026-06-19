<template>
  <div class="efficiency-chart" ref="chartRef"></div>
</template>

<script>
import { ref, watch, onMounted } from 'vue'

export default {
  name: 'EfficiencyChart',
  props: {
    efficiency: { type: Object, default: null },
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
        grid: { top: 20, right: 20, bottom: 30, left: 50 },
        xAxis: {
          type: 'category',
          data: ['燃烧效率', '热效率'],
          axisLine: { lineStyle: { color: '#334466' } },
          axisLabel: { color: '#8899aa' },
        },
        yAxis: {
          type: 'value',
          min: 0, max: 100,
          axisLine: { lineStyle: { color: '#334466' } },
          axisLabel: { color: '#8899aa', formatter: '{value}%' },
          splitLine: { lineStyle: { color: '#1a2744' } },
        },
        series: [{
          type: 'bar',
          barWidth: 50,
          data: [
            {
              value: ((props.efficiency?.combustion_efficiency || 0) * 100).toFixed(1),
              itemStyle: {
                color: {
                  type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
                  colorStops: [
                    { offset: 0, color: '#00d4ff' },
                    { offset: 1, color: '#0066cc' },
                  ],
                },
                borderRadius: [4, 4, 0, 0],
              },
            },
            {
              value: ((props.efficiency?.thermal_efficiency || 0) * 100).toFixed(1),
              itemStyle: {
                color: {
                  type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
                  colorStops: [
                    { offset: 0, color: '#00ff88' },
                    { offset: 1, color: '#00aa55' },
                  ],
                },
                borderRadius: [4, 4, 0, 0],
              },
            },
          ],
          label: {
            show: true, position: 'top',
            color: '#ccc', formatter: '{c}%',
          },
        }],
      }
      chart.setOption(option)
    }

    const updateChart = () => {
      if (!chart || !props.efficiency) return
      chart.setOption({
        series: [{
          data: [
            {
              value: ((props.efficiency?.combustion_efficiency || 0) * 100).toFixed(1),
              itemStyle: {
                color: {
                  type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
                  colorStops: [
                    { offset: 0, color: '#00d4ff' },
                    { offset: 1, color: '#0066cc' },
                  ],
                },
                borderRadius: [4, 4, 0, 0],
              },
            },
            {
              value: ((props.efficiency?.thermal_efficiency || 0) * 100).toFixed(1),
              itemStyle: {
                color: {
                  type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
                  colorStops: [
                    { offset: 0, color: '#00ff88' },
                    { offset: 1, color: '#00aa55' },
                  ],
                },
                borderRadius: [4, 4, 0, 0],
              },
            },
          ],
        }],
      })
    }

    onMounted(initChart)
    watch(() => props.efficiency, updateChart, { deep: true })

    return { chartRef }
  },
}
</script>

<style scoped>
.efficiency-chart {
  width: 100%;
  height: 200px;
}
</style>
