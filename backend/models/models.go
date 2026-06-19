package models

import "time"

type SensorReading struct {
	ID        int64     `json:"id"`
	SensorID  string    `json:"sensor_id"`
	Type      string    `json:"type"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
	Quality   float64   `json:"quality"`
}

type TemperatureField struct {
	Grid     [][]float64 `json:"grid"`
	Rows     int         `json:"rows"`
	Cols     int         `json:"cols"`
	MaxTemp  float64     `json:"max_temp"`
	MinTemp  float64     `json:"min_temp"`
	AvgTemp  float64     `json:"avg_temp"`
	Timestamp time.Time  `json:"timestamp"`
}

type CombustionState struct {
	Stable          bool      `json:"stable"`
	InstabilityIndex float64  `json:"instability_index"`
	PressureOscFreq  float64  `json:"pressure_osc_freq"`
	PressureOscAmp   float64  `json:"pressure_osc_amp"`
	FlameIntensity   float64  `json:"flame_intensity"`
	Timestamp        time.Time `json:"timestamp"`
}

type ThermalEfficiency struct {
	CombustionEfficiency float64   `json:"combustion_efficiency"`
	ThermalEfficiency    float64   `json:"thermal_efficiency"`
	HeatReleaseRate      float64   `json:"heat_release_rate"`
	FuelAirRatio         float64   `json:"fuel_air_ratio"`
	ExhaustTemp          float64   `json:"exhaust_temp"`
	Timestamp            time.Time `json:"timestamp"`
}

type Alarm struct {
	ID        int64     `json:"id"`
	Level     string    `json:"level"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	SensorID  string    `json:"sensor_id"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
	Acknowledged bool   `json:"acknowledged"`
}

type SystemStatus struct {
	Online          bool      `json:"online"`
	SensorCount     int       `json:"sensor_count"`
	ActiveAlarms    int       `json:"active_alarms"`
	AIServiceOnline bool      `json:"ai_service_online"`
	Uptime          float64   `json:"uptime"`
	Timestamp       time.Time `json:"timestamp"`
}
