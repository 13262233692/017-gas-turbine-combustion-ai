<template>
  <div class="alarm-list">
    <div v-if="alarms.length === 0" class="no-alarms">
      <span class="no-alarm-icon">✅</span>
      <span>系统运行正常，暂无活跃报警</span>
    </div>
    <div v-else class="alarm-items">
      <div class="alarm-item" v-for="alarm in alarms" :key="alarm.id"
        :class="alarm.level">
        <div class="alarm-left">
          <span class="alarm-icon">{{ alarm.level === 'critical' ? '🔴' : '🟡' }}</span>
          <div class="alarm-info">
            <div class="alarm-message">{{ alarm.message }}</div>
            <div class="alarm-meta">
              <span class="alarm-type">{{ alarmTypeLabel(alarm.type) }}</span>
              <span class="alarm-time">{{ formatTime(alarm.timestamp) }}</span>
            </div>
          </div>
        </div>
        <div class="alarm-right">
          <button class="ack-btn" @click="$emit('acknowledge', alarm.id)"
            v-if="!alarm.acknowledged" :disabled="alarm.acknowledged">
            确认
          </button>
          <span v-else class="acked">已确认</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AlarmList',
  props: {
    alarms: { type: Array, default: () => [] },
  },
  emits: ['acknowledge'],
  setup() {
    const alarmTypeLabel = (type) => {
      const map = {
        over_temperature: '超温报警',
        under_temperature: '低温报警',
        combustion_instability: '燃烧不稳定',
        low_efficiency: '效率偏低',
      }
      return map[type] || type
    }

    const formatTime = (ts) => {
      if (!ts) return ''
      const d = new Date(ts)
      return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    }

    return { alarmTypeLabel, formatTime }
  },
}
</script>

<style scoped>
.alarm-list {
  max-height: 350px;
  overflow-y: auto;
}
.no-alarms {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 30px;
  color: #00ff88;
  font-size: 14px;
}
.no-alarm-icon {
  font-size: 20px;
}
.alarm-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.alarm-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 6px;
  border-left: 3px solid;
  background: #0d1b2a;
}
.alarm-item.critical {
  border-left-color: #ff3344;
  background: rgba(255, 51, 68, 0.05);
}
.alarm-item.warning {
  border-left-color: #ffaa00;
  background: rgba(255, 170, 0, 0.05);
}
.alarm-left {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}
.alarm-icon {
  font-size: 14px;
  margin-top: 2px;
}
.alarm-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.alarm-message {
  font-size: 13px;
  color: #dde4ec;
}
.alarm-meta {
  display: flex;
  gap: 12px;
  font-size: 11px;
}
.alarm-type {
  color: #00d4ff;
}
.alarm-time {
  color: #667788;
}
.ack-btn {
  padding: 4px 12px;
  border: 1px solid #00d4ff;
  background: transparent;
  color: #00d4ff;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}
.ack-btn:hover {
  background: #00d4ff;
  color: #0d1b2a;
}
.acked {
  color: #00ff88;
  font-size: 12px;
}
</style>
