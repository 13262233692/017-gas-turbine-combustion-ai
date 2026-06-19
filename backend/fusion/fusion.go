package fusion

import (
	"math"
	"sort"
	"sync"
	"time"

	"gas-turbine-combustion-ai/config"
	"gas-turbine-combustion-ai/models"
)

type FusionEngine struct {
	cfg       *config.Config
	mu        sync.RWMutex
	readings  map[string]*models.SensorReading
	field     *models.TemperatureField
	state     *models.CombustionState
	efficiency *models.ThermalEfficiency
}

func NewFusionEngine(cfg *config.Config) *FusionEngine {
	return &FusionEngine{
		cfg:      cfg,
		readings: make(map[string]*models.SensorReading),
	}
}

func (f *FusionEngine) UpdateReadings(readings map[string]*models.SensorReading) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for k, v := range readings {
		f.readings[k] = v
	}
}

func (f *FusionEngine) ReconstructTemperatureField() *models.TemperatureField {
	f.mu.Lock()
	defer f.mu.Unlock()

	rows := 16
	cols := 16
	grid := make([][]float64, rows)
	for i := range grid {
		grid[i] = make([]float64, cols)
	}

	var tempSensors []*models.SensorReading
	for _, r := range f.readings {
		if r.Type == "temperature" {
			tempSensors = append(tempSensors, r)
		}
	}

	if len(tempSensors) == 0 {
		return &models.TemperatureField{
			Grid:      grid,
			Rows:      rows,
			Cols:      cols,
			Timestamp: time.Now(),
		}
	}

	centerTemp := weightedAverage(tempSensors)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			ri := float64(i-rows/2) / float64(rows/2)
			rj := float64(j-cols/2) / float64(cols/2)
			radius := math.Sqrt(ri*ri + rj*rj)

			localTemp := centerTemp * (1.0 - 0.3*radius*radius)
			angle := math.Atan2(ri, rj)
			localTemp += 50 * math.Sin(3*angle) * (1 - radius)
			localTemp += (math.Sin(float64(time.Now().UnixNano())/1e9*2*math.Pi+float64(i+j)*0.3)) * 15

			grid[i][j] = localTemp
		}
	}

	maxT, minT, sumT := 0.0, math.MaxFloat64, 0.0
	for i := range grid {
		for j := range grid[i] {
			if grid[i][j] > maxT {
				maxT = grid[i][j]
			}
			if grid[i][j] < minT {
				minT = grid[i][j]
			}
			sumT += grid[i][j]
		}
	}

	f.field = &models.TemperatureField{
		Grid:      grid,
		Rows:      rows,
		Cols:      cols,
		MaxTemp:   maxT,
		MinTemp:   minT,
		AvgTemp:   sumT / float64(rows*cols),
		Timestamp: time.Now(),
	}

	return f.field
}

func (f *FusionEngine) DetectInstability() *models.CombustionState {
	f.mu.Lock()
	defer f.mu.Unlock()

	var pressureSensors []*models.SensorReading
	for _, r := range f.readings {
		if r.Type == "pressure" {
			pressureSensors = append(pressureSensors, r)
		}
	}

	instabilityIndex := 0.0
	oscFreq := 0.0
	oscAmp := 0.0
	flameIntensity := 0.8

	if len(pressureSensors) > 0 {
		var values []float64
		for _, s := range pressureSensors {
			values = append(values, s.Value)
		}
		sort.Float64s(values)
		mean := 0.0
		for _, v := range values {
			mean += v
		}
		mean /= float64(len(values))

		variance := 0.0
		for _, v := range values {
			variance += (v - mean) * (v - mean)
		}
		variance /= float64(len(values))

		oscAmp = math.Sqrt(variance)
		instabilityIndex = oscAmp / mean
		oscFreq = 120.0 + 30*instabilityIndex
	}

	var tempSensors []*models.SensorReading
	for _, r := range f.readings {
		if r.Type == "temperature" {
			tempSensors = append(tempSensors, r)
		}
	}
	if len(tempSensors) > 0 {
		avgTemp := weightedAverage(tempSensors)
		flameIntensity = math.Min(1.0, avgTemp/1600.0)
	}

	stable := instabilityIndex < f.cfg.Alarm.InstabilityThreshold

	f.state = &models.CombustionState{
		Stable:           stable,
		InstabilityIndex: instabilityIndex,
		PressureOscFreq:  oscFreq,
		PressureOscAmp:   oscAmp,
		FlameIntensity:   flameIntensity,
		Timestamp:        time.Now(),
	}

	return f.state
}

func (f *FusionEngine) AnalyzeEfficiency() *models.ThermalEfficiency {
	f.mu.Lock()
	defer f.mu.Unlock()

	var tempSensors []*models.SensorReading
	var flowSensors []*models.SensorReading
	for _, r := range f.readings {
		switch r.Type {
		case "temperature":
			tempSensors = append(tempSensors, r)
		case "flow_rate":
			flowSensors = append(flowSensors, r)
		}
	}

	avgTemp := 1200.0
	if len(tempSensors) > 0 {
		avgTemp = weightedAverage(tempSensors)
	}

	avgFlow := 2.5
	if len(flowSensors) > 0 {
		avgFlow = weightedAverage(flowSensors)
	}

	combEff := math.Min(0.99, 0.88+0.08*(avgTemp-1100)/400)
	if combEff < 0.7 {
		combEff = 0.7
	}

	thermalEff := combEff * 0.92
	heatRelease := avgFlow * 43e6 * combEff
	fuelAir := 0.028 + 0.005*math.Sin(float64(time.Now().Unix())/25.0)
	exhaustTemp := avgTemp * 0.65

	f.efficiency = &models.ThermalEfficiency{
		CombustionEfficiency: combEff,
		ThermalEfficiency:    thermalEff,
		HeatReleaseRate:      heatRelease / 1e6,
		FuelAirRatio:         fuelAir,
		ExhaustTemp:          exhaustTemp,
		Timestamp:            time.Now(),
	}

	return f.efficiency
}

func (f *FusionEngine) GetField() *models.TemperatureField {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.field
}

func (f *FusionEngine) GetState() *models.CombustionState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *FusionEngine) GetEfficiency() *models.ThermalEfficiency {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.efficiency
}

func weightedAverage(sensors []*models.SensorReading) float64 {
	totalWeight := 0.0
	weightedSum := 0.0
	for _, s := range sensors {
		w := s.Quality
		weightedSum += s.Value * w
		totalWeight += w
	}
	if totalWeight == 0 {
		return 0
	}
	return weightedSum / totalWeight
}
