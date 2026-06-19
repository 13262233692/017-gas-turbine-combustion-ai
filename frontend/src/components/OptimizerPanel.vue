<template>
  <div class="optimizer-panel">
    <div class="mode-display">
      <div class="current-mode">
        <span class="mode-label">当前工况</span>
        <span class="mode-value" :class="modeClass">{{ modeLabels[optimization?.current_mode] || '正常' }}</span>
      </div>
      <div class="transition-bar" v-if="optimization?.transition_progress < 1">
        <div class="trans-label">切换进度: {{ (optimization?.transition_progress * 100 || 0).toFixed(0) }}%</div>
        <div class="trans-track">
          <div class="trans-fill" :style="{ width: (optimization?.transition_progress || 0) * 100 + '%' }"></div>
        </div>
      </div>
    </div>

    <div class="optim-metrics">
      <div class="metric-item">
        <div class="metric-circle">
          <svg viewBox="0 0 100 100">
            <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
            <circle cx="50" cy="50" r="42" fill="none" stroke="#00ff88" stroke-width="8"
              :stroke-dasharray="optEffArc + ' ' + (264 - optEffArc)"
              transform="rotate(-90 50 50)" stroke-linecap="round"/>
          </svg>
          <span class="metric-val">{{ (optimization?.optimized_efficiency * 100 || 0).toFixed(1) }}%</span>
        </div>
        <label>优化效率</label>
      </div>
      <div class="metric-item">
        <div class="metric-val-lg" :class="{ positive: (optimization?.efficiency_gain || 0) > 0 }">
          {{ (optimization?.efficiency_gain * 100 || 0).toFixed(2) }}%
        </div>
        <label>效率增益</label>
      </div>
    </div>

    <div class="power-info">
      <div class="info-row">
        <span class="info-label">输出功率</span>
        <span class="info-value">{{ (optimization?.power_output || 0).toFixed(1) }} MW</span>
      </div>
      <div class="info-row">
        <span class="info-label">负载率</span>
        <div class="load-bar">
          <div class="load-fill" :style="{ width: (optimization?.load_fraction || 0) * 100 + '%' }" :class="loadClass"></div>
        </div>
        <span class="info-value">{{ (optimization?.load_fraction * 100 || 0).toFixed(0) }}%</span>
      </div>
    </div>

    <div class="mode-switch">
      <label>工况切换</label>
      <div class="mode-grid">
        <button v-for="m in operatingModes" :key="m.value"
          class="switch-btn" :class="{ active: optimization?.current_mode === m.value }"
          @click="$emit('setMode', m.value)">
          <span class="switch-icon">{{ m.icon }}</span>
          <span class="switch-text">{{ m.label }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'OptimizerPanel',
  props: {
    optimization: Object,
  },
  emits: ['setMode'],
  setup(props) {
    const optEffArc = computed(() => (props.optimization?.optimized_efficiency || 0) * 264)
    const loadClass = computed(() => {
      const f = props.optimization?.load_fraction || 0
      if (f > 0.9) return 'high'
      if (f > 0.5) return 'normal'
      return 'low'
    })
    const modeClass = computed(() => {
      const m = props.optimization?.current_mode
      if (m === 'emergency') return 'emergency'
      if (m === 'peak_load') return 'peak'
      return 'normal'
    })
    const modeLabels = {
      startup: '启动', normal: '正常', peak_load: '峰值',
      low_load: '低负荷', shutdown: '停机', emergency: '紧急',
    }
    const operatingModes = [
      { value: 'startup', label: '启动', icon: '🔄' },
      { value: 'normal', label: '正常', icon: '✅' },
      { value: 'peak_load', label: '峰值', icon: '⚡' },
      { value: 'low_load', label: '低负荷', icon: '📉' },
      { value: 'shutdown', label: '停机', icon: '⏹' },
      { value: 'emergency', label: '紧急', icon: '🚨' },
    ]
    return { optEffArc, loadClass, modeClass, modeLabels, operatingModes }
  },
}
</script>

<style scoped>
.optimizer-panel { height: 100%; display: flex; flex-direction: column; }
.mode-display { margin-bottom: 12px; }
.current-mode { display: flex; justify-content: space-between; align-items: center; }
.mode-label { font-size: 12px; color: var(--text-secondary); }
.mode-value { font-size: 16px; font-weight: 700; padding: 2px 10px; border-radius: 4px; }
.mode-value.normal { color: var(--accent-green); background: rgba(0,255,136,0.1); }
.mode-value.peak { color: var(--accent-yellow); background: rgba(255,170,0,0.1); }
.mode-value.emergency { color: var(--accent-red); background: rgba(255,51,68,0.1); animation: pulse 1s infinite; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.6; } }
.transition-bar { margin-top: 8px; }
.trans-label { font-size: 11px; color: var(--text-secondary); margin-bottom: 4px; }
.trans-track { height: 4px; background: #1a2744; border-radius: 2px; }
.trans-fill { height: 100%; background: var(--accent-cyan); border-radius: 2px; transition: width 0.5s; }
.optim-metrics { display: flex; gap: 12px; margin-bottom: 12px; }
.metric-item { flex: 1; text-align: center; }
.metric-circle { position: relative; width: 72px; height: 72px; margin: 0 auto 4px; }
.metric-circle svg { width: 100%; height: 100%; }
.metric-val { position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); font-size: 13px; font-weight: 600; }
.metric-item label { font-size: 11px; color: var(--text-secondary); }
.metric-val-lg { font-size: 22px; font-weight: 700; color: var(--text-primary); text-align: center; margin: 14px 0 4px; }
.metric-val-lg.positive { color: var(--accent-green); }
.power-info { background: rgba(0,0,0,0.2); border-radius: 6px; padding: 8px 12px; margin-bottom: 12px; }
.info-row { display: flex; align-items: center; justify-content: space-between; padding: 4px 0; font-size: 12px; }
.info-label { color: var(--text-secondary); }
.info-value { color: var(--accent-cyan); font-weight: 500; }
.load-bar { flex: 1; height: 6px; background: #1a2744; border-radius: 3px; margin: 0 8px; }
.load-fill { height: 100%; border-radius: 3px; transition: width 0.5s; }
.load-fill.low { background: var(--accent-cyan); }
.load-fill.normal { background: var(--accent-green); }
.load-fill.high { background: var(--accent-yellow); }
.mode-switch label { display: block; font-size: 11px; color: var(--text-secondary); margin-bottom: 6px; }
.mode-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 4px; }
.switch-btn { padding: 6px 4px; border: 1px solid var(--border-color); background: transparent; color: var(--text-secondary); border-radius: 4px; cursor: pointer; font-size: 10px; transition: all 0.3s; display: flex; flex-direction: column; align-items: center; gap: 2px; }
.switch-btn.active { border-color: var(--accent-cyan); color: var(--accent-cyan); background: rgba(0,212,255,0.1); }
.switch-btn:hover { border-color: var(--accent-cyan); }
.switch-icon { font-size: 14px; }
</style>
