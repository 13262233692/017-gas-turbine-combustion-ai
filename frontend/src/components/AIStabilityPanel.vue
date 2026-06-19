<template>
  <div class="ai-stability-panel">
    <div class="ai-header">
      <div class="ai-status" :class="{ active: aiStability?.enabled }">
        <span class="ai-dot"></span>
        <span>{{ aiStability?.enabled ? 'AI 控制激活' : 'AI 控制关闭' }}</span>
      </div>
      <button class="ai-toggle" :class="{ active: aiStability?.enabled }" @click="toggleAI">
        {{ aiStability?.enabled ? '关闭' : '启用' }}
      </button>
    </div>

    <div class="prediction-display">
      <div class="pred-ring">
        <svg viewBox="0 0 100 100">
          <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
          <circle cx="50" cy="50" r="42" fill="none" :stroke="predColor" stroke-width="8"
            :stroke-dasharray="predArc + ' ' + (264 - predArc)"
            transform="rotate(-90 50 50)" stroke-linecap="round"/>
        </svg>
        <div class="pred-content">
          <span class="pred-val">{{ (aiStability?.predicted_instability * 100 || 0).toFixed(1) }}</span>
          <span class="pred-label">预测指数%</span>
        </div>
      </div>
    </div>

    <div class="trend-row">
      <span class="trend-label">趋势</span>
      <span class="trend-badge" :class="trendClass">
        {{ trendLabels[aiStability?.instability_trend] || '稳定' }}
      </span>
      <span class="confidence">置信度 {{ (aiStability?.confidence * 100 || 0).toFixed(0) }}%</span>
    </div>

    <div class="tti-row" v-if="aiStability?.time_to_instability > 0">
      <span class="tti-label">距不稳定</span>
      <span class="tti-val">{{ (aiStability?.time_to_instability || 0).toFixed(1) }}s</span>
    </div>

    <div class="damping-section">
      <div class="damping-row">
        <span class="damp-label">抑制策略</span>
        <span class="damp-val" :class="dampingClass">
          {{ dampingLabels[aiStability?.damping_action] || '无' }}
        </span>
      </div>
      <div class="damping-row">
        <span class="damp-label">阻尼系数</span>
        <div class="damp-bar">
          <div class="damp-fill" :style="{ width: (aiStability?.damping_factor || 0) * 100 + '%' }"></div>
        </div>
        <span class="damp-val-sm">{{ (aiStability?.damping_factor * 100 || 0).toFixed(1) }}%</span>
      </div>
      <div class="damping-row">
        <span class="damp-label">抑制增益</span>
        <div class="damp-bar">
          <div class="damp-fill suppress" :style="{ width: (aiStability?.suppression_gain || 0) * 100 + '%' }"></div>
        </div>
        <span class="damp-val-sm">{{ (aiStability?.suppression_gain * 100 || 0).toFixed(1) }}%</span>
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'AIStabilityPanel',
  props: {
    aiStability: Object,
  },
  emits: ['toggle'],
  setup(props) {
    const predArc = computed(() => (props.aiStability?.predicted_instability || 0) * 264)
    const predColor = computed(() => {
      const v = props.aiStability?.predicted_instability || 0
      if (v > 0.35) return '#ff3344'
      if (v > 0.25) return '#ffaa00'
      return '#00ff88'
    })
    const trendClass = computed(() => {
      const t = props.aiStability?.instability_trend
      if (t === 'increasing') return 'up'
      if (t === 'decreasing') return 'down'
      return 'stable'
    })
    const trendLabels = { stable: '稳定', increasing: '上升 ↑', decreasing: '下降 ↓' }
    const dampingLabels = { none: '无', minor_adjustment: '微调', active_damping: '主动阻尼', emergency_suppression: '紧急抑制' }
    const dampingClass = computed(() => {
      const a = props.aiStability?.damping_action
      if (a === 'emergency_suppression') return 'emergency'
      if (a === 'active_damping') return 'active'
      if (a === 'minor_adjustment') return 'minor'
      return ''
    })
    const toggleAI = () => {}
    return { predArc, predColor, trendClass, trendLabels, dampingLabels, dampingClass, toggleAI }
  },
}
</script>

<style scoped>
.ai-stability-panel { height: 100%; display: flex; flex-direction: column; }
.ai-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.ai-status { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-secondary); }
.ai-dot { width: 8px; height: 8px; border-radius: 50%; background: #556677; }
.ai-status.active .ai-dot { background: var(--accent-green); box-shadow: 0 0 8px var(--accent-green); }
.ai-toggle { padding: 4px 14px; border: 1px solid var(--border-color); background: transparent; color: var(--text-secondary); border-radius: 4px; cursor: pointer; font-size: 11px; transition: all 0.3s; }
.ai-toggle.active { border-color: var(--accent-green); color: var(--accent-green); background: rgba(0,255,136,0.1); }
.ai-toggle:hover { border-color: var(--accent-cyan); }
.prediction-display { text-align: center; margin-bottom: 10px; }
.pred-ring { position: relative; width: 100px; height: 100px; margin: 0 auto; }
.pred-ring svg { width: 100%; height: 100%; }
.pred-content { position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); text-align: center; }
.pred-val { display: block; font-size: 20px; font-weight: 700; color: var(--text-primary); }
.pred-label { font-size: 9px; color: var(--text-secondary); }
.trend-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; padding: 6px 10px; background: rgba(0,0,0,0.2); border-radius: 4px; }
.trend-label { font-size: 11px; color: var(--text-secondary); }
.trend-badge { font-size: 12px; font-weight: 600; padding: 1px 8px; border-radius: 3px; }
.trend-badge.stable { color: var(--accent-green); background: rgba(0,255,136,0.1); }
.trend-badge.up { color: var(--accent-red); background: rgba(255,51,68,0.1); }
.trend-badge.down { color: var(--accent-cyan); background: rgba(0,212,255,0.1); }
.confidence { margin-left: auto; font-size: 10px; color: var(--text-muted); }
.tti-row { display: flex; justify-content: space-between; padding: 4px 10px; margin-bottom: 8px; font-size: 11px; }
.tti-label { color: var(--text-secondary); }
.tti-val { color: var(--accent-yellow); font-weight: 600; }
.damping-section { background: rgba(0,0,0,0.2); border-radius: 6px; padding: 8px 10px; }
.damping-row { display: flex; align-items: center; gap: 6px; margin-bottom: 6px; font-size: 11px; }
.damping-row:last-child { margin-bottom: 0; }
.damp-label { color: var(--text-secondary); width: 56px; flex-shrink: 0; }
.damp-val { font-weight: 600; }
.damp-val.minor { color: var(--accent-yellow); }
.damp-val.active { color: var(--accent-orange); }
.damp-val.emergency { color: var(--accent-red); animation: pulse 1s infinite; }
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.6; } }
.damp-bar { flex: 1; height: 4px; background: #1a2744; border-radius: 2px; }
.damp-fill { height: 100%; border-radius: 2px; background: var(--accent-cyan); transition: width 0.5s; }
.damp-fill.suppress { background: var(--accent-orange); }
.damp-val-sm { font-size: 10px; color: var(--text-secondary); width: 36px; text-align: right; }
</style>
