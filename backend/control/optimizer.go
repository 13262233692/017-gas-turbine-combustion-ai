package control

import (
	"math"
	"sync"
	"time"

	"gas-turbine-combustion-ai/models"
)

const (
	ModeStartup    = "startup"
	ModeNormal     = "normal"
	ModePeakLoad   = "peak_load"
	ModeLowLoad    = "low_load"
	ModeShutdown   = "shutdown"
	ModeEmergency  = "emergency"
)

type OperatingPoint struct {
	TargetTemp       float64
	TargetFuelAir    float64
	TargetPressure   float64
	MinEfficiency    float64
	MaxEmissionNOx   float64
	PowerOutput      float64
	LoadFraction     float64
}

var operatingPoints = map[string]OperatingPoint{
	ModeStartup:  {TargetTemp: 900, TargetFuelAir: 0.028, TargetPressure: 0.8, MinEfficiency: 0.60, MaxEmissionNOx: 50, PowerOutput: 20, LoadFraction: 0.2},
	ModeNormal:   {TargetTemp: 1200, TargetFuelAir: 0.035, TargetPressure: 1.5, MinEfficiency: 0.85, MaxEmissionNOx: 25, PowerOutput: 100, LoadFraction: 0.7},
	ModePeakLoad: {TargetTemp: 1450, TargetFuelAir: 0.040, TargetPressure: 2.0, MinEfficiency: 0.82, MaxEmissionNOx: 35, PowerOutput: 140, LoadFraction: 1.0},
	ModeLowLoad:  {TargetTemp: 1000, TargetFuelAir: 0.030, TargetPressure: 1.0, MinEfficiency: 0.75, MaxEmissionNOx: 20, PowerOutput: 40, LoadFraction: 0.3},
	ModeShutdown: {TargetTemp: 700, TargetFuelAir: 0.020, TargetPressure: 0.5, MinEfficiency: 0.50, MaxEmissionNOx: 15, PowerOutput: 0, LoadFraction: 0.0},
	ModeEmergency:{TargetTemp: 800, TargetFuelAir: 0.025, TargetPressure: 0.6, MinEfficiency: 0.40, MaxEmissionNOx: 40, PowerOutput: 0, LoadFraction: 0.0},
}

type EfficiencyOptimizer struct {
	mu             sync.RWMutex
	currentMode    string
	targetMode     string
	transitionProgress float64
	opPoint        OperatingPoint
	optimizationResult *models.OptimizationResult
	effHistory     []float64
	effWindow      int
}

func NewEfficiencyOptimizer() *EfficiencyOptimizer {
	return &EfficiencyOptimizer{
		currentMode:    ModeNormal,
		targetMode:     ModeNormal,
		transitionProgress: 1.0,
		opPoint:        operatingPoints[ModeNormal],
		effHistory:     make([]float64, 0, 120),
		effWindow:      30,
		optimizationResult: &models.OptimizationResult{
			CurrentMode:     ModeNormal,
			TargetMode:      ModeNormal,
			TransitionProgress: 1.0,
			OptimizedEfficiency: 0.85,
			EfficiencyGain:  0.0,
			PowerOutput:     100.0,
			LoadFraction:    0.7,
			Timestamp:       time.Now(),
		},
	}
}

func (eo *EfficiencyOptimizer) SetMode(mode string) {
	eo.mu.Lock()
	defer eo.mu.Unlock()
	if _, ok := operatingPoints[mode]; ok && mode != eo.currentMode {
		eo.targetMode = mode
		eo.transitionProgress = 0.0
	}
}

func (eo *EfficiencyOptimizer) GetMode() string {
	eo.mu.RLock()
	defer eo.mu.RUnlock()
	return eo.currentMode
}

func (eo *EfficiencyOptimizer) Update(efficiency *models.ThermalEfficiency, state *models.CombustionState, field *models.TemperatureField) *models.OptimizationResult {
	eo.mu.Lock()
	defer eo.mu.Unlock()

	if efficiency != nil {
		eo.effHistory = append(eo.effHistory, efficiency.ThermalEfficiency)
		if len(eo.effHistory) > 120 {
			eo.effHistory = eo.effHistory[len(eo.effHistory)-120:]
		}
	}

	if eo.transitionProgress < 1.0 {
		eo.transitionProgress += 0.02
		if eo.transitionProgress >= 1.0 {
			eo.transitionProgress = 1.0
			eo.currentMode = eo.targetMode
		}
	}

	src := operatingPoints[eo.currentMode]
	dst := operatingPoints[eo.targetMode]
	t := eo.transitionProgress
	smoothT := t * t * (3 - 2*t)

	eo.opPoint = OperatingPoint{
		TargetTemp:     src.TargetTemp + (dst.TargetTemp-src.TargetTemp)*smoothT,
		TargetFuelAir:  src.TargetFuelAir + (dst.TargetFuelAir-src.TargetFuelAir)*smoothT,
		TargetPressure: src.TargetPressure + (dst.TargetPressure-src.TargetPressure)*smoothT,
		MinEfficiency:  src.MinEfficiency + (dst.MinEfficiency-src.MinEfficiency)*smoothT,
		MaxEmissionNOx: src.MaxEmissionNOx + (dst.MaxEmissionNOx-src.MaxEmissionNOx)*smoothT,
		PowerOutput:    src.PowerOutput + (dst.PowerOutput-src.PowerOutput)*smoothT,
		LoadFraction:   src.LoadFraction + (dst.LoadFraction-src.LoadFraction)*smoothT,
	}

	currentEff := 0.0
	if efficiency != nil {
		currentEff = efficiency.ThermalEfficiency
	}

	effGain := 0.0
	if len(eo.effHistory) >= 10 {
		recentAvg := 0.0
		n := 0
		start := len(eo.effHistory) - 10
		for i := start; i < len(eo.effHistory); i++ {
			recentAvg += eo.effHistory[i]
			n++
		}
		recentAvg /= float64(n)

		predictedEff := recentAvg
		if state != nil && !state.Stable {
			predictedEff -= 0.03
		}
		if field != nil {
			tempSpread := field.MaxTemp - field.MinTemp
			if tempSpread > 400 {
				predictedEff -= 0.02
			}
			if tempSpread < 200 {
				predictedEff += 0.01
			}
		}
		predictedEff = math.Max(0.5, math.Min(0.98, predictedEff))

		optimizedEff := predictedEff
		if currentEff < eo.opPoint.MinEfficiency {
			optimizedEff = eo.opPoint.MinEfficiency + 0.02
			effGain = optimizedEff - currentEff
		} else {
			marginal := 0.005 * (1.0 - currentEff)
			optimizedEff = currentEff + marginal
			effGain = marginal
		}

		eo.optimizationResult.OptimizedEfficiency = math.Round(optimizedEff*1000) / 1000
		eo.optimizationResult.EfficiencyGain = math.Round(effGain*1000) / 1000
	}

	eo.optimizationResult.CurrentMode = eo.currentMode
	eo.optimizationResult.TargetMode = eo.targetMode
	eo.optimizationResult.TransitionProgress = math.Round(eo.transitionProgress*100) / 100
	eo.optimizationResult.PowerOutput = math.Round(eo.opPoint.PowerOutput*10) / 10
	eo.optimizationResult.LoadFraction = math.Round(eo.opPoint.LoadFraction*100) / 100
	eo.optimizationResult.Timestamp = time.Now()

	return eo.optimizationResult
}

func (eo *EfficiencyOptimizer) GetOperatingPoint() OperatingPoint {
	eo.mu.RLock()
	defer eo.mu.RUnlock()
	return eo.opPoint
}

func (eo *EfficiencyOptimizer) GetResult() *models.OptimizationResult {
	eo.mu.RLock()
	defer eo.mu.RUnlock()
	out := *eo.optimizationResult
	return &out
}
