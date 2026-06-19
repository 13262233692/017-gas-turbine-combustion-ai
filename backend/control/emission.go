package control

import (
	"math"
	"sync"
	"time"

	"gas-turbine-combustion-ai/models"
)

type EmissionModel struct {
	mu          sync.RWMutex
	nOxPPM      float64
	cOPPM       float64
	nOxLimit    float64
	cOLimit     float64
	constraintActive bool
	result      *models.EmissionResult
	nOxHistory  []float64
	cOHistory   []float64
	fuelAirAdj  float64
	tempAdj     float64
}

func NewEmissionModel() *EmissionModel {
	return &EmissionModel{
		nOxLimit:    25.0,
		cOLimit:     50.0,
		nOxHistory:  make([]float64, 0, 120),
		cOHistory:   make([]float64, 0, 120),
		fuelAirAdj:  0.0,
		tempAdj:     0.0,
		result: &models.EmissionResult{
			NOxPPM:           0,
			COPPM:            0,
			NOxLimit:         25.0,
			COLimit:          50.0,
			NOxCompliance:    true,
			COCompliance:     true,
			ConstraintActive: false,
			FuelAirAdjustment: 0,
			TempAdjustment:    0,
			Timestamp:        time.Now(),
		},
	}
}

func (em *EmissionModel) SetLimits(noxLimit, coLimit float64) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.nOxLimit = noxLimit
	em.cOLimit = coLimit
	em.result.NOxLimit = noxLimit
	em.result.COLimit = coLimit
}

func (em *EmissionModel) Update(readings map[string]*models.SensorReading, field *models.TemperatureField, efficiency *models.ThermalEfficiency) *models.EmissionResult {
	em.mu.Lock()
	defer em.mu.Unlock()

	avgTemp := 1200.0
	if field != nil {
		avgTemp = field.AvgTemp
		maxTemp := field.MaxTemp
		_ = maxTemp
	}

	fuelAirRatio := 0.035
	if efficiency != nil {
		fuelAirRatio = efficiency.FuelAirRatio
	}

	avgPressure := 1.5
	pressureCount := 0
	for _, r := range readings {
		if r.Type == "pressure" {
			avgPressure += r.Value
			pressureCount++
		}
	}
	if pressureCount > 0 {
		avgPressure /= float64(pressureCount + 1)
	}

	noxEmission := em.calculateNOx(avgTemp, fuelAirRatio, avgPressure)
	coEmission := em.calculateCO(avgTemp, fuelAirRatio)

	em.nOxPPM = noxEmission*0.3 + em.nOxPPM*0.7
	em.cOPPM = coEmission*0.3 + em.cOPPM*0.7

	em.nOxHistory = append(em.nOxHistory, em.nOxPPM)
	em.cOHistory = append(em.cOHistory, em.cOPPM)
	if len(em.nOxHistory) > 120 {
		em.nOxHistory = em.nOxHistory[len(em.nOxHistory)-120:]
	}
	if len(em.cOHistory) > 120 {
		em.cOHistory = em.cOHistory[len(em.cOHistory)-120:]
	}

	noxCompliance := em.nOxPPM <= em.nOxLimit
	coCompliance := em.cOPPM <= em.cOLimit
	em.constraintActive = !noxCompliance || !coCompliance

	em.fuelAirAdj = 0.0
	em.tempAdj = 0.0
	if em.constraintActive {
		if !noxCompliance {
			excessRatio := (em.nOxPPM - em.nOxLimit) / em.nOxLimit
			em.fuelAirAdj -= 0.005 * excessRatio
			em.tempAdj -= 10.0 * excessRatio
		}
		if !coCompliance {
			deficitRatio := (em.cOPPM - em.cOLimit) / em.cOLimit
			em.fuelAirAdj += 0.003 * deficitRatio
			em.tempAdj += 5.0 * deficitRatio
		}
		em.fuelAirAdj = math.Max(-0.01, math.Min(0.01, em.fuelAirAdj))
		em.tempAdj = math.Max(-30, math.Min(30, em.tempAdj))
	}

	em.result.NOxPPM = math.Round(em.nOxPPM*100) / 100
	em.result.COPPM = math.Round(em.cOPPM*100) / 100
	em.result.NOxCompliance = noxCompliance
	em.result.COCompliance = coCompliance
	em.result.ConstraintActive = em.constraintActive
	em.result.FuelAirAdjustment = math.Round(em.fuelAirAdj*10000) / 10000
	em.result.TempAdjustment = math.Round(em.tempAdj*10) / 10
	em.result.Timestamp = time.Now()

	return em.result
}

func (em *EmissionModel) calculateNOx(avgTemp, fuelAirRatio, pressure float64) float64 {
	activationTemp := 1800.0
	if avgTemp < activationTemp*0.8 {
		return 5.0 + 3.0*math.Exp(-0.002*(activationTemp-avgTemp))
	}
	thermalNOx := 50.0 * math.Exp(0.005*(avgTemp-activationTemp))
	promptNOx := 5.0 * (fuelAirRatio / 0.035)
	pressureFactor := math.Sqrt(pressure / 1.5)
	total := (thermalNOx + promptNOx) * pressureFactor
	return math.Max(0, total)
}

func (em *EmissionModel) calculateCO(avgTemp, fuelAirRatio float64) float64 {
	stoichRatio := 0.035
	leanExcess := (stoichRatio - fuelAirRatio) / stoichRatio

	if leanExcess > 0.3 {
		return 60.0 * (1.0 + leanExcess)
	}
	if avgTemp < 1100 {
		return 30.0 * math.Exp(-0.003*(avgTemp-1100))
	}
	if avgTemp > 1500 {
		return 5.0
	}
	baseCO := 15.0
	if fuelAirRatio < stoichRatio*0.9 {
		baseCO += 20.0 * (stoichRatio*0.9-fuelAirRatio) / stoichRatio
	}
	return baseCO
}

func (em *EmissionModel) GetAdjustments() (float64, float64) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.fuelAirAdj, em.tempAdj
}

func (em *EmissionModel) GetResult() *models.EmissionResult {
	em.mu.RLock()
	defer em.mu.RUnlock()
	out := *em.result
	return &out
}
