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

export default api
