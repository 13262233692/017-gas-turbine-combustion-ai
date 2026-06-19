<template>
  <div class="heatmap-container" ref="container">
    <canvas ref="canvas" :width="canvasSize" :height="canvasSize"></canvas>
  </div>
</template>

<script>
import { ref, watch, onMounted } from 'vue'

const COLOR_STOPS = [
  { pos: 0.0, r: 10, g: 0, b: 40 },
  { pos: 0.15, r: 20, g: 0, b: 120 },
  { pos: 0.3, r: 0, g: 50, b: 200 },
  { pos: 0.45, r: 0, g: 150, b: 200 },
  { pos: 0.55, r: 0, g: 210, b: 100 },
  { pos: 0.65, r: 180, g: 230, b: 0 },
  { pos: 0.75, r: 255, g: 200, b: 0 },
  { pos: 0.85, r: 255, g: 100, b: 0 },
  { pos: 0.95, r: 220, g: 20, b: 0 },
  { pos: 1.0, r: 255, g: 255, b: 255 },
]

function tempToColor(t) {
  let lower = COLOR_STOPS[0], upper = COLOR_STOPS[COLOR_STOPS.length - 1]
  for (let i = 0; i < COLOR_STOPS.length - 1; i++) {
    if (t >= COLOR_STOPS[i].pos && t <= COLOR_STOPS[i + 1].pos) {
      lower = COLOR_STOPS[i]
      upper = COLOR_STOPS[i + 1]
      break
    }
  }
  const range = upper.pos - lower.pos
  const f = range === 0 ? 0 : (t - lower.pos) / range
  return {
    r: Math.round(lower.r + (upper.r - lower.r) * f),
    g: Math.round(lower.g + (upper.g - lower.g) * f),
    b: Math.round(lower.b + (upper.b - lower.b) * f),
  }
}

export default {
  name: 'TemperatureHeatmap',
  props: {
    fieldData: { type: Object, default: null },
  },
  setup(props) {
    const canvas = ref(null)
    const container = ref(null)
    const canvasSize = ref(400)

    const drawHeatmap = () => {
      if (!canvas.value || !props.fieldData || !props.fieldData.grid) return
      const ctx = canvas.value.getContext('2d')
      const grid = props.fieldData.grid
      const rows = props.fieldData.rows
      const cols = props.fieldData.cols
      const size = canvasSize.value

      let minVal = Infinity, maxVal = -Infinity
      for (let i = 0; i < rows; i++) {
        for (let j = 0; j < cols; j++) {
          if (grid[i][j] < minVal) minVal = grid[i][j]
          if (grid[i][j] > maxVal) maxVal = grid[i][j]
        }
      }
      const range = maxVal - minVal || 1

      const cellW = size / cols
      const cellH = size / rows

      ctx.clearRect(0, 0, size, size)

      for (let i = 0; i < rows; i++) {
        for (let j = 0; j < cols; j++) {
          const normalized = (grid[i][j] - minVal) / range
          const color = tempToColor(normalized)
          ctx.fillStyle = `rgb(${color.r},${color.g},${color.b})`
          ctx.fillRect(j * cellW, i * cellH, cellW + 1, cellH + 1)
        }
      }

      const gradient = ctx.createLinearGradient(size, 0, size + 30, 0)
      COLOR_STOPS.forEach(s => {
        gradient.addColorStop(s.pos, `rgb(${s.r},${s.g},${s.b})`)
      })
      ctx.fillStyle = gradient
      ctx.fillRect(size + 5, 0, 20, size)

      ctx.fillStyle = '#8899aa'
      ctx.font = '10px monospace'
      const steps = 5
      for (let i = 0; i <= steps; i++) {
        const y = (i / steps) * size
        const val = maxVal - (i / steps) * range
        ctx.fillText(val.toFixed(0) + 'K', size + 28, y + 4)
      }
    }

    watch(() => props.fieldData, drawHeatmap, { deep: true })
    onMounted(drawHeatmap)

    return { canvas, container, canvasSize }
  },
}
</script>

<style scoped>
.heatmap-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 10px;
}
canvas {
  border-radius: 8px;
  image-rendering: pixelated;
}
</style>
