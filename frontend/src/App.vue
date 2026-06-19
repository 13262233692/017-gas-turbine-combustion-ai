<template>
  <div class="app">
    <header class="app-header">
      <div class="header-left">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"/>
            <path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/>
          </svg>
        </div>
        <div class="header-title">
          <h1>燃气轮机燃烧室智能监测系统</h1>
          <span class="subtitle">Gas Turbine Combustion Chamber Intelligent Monitoring</span>
        </div>
      </div>
      <div class="header-right">
        <div class="status-indicators">
          <div class="status-item" :class="{ online: wsConnected }">
            <span class="status-dot"></span>
            <span>{{ wsConnected ? '实时连接' : '连接断开' }}</span>
          </div>
          <div class="status-item" :class="{ online: systemOnline }">
            <span class="status-dot"></span>
            <span>系统 {{ systemOnline ? '正常' : '离线' }}</span>
          </div>
          <div class="status-item" :class="{ warning: activeAlarmCount > 0 }">
            <span class="alarm-badge" v-if="activeAlarmCount > 0">{{ activeAlarmCount }}</span>
            <span>{{ activeAlarmCount }} 条报警</span>
          </div>
        </div>
        <div class="time-display">{{ currentTime }}</div>
      </div>
    </header>

    <main class="app-main">
      <div class="dashboard-grid">
        <section class="panel temperature-panel">
          <div class="panel-header">
            <h2>🌡️ 温度场实时重建</h2>
            <span class="panel-badge">CNN + PDE</span>
          </div>
          <div class="panel-body">
            <TemperatureHeatmap :fieldData="temperatureField" />
            <div class="temp-stats">
              <div class="stat-item">
                <label>最高温度</label>
                <span class="stat-value high">{{ formatTemp(tempStats.max) }} K</span>
              </div>
              <div class="stat-item">
                <label>最低温度</label>
                <span class="stat-value low">{{ formatTemp(tempStats.min) }} K</span>
              </div>
              <div class="stat-item">
                <label>平均温度</label>
                <span class="stat-value">{{ formatTemp(tempStats.avg) }} K</span>
              </div>
            </div>
          </div>
        </section>

        <section class="panel combustion-panel">
          <div class="panel-header">
            <h2>🔥 燃烧稳定性检测</h2>
            <span class="stability-badge" :class="stabilityClass">
              {{ combustionState?.stable ? '稳定' : '不稳定' }}
            </span>
          </div>
          <div class="panel-body">
            <CombustionGauge :state="combustionState" />
            <div class="combustion-details">
              <div class="detail-row">
                <span class="detail-label">不稳定指数</span>
                <div class="detail-bar">
                  <div class="bar-fill" :style="{ width: instabilityPercent + '%' }" :class="stabilityClass"></div>
                </div>
                <span class="detail-value">{{ (combustionState?.instability_index * 100).toFixed(1) }}%</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">压力振荡频率</span>
                <span class="detail-value">{{ combustionState?.pressure_osc_freq?.toFixed(1) || '0.0' }} Hz</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">压力振荡幅值</span>
                <span class="detail-value">{{ combustionState?.pressure_osc_amp?.toFixed(3) || '0.000' }} MPa</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">火焰强度</span>
                <div class="detail-bar">
                  <div class="bar-fill flame" :style="{ width: (combustionState?.flame_intensity || 0) * 100 + '%' }"></div>
                </div>
                <span class="detail-value">{{ ((combustionState?.flame_intensity || 0) * 100).toFixed(1) }}%</span>
              </div>
            </div>
          </div>
        </section>

        <section class="panel efficiency-panel">
          <div class="panel-header">
            <h2>⚡ 热效率分析</h2>
          </div>
          <div class="panel-body">
            <EfficiencyChart :efficiency="efficiency" />
            <div class="efficiency-grid">
              <div class="eff-item">
                <div class="eff-circle">
                  <svg viewBox="0 0 100 100">
                    <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
                    <circle cx="50" cy="50" r="42" fill="none" stroke="#00d4ff" stroke-width="8"
                      :stroke-dasharray="combustionEffArc + ' ' + (264 - combustionEffArc)"
                      transform="rotate(-90 50 50)" stroke-linecap="round"/>
                  </svg>
                  <span class="eff-percent">{{ ((efficiency?.combustion_efficiency || 0) * 100).toFixed(1) }}%</span>
                </div>
                <label>燃烧效率</label>
              </div>
              <div class="eff-item">
                <div class="eff-circle">
                  <svg viewBox="0 0 100 100">
                    <circle cx="50" cy="50" r="42" fill="none" stroke="#1a2744" stroke-width="8"/>
                    <circle cx="50" cy="50" r="42" fill="none" stroke="#00ff88" stroke-width="8"
                      :stroke-dasharray="thermalEffArc + ' ' + (264 - thermalEffArc)"
                      transform="rotate(-90 50 50)" stroke-linecap="round"/>
                  </svg>
                  <span class="eff-percent">{{ ((efficiency?.thermal_efficiency || 0) * 100).toFixed(1) }}%</span>
                </div>
                <label>热效率</label>
              </div>
            </div>
            <div class="eff-details">
              <div class="detail-row">
                <span class="detail-label">热释放率</span>
                <span class="detail-value">{{ (efficiency?.heat_release_rate || 0).toFixed(1) }} MW</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">燃空比</span>
                <span class="detail-value">{{ (efficiency?.fuel_air_ratio || 0).toFixed(4) }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">排气温度</span>
                <span class="detail-value">{{ (efficiency?.exhaust_temp || 0).toFixed(1) }} K</span>
              </div>
            </div>
          </div>
        </section>

        <section class="panel sensor-panel">
          <div class="panel-header">
            <h2>📡 多传感器融合</h2>
            <span class="sensor-count">{{ sensorCount }} 传感器在线</span>
          </div>
          <div class="panel-body">
            <SensorGrid :sensors="sensors" />
          </div>
        </section>

        <section class="panel alarm-panel">
          <div class="panel-header">
            <h2>🚨 异常报警系统</h2>
            <span class="alarm-count" :class="{ active: activeAlarmCount > 0 }">
              {{ activeAlarmCount }} 条活跃报警
            </span>
          </div>
          <div class="panel-body">
            <AlarmList :alarms="alarms" @acknowledge="handleAcknowledge" />
          </div>
        </section>

        <section class="panel trend-panel">
          <div class="panel-header">
            <h2>📈 温度趋势</h2>
          </div>
          <div class="panel-body">
            <TemperatureTrend :history="tempHistory" />
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { wsService } from './services/websocket.js'
import { fetchAlarms, acknowledgeAlarm, fetchSystemStatus } from './services/api.js'
import TemperatureHeatmap from './components/TemperatureHeatmap.vue'
import CombustionGauge from './components/CombustionGauge.vue'
import EfficiencyChart from './components/EfficiencyChart.vue'
import SensorGrid from './components/SensorGrid.vue'
import AlarmList from './components/AlarmList.vue'
import TemperatureTrend from './components/TemperatureTrend.vue'

export default {
  name: 'App',
  components: { TemperatureHeatmap, CombustionGauge, EfficiencyChart, SensorGrid, AlarmList, TemperatureTrend },
  setup() {
    const wsConnected = ref(false)
    const systemOnline = ref(false)
    const temperatureField = ref(null)
    const combustionState = ref(null)
    const efficiency = ref(null)
    const sensors = ref({})
    const alarms = ref([])
    const activeAlarmCount = ref(0)
    const currentTime = ref('')
    const tempHistory = ref([])
    const startTime = Date.now()

    let timeInterval = null
    let statusInterval = null

    const tempStats = computed(() => {
      if (!temperatureField.value) return { max: 0, min: 0, avg: 0 }
      return {
        max: temperatureField.value.max_temp,
        min: temperatureField.value.min_temp,
        avg: temperatureField.value.avg_temp,
      }
    })

    const instabilityPercent = computed(() => {
      return Math.min(100, (combustionState.value?.instability_index || 0) * 100)
    })

    const stabilityClass = computed(() => {
      if (!combustionState.value) return 'stable'
      return combustionState.value.stable ? 'stable' : 'unstable'
    })

    const sensorCount = computed(() => Object.keys(sensors.value).length)

    const combustionEffArc = computed(() => {
      const val = efficiency.value?.combustion_efficiency || 0
      return val * 264
    })

    const thermalEffArc = computed(() => {
      const val = efficiency.value?.thermal_efficiency || 0
      return val * 264
    })

    const formatTemp = (val) => (val || 0).toFixed(1)

    const updateTime = () => {
      const now = new Date()
      currentTime.value = now.toLocaleString('zh-CN', {
        year: 'numeric', month: '2-digit', day: '2-digit',
        hour: '2-digit', minute: '2-digit', second: '2-digit'
      })
    }

    const loadSystemStatus = async () => {
      try {
        const res = await fetchSystemStatus()
        systemOnline.value = res.data.online
        activeAlarmCount.value = res.data.active_alarms
      } catch (e) {
        systemOnline.value = false
      }
    }

    const loadAlarms = async () => {
      try {
        const res = await fetchAlarms(true, 20)
        alarms.value = res.data.alarms || []
        activeAlarmCount.value = res.data.count || 0
      } catch (e) {
        console.error('Failed to load alarms:', e)
      }
    }

    const handleAcknowledge = async (id) => {
      try {
        await acknowledgeAlarm(id)
        await loadAlarms()
      } catch (e) {
        console.error('Failed to acknowledge alarm:', e)
      }
    }

    const addTempHistoryPoint = () => {
      if (temperatureField.value) {
        tempHistory.value.push({
          time: new Date(),
          max: temperatureField.value.max_temp,
          min: temperatureField.value.min_temp,
          avg: temperatureField.value.avg_temp,
        })
        if (tempHistory.value.length > 120) {
          tempHistory.value = tempHistory.value.slice(-120)
        }
      }
    }

    onMounted(() => {
      wsService.on('connection', (data) => {
        wsConnected.value = data.status === 'connected'
      })

      wsService.on('temperature_field', (data) => {
        temperatureField.value = data
        addTempHistoryPoint()
      })

      wsService.on('combustion_state', (data) => {
        combustionState.value = data
      })

      wsService.on('efficiency', (data) => {
        efficiency.value = data
      })

      wsService.on('sensors', (data) => {
        sensors.value = data
      })

      wsService.on('alarm', (data) => {
        alarms.value.unshift(data)
        if (alarms.value.length > 50) alarms.value = alarms.value.slice(0, 50)
        activeAlarmCount.value++
      })

      wsService.connect()

      updateTime()
      timeInterval = setInterval(updateTime, 1000)
      loadSystemStatus()
      loadAlarms()
      statusInterval = setInterval(() => {
        loadSystemStatus()
        loadAlarms()
      }, 5000)
    })

    onUnmounted(() => {
      wsService.disconnect()
      if (timeInterval) clearInterval(timeInterval)
      if (statusInterval) clearInterval(statusInterval)
    })

    return {
      wsConnected, systemOnline, temperatureField, combustionState,
      efficiency, sensors, alarms, activeAlarmCount, currentTime,
      tempHistory, tempStats, instabilityPercent, stabilityClass,
      sensorCount, combustionEffArc, thermalEffArc, formatTemp,
      handleAcknowledge,
    }
  },
}
</script>

<style>
@import './styles/main.css';
</style>
