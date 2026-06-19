package control

import (
	"math"
	"sync"
	"time"

	"gas-turbine-combustion-ai/models"
)

type PIDController struct {
	Kp        float64
	Ki        float64
	Kd        float64
	Setpoint  float64
	Output    float64
	prevError float64
	integral  float64
}

func NewPIDController(kp, ki, kd, setpoint float64) *PIDController {
	return &PIDController{Kp: kp, Ki: ki, Kd: kd, Setpoint: setpoint}
}

func (pid *PIDController) Update(processVar float64, dt float64) float64 {
	err := pid.Setpoint - processVar
	pid.integral += err * dt
	if pid.integral > 1.0 {
		pid.integral = 1.0
	}
	if pid.integral < -1.0 {
		pid.integral = -1.0
	}
	derivative := 0.0
	if dt > 0 {
		derivative = (err - pid.prevError) / dt
	}
	pid.Output = pid.Kp*err + pid.Ki*pid.integral + pid.Kd*derivative
	pid.prevError = err
	if pid.Output > 1.0 {
		pid.Output = 1.0
	}
	if pid.Output < -1.0 {
		pid.Output = -1.0
	}
	return pid.Output
}

type CombustionController struct {
	mu             sync.RWMutex
	enabled        bool
	mode           string
	tempPID        *PIDController
	fuelAirPID     *PIDController
	pressurePID    *PIDController
	lastUpdate     time.Time
	controlOutput  *models.ControlOutput
	tempHistory    []float64
	pressureHistory []float64
}

func NewCombustionController() *CombustionController {
	return &CombustionController{
		enabled: false,
		mode:    "manual",
		tempPID:     NewPIDController(0.8, 0.15, 0.3, 1200.0),
		fuelAirPID:  NewPIDController(0.5, 0.08, 0.2, 0.035),
		pressurePID: NewPIDController(0.6, 0.1, 0.25, 1.5),
		controlOutput: &models.ControlOutput{
			FuelValvePos:    0.5,
			AirDamperPos:    0.5,
			TargetTemp:      1200.0,
			TargetFuelAir:   0.035,
			TargetPressure:  1.5,
			Mode:            "manual",
			Enabled:         false,
			Timestamp:       time.Now(),
		},
		tempHistory:     make([]float64, 0, 60),
		pressureHistory: make([]float64, 0, 60),
	}
}

func (cc *CombustionController) SetEnabled(enabled bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.enabled = enabled
	cc.controlOutput.Enabled = enabled
	if enabled {
		cc.mode = "auto"
		cc.controlOutput.Mode = "auto"
		cc.lastUpdate = time.Now()
	} else {
		cc.mode = "manual"
		cc.controlOutput.Mode = "manual"
	}
}

func (cc *CombustionController) SetMode(mode string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.mode = mode
	cc.controlOutput.Mode = mode
}

func (cc *CombustionController) SetTargets(targetTemp, targetFuelAir, targetPressure float64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.tempPID.Setpoint = targetTemp
	cc.fuelAirPID.Setpoint = targetFuelAir
	cc.pressurePID.Setpoint = targetPressure
	cc.controlOutput.TargetTemp = targetTemp
	cc.controlOutput.TargetFuelAir = targetFuelAir
	cc.controlOutput.TargetPressure = targetPressure
}

func (cc *CombustionController) Update(readings map[string]*models.SensorReading, field *models.TemperatureField, state *models.CombustionState) *models.ControlOutput {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if !cc.enabled {
		cc.controlOutput.Timestamp = time.Now()
		return cc.controlOutput
	}

	now := time.Now()
	if cc.lastUpdate.IsZero() {
		cc.lastUpdate = now
		cc.controlOutput.Timestamp = now
		return cc.controlOutput
	}
	dt := now.Sub(cc.lastUpdate).Seconds()
	if dt < 0.01 {
		cc.controlOutput.Timestamp = now
		return cc.controlOutput
	}
	cc.lastUpdate = now

	avgTemp := 0.0
	if field != nil {
		avgTemp = field.AvgTemp
		cc.tempHistory = append(cc.tempHistory, avgTemp)
		if len(cc.tempHistory) > 60 {
			cc.tempHistory = cc.tempHistory[len(cc.tempHistory)-60:]
		}
	}

	avgPressure := 0.0
	pressureCount := 0
	avgFlowRate := 0.0
	flowCount := 0
	for _, r := range readings {
		if r.Type == "pressure" {
			avgPressure += r.Value
			pressureCount++
		}
		if r.Type == "flow_rate" {
			avgFlowRate += r.Value
			flowCount++
		}
	}
	if pressureCount > 0 {
		avgPressure /= float64(pressureCount)
		cc.pressureHistory = append(cc.pressureHistory, avgPressure)
		if len(cc.pressureHistory) > 60 {
			cc.pressureHistory = cc.pressureHistory[len(cc.pressureHistory)-60:]
		}
	}
	if flowCount > 0 {
		avgFlowRate /= float64(flowCount)
	}

	currentFuelAir := 0.035
	if avgFlowRate > 0 {
		currentFuelAir = avgFlowRate * 0.033 / (avgFlowRate * 0.95)
	}

	tempCtrl := cc.tempPID.Update(avgTemp, dt)
	fuelAirCtrl := cc.fuelAirPID.Update(currentFuelAir, dt)
	pressureCtrl := cc.pressurePID.Update(avgPressure, dt)

	instabilityFactor := 0.0
	if state != nil {
		instabilityFactor = state.InstabilityIndex
	}

	stabilityDamping := 1.0 - 0.5*instabilityFactor
	if stabilityDamping < 0.3 {
		stabilityDamping = 0.3
	}

	fuelAdj := 0.5 + (tempCtrl*0.4+fuelAirCtrl*0.3+pressureCtrl*0.3)*stabilityDamping
	airAdj := 0.5 + (-tempCtrl*0.3+fuelAirCtrl*0.4-pressureCtrl*0.2)*stabilityDamping

	if instabilityFactor > 0.5 {
		fuelAdj = fuelAdj*0.7 + 0.5*0.3
		airAdj = airAdj*0.7 + 0.5*0.3
	}

	fuelAdj = math.Max(0.1, math.Min(1.0, fuelAdj))
	airAdj = math.Max(0.1, math.Min(1.0, airAdj))

	cc.controlOutput.FuelValvePos = fuelAdj*0.15 + cc.controlOutput.FuelValvePos*0.85
	cc.controlOutput.AirDamperPos = airAdj*0.15 + cc.controlOutput.AirDamperPos*0.85
	cc.controlOutput.Timestamp = now
	cc.controlOutput.Enabled = true
	cc.controlOutput.Mode = cc.mode

	return cc.controlOutput
}

func (cc *CombustionController) GetOutput() *models.ControlOutput {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	out := *cc.controlOutput
	return &out
}
