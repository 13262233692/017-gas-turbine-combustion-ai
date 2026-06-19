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

type ControlOutput struct {
	FuelValvePos   float64   `json:"fuel_valve_pos"`
	AirDamperPos   float64   `json:"air_damper_pos"`
	TargetTemp     float64   `json:"target_temp"`
	TargetFuelAir  float64   `json:"target_fuel_air"`
	TargetPressure float64   `json:"target_pressure"`
	Mode           string    `json:"mode"`
	Enabled        bool      `json:"enabled"`
	Timestamp      time.Time `json:"timestamp"`
}

type OptimizationResult struct {
	CurrentMode         string    `json:"current_mode"`
	TargetMode          string    `json:"target_mode"`
	TransitionProgress  float64   `json:"transition_progress"`
	OptimizedEfficiency float64   `json:"optimized_efficiency"`
	EfficiencyGain      float64   `json:"efficiency_gain"`
	PowerOutput         float64   `json:"power_output"`
	LoadFraction        float64   `json:"load_fraction"`
	Timestamp           time.Time `json:"timestamp"`
}

type EmissionResult struct {
	NOxPPM            float64   `json:"nox_ppm"`
	COPPM             float64   `json:"co_ppm"`
	NOxLimit          float64   `json:"nox_limit"`
	COLimit           float64   `json:"co_limit"`
	NOxCompliance     bool      `json:"nox_compliance"`
	COCompliance      bool      `json:"co_compliance"`
	ConstraintActive  bool      `json:"constraint_active"`
	FuelAirAdjustment float64   `json:"fuel_air_adjustment"`
	TempAdjustment    float64   `json:"temp_adjustment"`
	Timestamp         time.Time `json:"timestamp"`
}

type ZoneStatus struct {
	ID         string  `json:"id"`
	MinTemp    float64 `json:"min_temp"`
	MaxTemp    float64 `json:"max_temp"`
	AvgTemp    float64 `json:"avg_temp"`
	HeatLoad   float64 `json:"heat_load"`
	TargetLoad float64 `json:"target_load"`
	Adjustment float64 `json:"adjustment"`
}

type ThermalBalanceResult struct {
	Zones        []ZoneStatus      `json:"zones"`
	BalanceIndex float64           `json:"balance_index"`
	MaxImbalance float64           `json:"max_imbalance"`
	Adjustments  map[string]float64 `json:"adjustments"`
	Timestamp    time.Time         `json:"timestamp"`
}

type AIStabilityResult struct {
	PredictedInstability float64   `json:"predicted_instability"`
	InstabilityTrend     string    `json:"instability_trend"`
	TimeToInstability    float64   `json:"time_to_instability"`
	DampingAction        string    `json:"damping_action"`
	DampingFactor        float64   `json:"damping_factor"`
	SuppressionGain      float64   `json:"suppression_gain"`
	Confidence           float64   `json:"confidence"`
	Enabled              bool      `json:"enabled"`
	Timestamp            time.Time `json:"timestamp"`
}
