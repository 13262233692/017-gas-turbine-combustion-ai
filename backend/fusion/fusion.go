package fusion

import (
	"math"
	"sort"
	"sync"
	"time"

	"gas-turbine-combustion-ai/config"
	"gas-turbine-combustion-ai/models"
)

const (
	PhysicalMaxTempK     = 2200.0
	PhysicalMinTempK     = 600.0
	MaxGradientKPerCell  = 80.0
	PDEDiffusivity       = 0.45
	PDEMaxIterations     = 20
	PDEConvergenceTol    = 0.05
	CFLSafetyFactor      = 0.4
	IDWPower             = 2.0
	OutlierSigmaThreshold = 3.0
)

type SensorInfo struct {
	id    string
	x, y  float64
	value float64
	weight float64
	quality float64
}

type KalmanState struct {
	x      float64
	P      float64
	Q      float64
	R      float64
}

type FusionEngine struct {
	cfg        *config.Config
	mu         sync.RWMutex
	readings   map[string]*models.SensorReading
	field      *models.TemperatureField
	prevField  *models.TemperatureField
	state      *models.CombustionState
	efficiency *models.ThermalEfficiency
	kalman     map[string]*KalmanState
	sensorPos  map[string][2]float64
	fieldGrid  [][]float64
	residual   float64
}

func NewFusionEngine(cfg *config.Config) *FusionEngine {
	fe := &FusionEngine{
		cfg:       cfg,
		readings:  make(map[string]*models.SensorReading),
		kalman:    make(map[string]*KalmanState),
		sensorPos: make(map[string][2]float64),
	}
	fe.initSensorPositions()
	return fe
}

func (f *FusionEngine) initSensorPositions() {
	totalTemp := f.cfg.Sensors.TemperatureCount
	rings := 3
	perRing := totalTemp / rings
	innerR := 0.2
	outerR := 0.5
	for ring := 0; ring < rings; ring++ {
		r := innerR + float64(ring)*((outerR-innerR)/float64(rings-1))
		angleOffset := float64(ring) * 0.25
		for i := 0; i < perRing; i++ {
			idx := ring*perRing + i
			if idx >= totalTemp {
				break
			}
			id := "T_" + pad(idx, 2)
			angle := 2.0*math.Pi*float64(i)/float64(perRing) + angleOffset
			f.sensorPos[id] = [2]float64{
				0.5 + r*math.Cos(angle),
				0.5 + r*math.Sin(angle),
			}
		}
	}
	totalPress := f.cfg.Sensors.PressureCount
	for i := 0; i < totalPress; i++ {
		id := "P_" + pad(i, 2)
		angle := 2.0 * math.Pi * float64(i) / float64(totalPress)
		r := 0.5
		f.sensorPos[id] = [2]float64{
			0.5 + r*math.Cos(angle),
			0.5 + r*math.Sin(angle),
		}
	}
	totalFlow := f.cfg.Sensors.FlowRateCount
	for i := 0; i < totalFlow; i++ {
		id := "F_" + pad(i, 2)
		angle := 2.0 * math.Pi * float64(i) / float64(totalFlow)
		r := 0.2
		f.sensorPos[id] = [2]float64{
			0.5 + r*math.Cos(angle),
			0.5 + r*math.Sin(angle),
		}
	}
}

func (f *FusionEngine) UpdateReadings(readings map[string]*models.SensorReading) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for id, reading := range readings {
		f.readings[id] = reading
		f.updateKalman(id, reading)
	}
}

func (f *FusionEngine) updateKalman(id string, reading *models.SensorReading) {
	if _, ok := f.kalman[id]; !ok {
		f.kalman[id] = &KalmanState{
			x: reading.Value,
			P: 100.0,
			Q: 5.0,
			R: 25.0 / (reading.Quality + 0.001),
		}
		return
	}

	k := f.kalman[id]
	k.R = 25.0 / (reading.Quality + 0.001)

	k.P += k.Q

	K := k.P / (k.P + k.R)
	k.x = k.x + K*(reading.Value - k.x)
	k.P = (1.0 - K) * k.P
}

func (f *FusionEngine) ReconstructTemperatureField() *models.TemperatureField {
	f.mu.Lock()
	defer f.mu.Unlock()

	rows := 16
	cols := 16

	var tempSensors []SensorInfo
	for _, r := range f.readings {
		if r.Type == "temperature" {
			pos, ok := f.sensorPos[r.SensorID]
			if !ok {
				pos = [2]float64{0.5, 0.5}
			}

			filteredVal := r.Value
			if ks, ok := f.kalman[r.SensorID]; ok {
				filteredVal = ks.x
			}

			tempSensors = append(tempSensors, SensorInfo{
				id:      r.SensorID,
				x:       pos[0],
				y:       pos[1],
				value:   filteredVal,
				quality: r.Quality,
				weight:  r.Quality,
			})
		}
	}

	if len(tempSensors) == 0 {
		return &models.TemperatureField{
			Grid:      f.fieldGrid,
			Rows:      rows,
			Cols:      cols,
			Timestamp: time.Now(),
		}
	}

	values := make([]float64, len(tempSensors))
	for i, s := range tempSensors {
		values[i] = s.value
	}
	mean, std := meanAndStd(values)
	filteredSensors := make([]SensorInfo, 0, len(tempSensors))
	for _, s := range tempSensors {
		if math.Abs(s.value - mean) <= OutlierSigmaThreshold * std {
			filteredSensors = append(filteredSensors, s)
		}
	}
	if len(filteredSensors) < 3 {
		filteredSensors = tempSensors
	}

	grid := f.idwInterpolate(filteredSensors, rows, cols)

	grid = f.solvePDE(grid, rows, cols)

	grid = f.applyPhysicalConstraints(grid, rows, cols)

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

	f.prevField = f.field
	f.fieldGrid = grid
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

func (f *FusionEngine) idwInterpolate(sensors []SensorInfo, rows, cols int) [][]float64 {
	grid := make([][]float64, rows)
	for i := range grid {
		grid[i] = make([]float64, cols)
	}

	centerTemp := 0.0
	totalWeight := 0.0
	for _, s := range sensors {
		dx := s.x - 0.5
		dy := s.y - 0.5
		r := math.Sqrt(dx*dx + dy*dy)
		w := s.quality * (1.0 - r)
		centerTemp += s.value * w
		totalWeight += w
	}
	if totalWeight > 0 {
		centerTemp /= totalWeight
	} else {
		centerTemp = 1300.0
	}

	edgeTemp := 0.0
	edgeWeight := 0.0
	for _, s := range sensors {
		dx := s.x - 0.5
		dy := s.y - 0.5
		r := math.Sqrt(dx*dx + dy*dy)
		if r > 0.5 {
			w := s.quality * (r - 0.5)
			edgeTemp += s.value * w
			edgeWeight += w
		}
	}
	if edgeWeight > 0 {
		edgeTemp /= edgeWeight
	} else {
		edgeTemp = centerTemp * 0.65
	}

	residuals := make([]SensorInfo, len(sensors))
	for idx, s := range sensors {
		dx := s.x - 0.5
		dy := s.y - 0.5
		r := math.Sqrt(dx*dx + dy*dy)
		baseTemp := centerTemp - (centerTemp - edgeTemp) * r * r
		residual := s.value - baseTemp
		residuals[idx] = SensorInfo{
			id:      s.id,
			x:       s.x,
			y:       s.y,
			value:   residual,
			quality: s.quality,
		}
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			gx := float64(j) / float64(cols - 1)
			gy := float64(i) / float64(rows - 1)

			dx := gx - 0.5
			dy := gy - 0.5
			r := math.Sqrt(dx*dx + dy*dy)
			baseTemp := centerTemp - (centerTemp - edgeTemp) * r * r

			residual := 0.0
			residualWeight := 0.0
			for _, s := range residuals {
				sdx := gx - s.x
				sdy := gy - s.y
				dist := math.Sqrt(sdx*sdx + sdy*sdy)

				if dist < 0.001 {
					residual = s.value
					residualWeight = 1.0
					break
				}

				w := s.quality / (dist * dist)
				residual += s.value * w
				residualWeight += w
			}

			if residualWeight > 0 {
				residual /= residualWeight
			}

			grid[i][j] = baseTemp + residual*0.7
		}
	}

	return grid
}

func (f *FusionEngine) solvePDE(grid [][]float64, rows, cols int) [][]float64 {
	dx := 1.0 / float64(cols - 1)
	dy := 1.0 / float64(rows - 1)

	dtMax := CFLSafetyFactor * dx * dx / (2.0 * PDEDiffusivity)
	dt := math.Min(dtMax, 0.05)

	result := make([][]float64, rows)
	temp := make([][]float64, rows)
	for i := range grid {
		result[i] = make([]float64, cols)
		temp[i] = make([]float64, cols)
		copy(result[i], grid[i])
		copy(temp[i], grid[i])
	}

	totalDiffusion := 0.0
	for iter := 0; iter < PDEMaxIterations; iter++ {
		maxChange := 0.0

		for i := 1; i < rows - 1; i++ {
			for j := 1; j < cols - 1; j++ {
				lap := (result[i-1][j] + result[i+1][j] +
					result[i][j-1] + result[i][j+1] -
					4.0 * result[i][j]) / (dx * dy)

				temp[i][j] = result[i][j] + PDEDiffusivity * dt * lap

				if i > 0 && i < rows - 1 && j > 0 && j < cols - 1 {
					change := math.Abs(temp[i][j] - result[i][j])
					if change > maxChange {
						maxChange = change
					}
				}
			}
		}

		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				result[i][j], temp[i][j] = temp[i][j], result[i][j]
			}
		}

		for i := 0; i < rows; i++ {
			result[i][0] = result[i][1]
			result[i][cols-1] = result[i][cols-2]
		}
		for j := 0; j < cols; j++ {
			result[0][j] = result[1][j]
			result[rows-1][j] = result[rows-2][j]
		}

		totalDiffusion += maxChange
		if maxChange < PDEConvergenceTol {
			break
		}
	}

	f.residual = totalDiffusion

	return result
}

func (f *FusionEngine) applyPhysicalConstraints(grid [][]float64, rows, cols int) [][]float64 {
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if grid[i][j] > PhysicalMaxTempK {
				grid[i][j] = PhysicalMaxTempK
			}
			if grid[i][j] < PhysicalMinTempK {
				grid[i][j] = PhysicalMinTempK
			}
		}
	}

	for iter := 0; iter < 3; iter++ {
		changed := false
		for i := 1; i < rows - 1; i++ {
			for j := 1; j < cols - 1; j++ {
				neighbors := []float64{
					grid[i-1][j], grid[i+1][j],
					grid[i][j-1], grid[i][j+1],
				}
				nMax, nMin := 0.0, math.MaxFloat64
				for _, n := range neighbors {
					if n > nMax { nMax = n }
					if n < nMin { nMin = n }
				}

				if grid[i][j] > nMax + MaxGradientKPerCell {
					grid[i][j] = nMax + MaxGradientKPerCell * 0.5
					changed = true
				}
				if grid[i][j] < nMin - MaxGradientKPerCell {
					grid[i][j] = nMin - MaxGradientKPerCell * 0.5
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}

	centerI, centerJ := rows/2, cols/2
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			di := float64(i - centerI) / float64(rows/2)
			dj := float64(j - centerJ) / float64(cols/2)
			r := math.Sqrt(di*di + dj*dj)
			if r > 1.0 {
				edgeTemp := PhysicalMinTempK + 100.0
				grid[i][j] = grid[i][j] * 0.3 + edgeTemp * 0.7
			}
		}
	}

	return grid
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
		values := make([]float64, 0, len(pressureSensors))
		for _, s := range pressureSensors {
			filteredVal := s.Value
			if ks, ok := f.kalman[s.SensorID]; ok {
				filteredVal = ks.x
			}
			values = append(values, filteredVal)
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
		instabilityIndex = math.Min(1.0, oscAmp / mean / 0.15)
		oscFreq = 120.0 + 80*instabilityIndex
	}

	var tempSensors []*models.SensorReading
	for _, r := range f.readings {
		if r.Type == "temperature" {
			tempSensors = append(tempSensors, r)
		}
	}
	if len(tempSensors) > 0 {
		avgTemp := f.kalmanWeightedAverage(tempSensors)
		flameIntensity = math.Min(1.0, math.Max(0, (avgTemp - 800.0) / 1000.0))
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
		avgTemp = f.kalmanWeightedAverage(tempSensors)
	}

	avgFlow := 2.5
	if len(flowSensors) > 0 {
		avgFlow = f.kalmanWeightedAverage(flowSensors)
	}

	normTemp := (avgTemp - 800.0) / 800.0
	combEff := 0.7 + 0.28 * math.Sqrt(math.Max(0, normTemp))
	combEff = math.Min(0.99, combEff)

	thermalEff := combEff * (0.85 + 0.1 * math.Sin(float64(time.Now().Unix()) / 60.0))
	thermalEff = math.Min(0.95, thermalEff)

	heatRelease := avgFlow * 43e6 * combEff
	fuelAir := 0.028 + 0.003 * math.Sin(float64(time.Now().Unix()) / 25.0)
	exhaustTemp := 500.0 + avgTemp * 0.45

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

func (f *FusionEngine) kalmanWeightedAverage(sensors []*models.SensorReading) float64 {
	totalWeight := 0.0
	weightedSum := 0.0
	for _, s := range sensors {
		var w float64
		if ks, ok := f.kalman[s.SensorID]; ok {
			w = s.Quality / (ks.P + 0.1)
		} else {
			w = s.Quality
		}
		val := s.Value
		if ks, ok := f.kalman[s.SensorID]; ok {
			val = ks.x
		}
		weightedSum += val * w
		totalWeight += w
	}
	if totalWeight == 0 {
		return 0
	}
	return weightedSum / totalWeight
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

func meanAndStd(values []float64) (float64, float64) {
	n := len(values)
	if n == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(n)

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(n)
	return mean, math.Sqrt(variance)
}

func pad(n, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		if n < 10 {
			s += "0"
		}
		n /= 10
	}
	return s + itostr(n)
}

func itostr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n % 10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
