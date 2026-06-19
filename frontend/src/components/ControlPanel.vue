<template>
  <div class="control-panel">
    <div class="control-header">
      <div class="control-status" :class="{ active: controlOutput?.enabled }">
        <span class="status-dot"></span>
        <span>{{ controlOutput?.enabled ? '自动控制运行中' : '手动模式' }}</span>
      </div>
      <button class="ctrl-btn" :class="{ active: controlOutput?.enabled }" @click="toggleControl">
        {{ controlOutput?.enabled ? '停用' : '启用' }}
      </button>
    </div>

    <div class="control-body">
      <div class="gauge-row">
        <div class="gauge-item">
          <div class="gauge-ring">
            <svg viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
              <circle cx="50" cy="50" r="42" fill="none" :stroke="fuelColor" stroke-width="8"
                :stroke-dasharray="fuelArc + ' ' + (264 - fuelArc)"
                transform="rotate(-90 50 50)" stroke-linecap="round"/>
            </svg>
            <span class="gauge-val">{{ (controlOutput?.fuel_valve_pos * 100 || 0).toFixed(1) }}%</span>
          </div>
          <label>燃料阀位</label>
        </div>
        <div class="gauge-item">
          <div class="gauge-ring">
            <svg viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
              <circle cx="50" cy="50" r="42" fill="none" :stroke="airColor" stroke-width="8"
                :stroke-dasharray="airArc + ' ' + (264 - airArc)"
                transform="rotate(-90 50 50)" stroke-linecap="round"/>
            </svg>
            <span class="gauge-val">{{ (controlOutput?.air_damper_pos * 100 || 0).toFixed(1) }}%</span>
          </div>
          <label>风门开度</label>
        </div>
      </div>

      <div class="target-params">
        <div class="param-row">
          <span class="param-label">目标温度</span>
          <span class="param-value">{{ (controlOutput?.target_temp || 0).toFixed(0) }} K</span>
        </div>
        <div class="param-row">
          <span class="param-label">目标燃空比</span>
          <span class="param-value">{{ (controlOutput?.target_fuel_air || 0).toFixed(4) }}</span>
        </div>
        <div class="param-row">
          <span class="param-label">目标压力</span>
          <span class="param-value">{{ (controlOutput?.target_pressure || 0).toFixed(2) }} MPa</span>
        </div>
      </div>

      <div class="mode-selector">
        <label>控制模式</label>
        <div class="mode-buttons">
          <button v-for="m in modes" :key="m.value"
            class="mode-btn" :class="{ active: controlOutput?.mode === m.value }"
            @click="$emit('setMode', m.value)">
            {{ m.label }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'ControlPanel',
  props: {
    controlOutput: Object,
  },
  emits: ['toggle', 'setMode'],
  setup(props) {
    const fuelArc = computed(() => (props.controlOutput?.fuel_valve_pos || 0) * 264)
    const airArc = computed(() => (props.controlOutput?.air_damper_pos || 0) * 264)
    const fuelColor = computed(() => {
      const v = props.controlOutput?.fuel_valve_pos || 0
      if (v > 0.85) return '#ff3344'
      if (v > 0.7) return '#ffaa00'
      return '#00d4ff'
    })
    const airColor = computed(() => {
      const v = props.controlOutput?.air_damper_pos || 0
      if (v > 0.85) return '#ff3344'
      if (v > 0.7) return '#ffaa00'
      return '#00ff88'
    })
    const modes = [
      { value: 'manual', label: '手动' },
      { value: 'auto', label: '自动' },
    ]
    const toggleControl = () => {
      props.controlOutput?.enabled ? null : null
    }
    return { fuelArc, airArc, fuelColor, airColor, modes, toggleControl }
  },
}
</script>

<style scoped>
.control-panel { height: 100%; display: flex; flex-direction: column; }
.control-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.control-status { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-secondary); }
.control-status .status-dot { width: 8px; height: 8px; border-radius: 50%; background: #556677; }
.control-status.active .status-dot { background: var(--accent-green); box-shadow: 0 0 8px var(--accent-green); }
.ctrl-btn { padding: 4px 16px; border: 1px solid var(--border-color); background: transparent; color: var(--text-secondary); border-radius: 4px; cursor: pointer; font-size: 12px; transition: all 0.3s; }
.ctrl-btn.active { border-color: var(--accent-green); color: var(--accent-green); background: rgba(0, 255, 136, 0.1); }
.ctrl-btn:hover { border-color: var(--accent-cyan); }
.gauge-row { display: flex; gap: 16px; margin-bottom: 16px; }
.gauge-item { flex: 1; text-align: center; }
.gauge-ring { position: relative; width: 80px; height: 80px; margin: 0 auto 4px; }
.gauge-ring svg { width: 100%; height: 100%; }
.gauge-val { position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); font-size: 14px; font-weight: 600; color: var(--text-primary); }
.gauge-item label { font-size: 11px; color: var(--text-secondary); }
.target-params { background: rgba(0,0,0,0.2); border-radius: 6px; padding: 8px 12px; margin-bottom: 12px; }
.param-row { display: flex; justify-content: space-between; padding: 4px 0; font-size: 12px; }
.param-label { color: var(--text-secondary); }
.param-value { color: var(--accent-cyan); font-weight: 500; }
.mode-selector label { display: block; font-size: 11px; color: var(--text-secondary); margin-bottom: 6px; }
.mode-buttons { display: flex; gap: 6px; }
.mode-btn { flex: 1; padding: 6px 8px; border: 1px solid var(--border-color); background: transparent; color: var(--text-secondary); border-radius: 4px; cursor: pointer; font-size: 11px; transition: all 0.3s; }
.mode-btn.active { border-color: var(--accent-cyan); color: var(--accent-cyan); background: rgba(0, 212, 255, 0.1); }
.mode-btn:hover { border-color: var(--accent-cyan); }
</style>
