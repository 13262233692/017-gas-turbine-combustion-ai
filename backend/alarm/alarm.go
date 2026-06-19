package alarm

import (
	"fmt"
	"sync"
	"time"

	"gas-turbine-combustion-ai/config"
	"gas-turbine-combustion-ai/models"
)

type Manager struct {
	cfg    *config.Config
	mu     sync.RWMutex
	alarms []*models.Alarm
	nextID int64
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg:    cfg,
		alarms: make([]*models.Alarm, 0),
		nextID: 1,
	}
}

func (m *Manager) Check(readings map[string]*models.SensorReading, state *models.CombustionState, efficiency *models.ThermalEfficiency) []*models.Alarm {
	var newAlarms []*models.Alarm

	for _, r := range readings {
		if r.Type == "temperature" {
			if r.Value > m.cfg.Alarm.MaxTempThreshold {
				if !m.hasActiveAlarm("over_temperature", r.SensorID) {
					a := m.createAlarm("critical", "over_temperature",
						fmt.Sprintf("传感器 %s 温度 %.1f K 超过阈值 %.1f K", r.SensorID, r.Value, m.cfg.Alarm.MaxTempThreshold),
						r.SensorID, r.Value, m.cfg.Alarm.MaxTempThreshold)
					newAlarms = append(newAlarms, a)
				} else {
					m.updateAlarmValue("over_temperature", r.SensorID, r.Value)
				}
			}
			if r.Value < m.cfg.Alarm.MinTempThreshold {
				if !m.hasActiveAlarm("under_temperature", r.SensorID) {
					a := m.createAlarm("warning", "under_temperature",
						fmt.Sprintf("传感器 %s 温度 %.1f K 低于阈值 %.1f K", r.SensorID, r.Value, m.cfg.Alarm.MinTempThreshold),
						r.SensorID, r.Value, m.cfg.Alarm.MinTempThreshold)
					newAlarms = append(newAlarms, a)
				} else {
					m.updateAlarmValue("under_temperature", r.SensorID, r.Value)
				}
			}
		}
	}

	if state != nil && !state.Stable {
		if !m.hasActiveAlarm("combustion_instability", "SYSTEM") {
			a := m.createAlarm("critical", "combustion_instability",
				fmt.Sprintf("燃烧不稳定指数 %.3f 超过阈值 %.3f", state.InstabilityIndex, m.cfg.Alarm.InstabilityThreshold),
				"SYSTEM", state.InstabilityIndex, m.cfg.Alarm.InstabilityThreshold)
			newAlarms = append(newAlarms, a)
		} else {
			m.updateAlarmValue("combustion_instability", "SYSTEM", state.InstabilityIndex)
		}
	}

	if efficiency != nil && efficiency.ThermalEfficiency < m.cfg.Alarm.EfficiencyMin {
		if !m.hasActiveAlarm("low_efficiency", "SYSTEM") {
			a := m.createAlarm("warning", "low_efficiency",
				fmt.Sprintf("热效率 %.2f%% 低于阈值 %.2f%%", efficiency.ThermalEfficiency*100, m.cfg.Alarm.EfficiencyMin*100),
				"SYSTEM", efficiency.ThermalEfficiency, m.cfg.Alarm.EfficiencyMin)
			newAlarms = append(newAlarms, a)
		} else {
			m.updateAlarmValue("low_efficiency", "SYSTEM", efficiency.ThermalEfficiency)
		}
	}

	m.cleanupResolved(readings, state, efficiency)

	return newAlarms
}

func (m *Manager) hasActiveAlarm(alarmType, sensorID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.alarms {
		if !a.Acknowledged && a.Type == alarmType && a.SensorID == sensorID {
			return true
		}
	}
	return false
}

func (m *Manager) updateAlarmValue(alarmType, sensorID string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.alarms {
		if !a.Acknowledged && a.Type == alarmType && a.SensorID == sensorID {
			a.Value = value
			a.Timestamp = time.Now()
			return
		}
	}
}

func (m *Manager) cleanupResolved(readings map[string]*models.SensorReading, state *models.CombustionState, efficiency *models.ThermalEfficiency) {
	m.mu.Lock()
	defer m.mu.Unlock()

	remaining := make([]*models.Alarm, 0, len(m.alarms))
	for _, a := range m.alarms {
		if a.Acknowledged {
			remaining = append(remaining, a)
			continue
		}

		resolved := false
		switch a.Type {
		case "over_temperature":
			if r, ok := readings[a.SensorID]; ok {
				if r.Value <= m.cfg.Alarm.MaxTempThreshold*0.98 {
					resolved = true
				}
			}
		case "under_temperature":
			if r, ok := readings[a.SensorID]; ok {
				if r.Value >= m.cfg.Alarm.MinTempThreshold*1.02 {
					resolved = true
				}
			}
		case "combustion_instability":
			if state != nil && state.Stable {
				resolved = true
			}
		case "low_efficiency":
			if efficiency != nil && efficiency.ThermalEfficiency >= m.cfg.Alarm.EfficiencyMin*1.02 {
				resolved = true
			}
		}

		if !resolved {
			remaining = append(remaining, a)
		}
	}
	m.alarms = remaining
}

func (m *Manager) createAlarm(level, alarmType, message, sensorID string, value, threshold float64) *models.Alarm {
	m.mu.Lock()
	defer m.mu.Unlock()

	alarm := &models.Alarm{
		ID:           m.nextID,
		Level:        level,
		Type:         alarmType,
		Message:      message,
		SensorID:     sensorID,
		Value:        value,
		Threshold:    threshold,
		Timestamp:    time.Now(),
		Acknowledged: false,
	}
	m.nextID++

	m.alarms = append(m.alarms, alarm)
	if len(m.alarms) > 500 {
		m.alarms = m.alarms[len(m.alarms)-500:]
	}

	return alarm
}

func (m *Manager) GetActive() []*models.Alarm {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var active []*models.Alarm
	for _, a := range m.alarms {
		if !a.Acknowledged {
			active = append(active, a)
		}
	}
	return active
}

func (m *Manager) GetAll(limit int) []*models.Alarm {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > len(m.alarms) {
		limit = len(m.alarms)
	}
	start := len(m.alarms) - limit
	if start < 0 {
		start = 0
	}
	result := make([]*models.Alarm, limit)
	copy(result, m.alarms[start:])
	return result
}

func (m *Manager) Acknowledge(id int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.alarms {
		if a.ID == id {
			a.Acknowledged = true
			return true
		}
	}
	return false
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.alarms)
}

func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, a := range m.alarms {
		if !a.Acknowledged {
			count++
		}
	}
	return count
}
