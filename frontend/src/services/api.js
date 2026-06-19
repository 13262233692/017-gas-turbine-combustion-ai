import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

export const fetchSensors = () => api.get('/sensors')
export const fetchTemperatureField = () => api.get('/temperature-field')
export const fetchCombustionState = () => api.get('/combustion-state')
export const fetchEfficiency = () => api.get('/efficiency')
export const fetchAlarms = (active = false, limit = 50) =>
  api.get('/alarms', { params: { active, limit } })
export const acknowledgeAlarm = (id) => api.post(`/alarms/${id}/acknowledge`)
export const fetchSystemStatus = () => api.get('/system/status')

export const fetchControlOutput = () => api.get('/control/output')
export const enableControl = () => api.post('/control/enable')
export const disableControl = () => api.post('/control/disable')
export const setControlTargets = (targets) => api.post('/control/targets', targets)
export const setControlMode = (mode) => api.post('/control/mode', { mode })

export const fetchOptimization = () => api.get('/optimization')
export const setOperatingMode = (mode) => api.post('/optimization/mode', { mode })

export const fetchEmission = () => api.get('/emission')
export const setEmissionLimits = (limits) => api.post('/emission/limits', limits)

export const fetchThermalBalance = () => api.get('/thermal-balance')

export const fetchAIStability = () => api.get('/ai-stability')
export const enableAIStability = () => api.post('/ai-stability/enable')
export const disableAIStability = () => api.post('/ai-stability/disable')

export default api
