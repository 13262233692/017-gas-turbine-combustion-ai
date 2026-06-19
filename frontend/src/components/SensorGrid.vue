<template>
  <div class="sensor-grid">
    <div class="sensor-category" v-for="category in categories" :key="category.type">
      <div class="category-header">
        <span class="category-icon">{{ category.icon }}</span>
        <span class="category-name">{{ category.name }}</span>
        <span class="category-count">{{ category.items.length }}</span>
      </div>
      <div class="sensor-list">
        <div class="sensor-card" v-for="s in category.items" :key="s.sensor_id"
          :class="{ warning: isWarning(s), critical: isCritical(s) }">
          <div class="sensor-id">{{ s.sensor_id }}</div>
          <div class="sensor-value">{{ s.value.toFixed(1) }}</div>
          <div class="sensor-unit">{{ s.unit }}</div>
          <div class="sensor-quality">
            <div class="quality-bar">
              <div class="quality-fill" :style="{ width: (s.quality * 100) + '%' }"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'SensorGrid',
  props: {
    sensors: { type: Object, default: () => ({}) },
  },
  setup(props) {
    const categories = computed(() => {
      const cats = {
        temperature: { type: 'temperature', icon: '🌡️', name: '温度传感器', items: [] },
        pressure: { type: 'pressure', icon: '📊', name: '压力传感器', items: [] },
        flow_rate: { type: 'flow_rate', icon: '💧', name: '流量传感器', items: [] },
      }

      Object.values(props.sensors).forEach(s => {
        if (cats[s.type]) {
          cats[s.type].items.push(s)
        }
      })

      return Object.values(cats).filter(c => c.items.length > 0)
    })

    const isWarning = (s) => {
      if (s.type === 'temperature') return s.value > 1500 || s.value < 900
      if (s.type === 'pressure') return s.value > 1.8 || s.value < 1.2
      return false
    }

    const isCritical = (s) => {
      if (s.type === 'temperature') return s.value > 1600 || s.value < 800
      if (s.type === 'pressure') return s.value > 2.0
      return false
    }

    return { categories, isWarning, isCritical }
  },
}
</script>

<style scoped>
.sensor-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}
.category-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-bottom: 1px solid #1a2744;
  font-size: 13px;
  color: #aabbcc;
}
.category-count {
  background: #1a2744;
  color: #00d4ff;
  border-radius: 10px;
  padding: 1px 8px;
  font-size: 11px;
}
.sensor-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 6px;
  padding: 4px;
}
.sensor-card {
  background: #0d1b2a;
  border: 1px solid #1a2744;
  border-radius: 6px;
  padding: 6px 8px;
  text-align: center;
  transition: all 0.3s;
}
.sensor-card:hover {
  border-color: #00d4ff;
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.2);
}
.sensor-card.warning {
  border-color: #ffaa00;
  background: rgba(255, 170, 0, 0.05);
}
.sensor-card.critical {
  border-color: #ff3344;
  background: rgba(255, 51, 68, 0.08);
  animation: pulse-critical 2s infinite;
}
@keyframes pulse-critical {
  0%, 100% { box-shadow: 0 0 5px rgba(255, 51, 68, 0.3); }
  50% { box-shadow: 0 0 15px rgba(255, 51, 68, 0.6); }
}
.sensor-id {
  font-size: 10px;
  color: #667788;
  margin-bottom: 2px;
}
.sensor-value {
  font-size: 16px;
  font-weight: bold;
  color: #e0e8f0;
}
.sensor-unit {
  font-size: 10px;
  color: #667788;
}
.quality-bar {
  height: 2px;
  background: #1a2744;
  border-radius: 1px;
  margin-top: 4px;
}
.quality-fill {
  height: 100%;
  background: #00ff88;
  border-radius: 1px;
  transition: width 0.3s;
}
</style>
