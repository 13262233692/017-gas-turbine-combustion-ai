<template>
  <div class="combustion-gauge">
    <div class="gauge-wrapper">
      <svg viewBox="0 0 200 200" class="gauge-svg">
        <circle cx="100" cy="100" r="85" fill="none" stroke="#1a2744" stroke-width="12" />
        <circle cx="100" cy="100" r="85" fill="none"
          :stroke="gaugeColor" stroke-width="12"
          :stroke-dasharray="gaugeArc + ' ' + (534 - gaugeArc)"
          transform="rotate(-90 100 100)" stroke-linecap="round"
          class="gauge-progress" />
        <text x="100" y="90" text-anchor="middle" fill="white" font-size="28" font-weight="bold">
          {{ instabilityPercent }}%
        </text>
        <text x="100" y="115" text-anchor="middle" fill="#8899aa" font-size="12">
          不稳定指数
        </text>
      </svg>
    </div>
    <div class="gauge-legend">
      <div class="legend-item stable">
        <span class="legend-dot"></span> 稳定 (0-35%)
      </div>
      <div class="legend-item warning">
        <span class="legend-dot"></span> 警告 (35-70%)
      </div>
      <div class="legend-item critical">
        <span class="legend-dot"></span> 危险 (>70%)
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'CombustionGauge',
  props: {
    state: { type: Object, default: null },
  },
  setup(props) {
    const instabilityPercent = computed(() => {
      return ((props.state?.instability_index || 0) * 100).toFixed(1)
    })

    const gaugeArc = computed(() => {
      const val = props.state?.instability_index || 0
      return Math.min(val, 1) * 534
    })

    const gaugeColor = computed(() => {
      const val = props.state?.instability_index || 0
      if (val < 0.35) return '#00ff88'
      if (val < 0.7) return '#ffaa00'
      return '#ff3344'
    })

    return { instabilityPercent, gaugeArc, gaugeColor }
  },
}
</script>

<style scoped>
.combustion-gauge {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.gauge-wrapper {
  width: 180px;
  height: 180px;
}
.gauge-svg {
  width: 100%;
  height: 100%;
}
.gauge-progress {
  transition: stroke-dasharray 0.5s ease, stroke 0.5s ease;
}
.gauge-legend {
  display: flex;
  gap: 16px;
  font-size: 11px;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #8899aa;
}
.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.stable .legend-dot { background: #00ff88; }
.warning .legend-dot { background: #ffaa00; }
.critical .legend-dot { background: #ff3344; }
</style>
