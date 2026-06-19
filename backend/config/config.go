package config

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Sensors  SensorsConfig  `json:"sensors"`
	Alarm    AlarmConfig    `json:"alarm"`
	AI       AIConfig       `json:"ai"`
}

type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
}

type SensorsConfig struct {
	TemperatureCount int     `json:"temperature_count"`
	PressureCount    int     `json:"pressure_count"`
	FlowRateCount    int     `json:"flow_rate_count"`
	SampleRateHz     float64 `json:"sample_rate_hz"`
}

type AlarmConfig struct {
	MaxTempThreshold     float64 `json:"max_temp_threshold"`
	MinTempThreshold     float64 `json:"min_temp_threshold"`
	InstabilityThreshold float64 `json:"instability_threshold"`
	EfficiencyMin        float64 `json:"efficiency_min"`
	CheckIntervalSec     int     `json:"check_interval_sec"`
}

type AIConfig struct {
	ModelPath      string  `json:"model_path"`
	PythonPath     string  `json:"python_path"`
	InferencePort  int     `json:"inference_port"`
	PredictionHz   float64 `json:"prediction_hz"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		Database: DatabaseConfig{
			Path: "./data/sensor.db",
		},
		Sensors: SensorsConfig{
			TemperatureCount: 24,
			PressureCount:    8,
			FlowRateCount:    4,
			SampleRateHz:     10.0,
		},
		Alarm: AlarmConfig{
			MaxTempThreshold:     1650.0,
			MinTempThreshold:     800.0,
			InstabilityThreshold: 0.35,
			EfficiencyMin:        0.85,
			CheckIntervalSec:     1,
		},
		AI: AIConfig{
			ModelPath:      "../ai/models/temperature_cnn_pde.onnx",
			PythonPath:     "python",
			InferencePort:  5000,
			PredictionHz:   2.0,
		},
	}
}
