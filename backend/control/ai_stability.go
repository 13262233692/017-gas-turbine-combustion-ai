package control

import (
	"math"
	"sync"
	"time"

	"gas-turbine-combustion-ai/models"
)

type StabilityController struct {
	mu              sync.RWMutex
	enabled         bool
	oscHistory      []float64
	freqHistory     []float64
	predictedIndex  float64
	dampingFactor   float64
	suppressionGain float64
	result          *models.AIStabilityResult
}

func NewStabilityController() *StabilityController {
	return &StabilityController{
		enabled:         true,
		oscHistory:      make([]float64, 0, 120),
		freqHistory:     make([]float64, 0, 120),
		predictedIndex:  0.0,
		dampingFactor:   0.0,
		suppressionGain: 0.0,
		result: &models.AIStabilityResult{
			PredictedInstability: 0.0,
			InstabilityTrend:     "stable",
			TimeToInstability:    0.0,
			DampingAction:        "none",
			DampingFactor:        0.0,
			SuppressionGain:      0.0,
			Confidence:           0.0,
			Enabled:              true,
			Timestamp:            time.Now(),
		},
	}
}

func (sc *StabilityController) SetEnabled(enabled bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.enabled = enabled
	sc.result.Enabled = enabled
}

func (sc *StabilityController) Update(state *models.CombustionState, field *models.TemperatureField, readings map[string]*models.SensorReading) *models.AIStabilityResult {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.enabled {
		sc.result.Timestamp = time.Now()
		return sc.result
	}

	currentIndex := 0.0
	currentFreq := 0.0
	if state != nil {
		currentIndex = state.InstabilityIndex
		currentFreq = state.PressureOscFreq
	}

	sc.oscHistory = append(sc.oscHistory, currentIndex)
	sc.freqHistory = append(sc.freqHistory, currentFreq)
	if len(sc.oscHistory) > 120 {
		sc.oscHistory = sc.oscHistory[len(sc.oscHistory)-120:]
	}
	if len(sc.freqHistory) > 120 {
		sc.freqHistory = sc.freqHistory[len(sc.freqHistory)-120:]
	}

	sc.predictedIndex = sc.predictInstability()

	trend := "stable"
	if len(sc.oscHistory) >= 10 {
		recent := sc.oscHistory[len(sc.oscHistory)-5:]
		older := sc.oscHistory[len(sc.oscHistory)-10 : len(sc.oscHistory)-5]

		recentAvg := avg(recent)
		olderAvg := avg(older)
		diff := recentAvg - olderAvg

		if diff > 0.02 {
			trend = "increasing"
		} else if diff < -0.02 {
			trend = "decreasing"
		} else {
			trend = "stable"
		}
	}

	timeToInstability := 0.0
	if trend == "increasing" && sc.predictedIndex < 0.35 {
		rate := 0.0
		if len(sc.oscHistory) >= 6 {
			rate = (sc.oscHistory[len(sc.oscHistory)-1] - sc.oscHistory[len(sc.oscHistory)-6]) / 6.0
		}
		if rate > 0.001 {
			timeToInstability = (0.35 - sc.predictedIndex) / rate
		} else {
			timeToInstability = 999.0
		}
	} else if sc.predictedIndex >= 0.35 {
		timeToInstability = 0.0
	} else {
		timeToInstability = -1.0
	}

	dampingAction := "none"
	sc.dampingFactor = 0.0
	sc.suppressionGain = 0.0

	if sc.predictedIndex > 0.25 {
		dampingAction = "minor_adjustment"
		sc.dampingFactor = 0.1
		sc.suppressionGain = 0.05
	}
	if sc.predictedIndex > 0.35 {
		dampingAction = "active_damping"
		sc.dampingFactor = 0.3
		sc.suppressionGain = 0.15
	}
	if sc.predictedIndex > 0.5 {
		dampingAction = "emergency_suppression"
		sc.dampingFactor = 0.6
		sc.suppressionGain = 0.35
	}

	if field != nil {
		tempSpread := field.MaxTemp - field.MinTemp
		if tempSpread > 500 {
			sc.suppressionGain += 0.05
		}
	}

	confidence := 0.5
	if len(sc.oscHistory) >= 30 {
		confidence = 0.8
	}
	if len(sc.oscHistory) >= 60 {
		confidence = 0.9
	}
	if len(sc.oscHistory) >= 90 {
		confidence = 0.95
	}

	sc.result.PredictedInstability = math.Round(sc.predictedIndex*1000) / 1000
	sc.result.InstabilityTrend = trend
	sc.result.TimeToInstability = math.Round(timeToInstability*10) / 10
	sc.result.DampingAction = dampingAction
	sc.result.DampingFactor = math.Round(sc.dampingFactor*1000) / 1000
	sc.result.SuppressionGain = math.Round(sc.suppressionGain*1000) / 1000
	sc.result.Confidence = math.Round(confidence*1000) / 1000
	sc.result.Enabled = sc.enabled
	sc.result.Timestamp = time.Now()

	return sc.result
}

func (sc *StabilityController) predictInstability() float64 {
	n := len(sc.oscHistory)
	if n < 3 {
		return sc.oscHistory[len(sc.oscHistory)-1]
	}

	weights := []float64{0.5, 0.3, 0.15, 0.05}
	wma := 0.0
	wSum := 0.0
	for i := 0; i < len(weights) && i < n; i++ {
		idx := n - 1 - i
		wma += sc.oscHistory[idx] * weights[i]
		wSum += weights[i]
	}
	if wSum > 0 {
		wma /= wSum
	}

	trend := 0.0
	if n >= 6 {
		half := n / 2
		firstHalf := avg(sc.oscHistory[:half])
		secondHalf := avg(sc.oscHistory[half:])
		trend = (secondHalf - firstHalf) * 2.0
	}

	predicted := wma + trend
	return math.Max(0, math.Min(1, predicted))
}

func (sc *StabilityController) GetDampingAdjustments() (float64, float64) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.dampingFactor, sc.suppressionGain
}

func (sc *StabilityController) GetResult() *models.AIStabilityResult {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	out := *sc.result
	return &out
}

func avg(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}
