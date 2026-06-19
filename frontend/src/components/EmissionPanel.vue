<template>
  <div class="emission-panel">
    <div class="emission-gauges">
      <div class="emission-item">
        <div class="emission-ring">
          <svg viewBox="0 0 100 100">
            <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
            <circle cx="50" cy="50" r="42" fill="none" :stroke="noxColor" stroke-width="8"
              :stroke-dasharray="noxArc + ' ' + (264 - noxArc)"
              transform="rotate(-90 50 50)" stroke-linecap="round"/>
          </svg>
          <div class="ring-content">
            <span class="ring-val">{{ (emission?.nox_ppm || 0).toFixed(1) }}</span>
            <span class="ring-unit">ppm</span>
          </div>
        </div>
        <label>NOx 排放</label>
        <div class="compliance" :class="{ pass: emission?.nox_compliance, fail: !emission?.nox_compliance }">
          {{ emission?.nox_compliance ? '达标' : '超标' }}
        </div>
      </div>
      <div class="emission-item">
        <div class="emission-ring">
          <svg viewBox="0 0 100 100">
            <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
            <circle cx="50" cy="50" r="42" fill="none" :stroke="coColor" stroke-width="8"
              :stroke-dasharray="coArc + ' ' + (264 - coArc)"
              transform="rotate(-90 50 50)" stroke-linecap="round"/>
          </svg>
          <div class="ring-content">
            <span class="ring-val">{{ (emission?.co_ppm || 0).toFixed(1) }}</span>
            <span class="ring-unit">ppm</span>
          </div>
        </div>
        <label>CO 排放</label>
        <div class="compliance" :class="{ pass: emission?.co_compliance, fail: !emission?.co_compliance }">
          {{ emission?.co_compliance ? '达标' : '超标' }}
        </div>
      </div>
    </div>

    <div class="limits-row">
      <div class="limit-item">
        <span class="limit-label">NOx 限值</span>
        <span class="limit-val">{{ (emission?.nox_limit || 0).toFixed(0) }} ppm</span>
      </div>
      <div class="limit-item">
        <span class="limit-label">CO 限值</span>
        <span class="limit-val">{{ (emission?.co_limit || 0).toFixed(0) }} ppm</span>
      </div>
    </div>

    <div class="constraint-banner" v-if="emission?.constraint_active">
      <span class="constraint-icon">⚠️</span>
      <span>排放约束激活 - 自动调整中</span>
    </div>

    <div class="adjustments" v-if="emission?.constraint_active">
      <div class="adj-row">
        <span class="adj-label">燃空比修正</span>
        <span class="adj-val" :class="{ up: emission.fuel_air_adjustment > 0, down: emission.fuel_air_adjustment < 0 }">
          {{ emission.fuel_air_adjustment > 0 ? '+' : '' }}{{ (emission?.fuel_air_adjustment || 0).toFixed(4) }}
        </span>
      </div>
      <div class="adj-row">
        <span class="adj-label">温度修正</span>
        <span class="adj-val" :class="{ up: emission.temp_adjustment > 0, down: emission.temp_adjustment < 0 }">
          {{ emission.temp_adjustment > 0 ? '+' : '' }}{{ (emission?.temp_adjustment || 0).toFixed(1) }} K
        </span>
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'EmissionPanel',
  props: {
    emission: Object,
  },
  setup(props) {
    const noxArc = computed(() => {
      const ratio = (props.emission?.nox_ppm || 0) / (props.emission?.nox_limit || 25)
      return Math.min(1, ratio) * 264
    })
    const coArc = computed(() => {
      const ratio = (props.emission?.co_ppm || 0) / (props.emission?.co_limit || 50)
      return Math.min(1, ratio) * 264
    })
    const noxColor = computed(() => {
      const ratio = (props.emission?.nox_ppm || 0) / (props.emission?.nox_limit || 25)
      if (ratio > 1) return '#ff3344'
      if (ratio > 0.8) return '#ffaa00'
      return '#00ff88'
    })
    const coColor = computed(() => {
      const ratio = (props.emission?.co_ppm || 0) / (props.emission?.co_limit || 50)
      if (ratio > 1) return '#ff3344'
      if (ratio > 0.8) return '#ffaa00'
      return '#00d4ff'
    })
    return { noxArc, coArc, noxColor, coColor }
  },
}
</script>

<style scoped>
.emission-panel { height: 100%; display: flex; flex-direction: column; }
.emission-gauges { display: flex; gap: 12px; margin-bottom: 12px; }
.emission-item { flex: 1; text-align: center; }
.emission-ring { position: relative; width: 80px; height: 80px; margin: 0 auto 4px; }
.emission-ring svg { width: 100%; height: 100%; }
.ring-content { position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); text-align: center; }
.ring-val { display: block; font-size: 14px; font-weight: 700; color: var(--text-primary); }
.ring-unit { font-size: 9px; color: var(--text-secondary); }
.emission-item label { display: block; font-size: 11px; color: var(--text-secondary); margin-bottom: 4px; }
.compliance { display: inline-block; font-size: 10px; padding: 1px 8px; border-radius: 3px; font-weight: 600; }
.compliance.pass { color: var(--accent-green); background: rgba(0,255,136,0.1); }
.compliance.fail { color: var(--accent-red); background: rgba(255,51,68,0.1); animation: pulse 1s infinite; }
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.6; } }
.limits-row { display: flex; gap: 12px; margin-bottom: 10px; }
.limit-item { flex: 1; display: flex; justify-content: space-between; padding: 4px 8px; background: rgba(0,0,0,0.2); border-radius: 4px; font-size: 11px; }
.limit-label { color: var(--text-secondary); }
.limit-val { color: var(--text-primary); font-weight: 500; }
.constraint-banner { display: flex; align-items: center; gap: 6px; padding: 6px 10px; background: rgba(255,170,0,0.1); border: 1px solid rgba(255,170,0,0.3); border-radius: 4px; font-size: 11px; color: var(--accent-yellow); margin-bottom: 8px; }
.adjustments { background: rgba(0,0,0,0.2); border-radius: 4px; padding: 6px 10px; }
.adj-row { display: flex; justify-content: space-between; padding: 3px 0; font-size: 11px; }
.adj-label { color: var(--text-secondary); }
.adj-val { font-weight: 600; }
.adj-val.up { color: var(--accent-red); }
.adj-val.down { color: var(--accent-green); }
</style>
