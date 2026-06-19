<template>
  <div class="thermal-panel">
    <div class="balance-score">
      <div class="score-ring">
        <svg viewBox="0 0 100 100">
          <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
          <circle cx="50" cy="50" r="42" fill="none" :stroke="balanceColor" stroke-width="8"
            :stroke-dasharray="balanceArc + ' ' + (264 - balanceArc)"
            transform="rotate(-90 50 50)" stroke-linecap="round"/>
        </svg>
        <div class="score-content">
          <span class="score-val">{{ (balance?.balance_index * 100 || 0).toFixed(1) }}</span>
          <span class="score-unit">平衡指数</span>
        </div>
      </div>
      <div class="imbalance-info">
        <span class="info-label">最大偏差</span>
        <span class="info-val" :class="{ warning: (balance?.max_imbalance || 0) > 0.05 }">
          {{ ((balance?.max_imbalance || 0) * 100).toFixed(1) }}%
        </span>
      </div>
    </div>

    <div class="zone-list">
      <div v-for="zone in balance?.zones || []" :key="zone.id" class="zone-row">
        <div class="zone-header">
          <span class="zone-name">{{ zoneLabels[zone.id] || zone.id }}</span>
          <span class="zone-temp">{{ zone.avg_temp.toFixed(0) }} K</span>
        </div>
        <div class="zone-bars">
          <div class="bar-group">
            <span class="bar-label">实际</span>
            <div class="bar-track">
              <div class="bar-fill actual" :style="{ width: (zone.heat_load * 100) + '%' }"></div>
            </div>
            <span class="bar-value">{{ (zone.heat_load * 100).toFixed(1) }}%</span>
          </div>
          <div class="bar-group">
            <span class="bar-label">目标</span>
            <div class="bar-track">
              <div class="bar-fill target" :style="{ width: (zone.target_load * 100) + '%' }"></div>
            </div>
            <span class="bar-value target">{{ (zone.target_load * 100).toFixed(1) }}%</span>
          </div>
        </div>
        <div class="zone-adjustment" v-if="zone.adjustment !== 0">
          <span class="adj-label">调节</span>
          <span class="adj-val" :class="{ up: zone.adjustment > 0, down: zone.adjustment < 0 }">
            {{ zone.adjustment > 0 ? '+' : '' }}{{ (zone.adjustment * 100).toFixed(1) }}%
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'ThermalBalancePanel',
  props: {
    balance: Object,
  },
  setup(props) {
    const balanceArc = computed(() => (props.balance?.balance_index || 0) * 264)
    const balanceColor = computed(() => {
      const idx = props.balance?.balance_index || 0
      if (idx < 0.7) return '#ff3344'
      if (idx < 0.9) return '#ffaa00'
      return '#00ff88'
    })
    const zoneLabels = {
      center: '中心区', inner: '内环区', middle: '中环区', outer: '外环区',
    }
    return { balanceArc, balanceColor, zoneLabels }
  },
}
</script>

<style scoped>
.thermal-panel { height: 100%; display: flex; flex-direction: column; }
.balance-score { display: flex; align-items: center; gap: 16px; margin-bottom: 12px; }
.score-ring { position: relative; width: 80px; height: 80px; flex-shrink: 0; }
.score-ring svg { width: 100%; height: 100%; }
.score-content { position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); text-align: center; }
.score-val { display: block; font-size: 16px; font-weight: 700; color: var(--text-primary); }
.score-unit { font-size: 9px; color: var(--text-secondary); }
.imbalance-info { display: flex; flex-direction: column; gap: 4px; }
.info-label { font-size: 11px; color: var(--text-secondary); }
.info-val { font-size: 18px; font-weight: 600; color: var(--accent-green); }
.info-val.warning { color: var(--accent-yellow); }
.zone-list { display: flex; flex-direction: column; gap: 8px; }
.zone-row { background: rgba(0,0,0,0.2); border-radius: 6px; padding: 8px 10px; }
.zone-header { display: flex; justify-content: space-between; margin-bottom: 6px; }
.zone-name { font-size: 12px; font-weight: 600; color: var(--text-primary); }
.zone-temp { font-size: 12px; color: var(--accent-cyan); }
.zone-bars { display: flex; flex-direction: column; gap: 3px; }
.bar-group { display: flex; align-items: center; gap: 6px; }
.bar-label { font-size: 9px; color: var(--text-muted); width: 20px; }
.bar-track { flex: 1; height: 4px; background: #1a2744; border-radius: 2px; }
.bar-fill { height: 100%; border-radius: 2px; transition: width 0.5s; }
.bar-fill.actual { background: var(--accent-cyan); }
.bar-fill.target { background: var(--text-muted); opacity: 0.5; }
.bar-value { font-size: 10px; color: var(--text-secondary); width: 36px; text-align: right; }
.bar-value.target { color: var(--text-muted); }
.zone-adjustment { display: flex; justify-content: flex-end; gap: 6px; margin-top: 4px; font-size: 10px; }
.adj-label { color: var(--text-secondary); }
.adj-val { font-weight: 600; }
.adj-val.up { color: var(--accent-red); }
.adj-val.down { color: var(--accent-green); }
</style>
