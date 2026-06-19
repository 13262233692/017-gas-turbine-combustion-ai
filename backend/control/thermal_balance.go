package control

import (
	"math"
	"sync"
	"time"

	"gas-turbine-combustion-ai/models"
)

type ThermalZone struct {
	ID         string
	MinTemp    float64
	MaxTemp    float64
	AvgTemp    float64
	HeatLoad   float64
	TargetLoad float64
	Weight     float64
}

type ThermalBalancer struct {
	mu       sync.RWMutex
	zones    []*ThermalZone
	result   *models.ThermalBalanceResult
	balanced bool
}

func NewThermalBalancer() *ThermalBalancer {
	zones := []*ThermalZone{
		{ID: "center", TargetLoad: 0.35, Weight: 1.2},
		{ID: "inner",  TargetLoad: 0.30, Weight: 1.0},
		{ID: "middle", TargetLoad: 0.20, Weight: 0.9},
		{ID: "outer",  TargetLoad: 0.15, Weight: 0.8},
	}
	return &ThermalBalancer{
		zones: zones,
		result: &models.ThermalBalanceResult{
			Zones:          make([]models.ZoneStatus, 4),
			BalanceIndex:   1.0,
			MaxImbalance:   0.0,
			Adjustments:    make(map[string]float64),
			Timestamp:      time.Now(),
		},
		balanced: true,
	}
}

func (tb *ThermalBalancer) Update(field *models.TemperatureField) *models.ThermalBalanceResult {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if field == nil || len(field.Grid) == 0 {
		tb.result.Timestamp = time.Now()
		return tb.result
	}

	rows := len(field.Grid)
	cols := len(field.Grid[0])
	centerR := rows / 2
	centerC := cols / 2

	zoneData := make(map[string][]float64)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			di := float64(i - centerR)
			dj := float64(j - centerC)
			dist := math.Sqrt(di*di+dj*dj) / math.Max(float64(centerR), float64(centerC))

			zoneID := ""
			switch {
			case dist < 0.25:
				zoneID = "center"
			case dist < 0.5:
				zoneID = "inner"
			case dist < 0.75:
				zoneID = "middle"
			default:
				zoneID = "outer"
			}
			zoneData[zoneID] = append(zoneData[zoneID], field.Grid[i][j])
		}
	}

	totalHeatLoad := 0.0
	for _, z := range tb.zones {
		data, ok := zoneData[z.ID]
		if !ok || len(data) == 0 {
			continue
		}
		minT, maxT, sumT := math.MaxFloat64, 0.0, 0.0
		for _, v := range data {
			if v < minT {
				minT = v
			}
			if v > maxT {
				maxT = v
			}
			sumT += v
		}
		z.MinTemp = minT
		z.MaxTemp = maxT
		z.AvgTemp = sumT / float64(len(data))
		z.HeatLoad = z.AvgTemp * z.Weight
		totalHeatLoad += z.HeatLoad
	}

	if totalHeatLoad > 0 {
		for _, z := range tb.zones {
			z.HeatLoad = z.HeatLoad / totalHeatLoad
		}
	}

	maxImbalance := 0.0
	balanceSum := 0.0
	adjustments := make(map[string]float64)
	tb.balanced = true

	for _, z := range tb.zones {
		deviation := z.HeatLoad - z.TargetLoad
		absDev := math.Abs(deviation)
		if absDev > maxImbalance {
			maxImbalance = absDev
		}
		balanceSum += 1.0 - absDev

		if absDev > 0.05 {
			tb.balanced = false
			adjustments[z.ID] = -deviation * 0.3
		} else {
			adjustments[z.ID] = 0.0
		}
	}

	balanceIndex := balanceSum / float64(len(tb.zones))
	if balanceIndex < 0 {
		balanceIndex = 0
	}

	tb.result.BalanceIndex = math.Round(balanceIndex*1000) / 1000
	tb.result.MaxImbalance = math.Round(maxImbalance*1000) / 1000
	tb.result.Adjustments = adjustments
	tb.result.Timestamp = time.Now()

	zoneStatuses := make([]models.ZoneStatus, len(tb.zones))
	for i, z := range tb.zones {
		zoneStatuses[i] = models.ZoneStatus{
			ID:         z.ID,
			MinTemp:    math.Round(z.MinTemp*10) / 10,
			MaxTemp:    math.Round(z.MaxTemp*10) / 10,
			AvgTemp:    math.Round(z.AvgTemp*10) / 10,
			HeatLoad:   math.Round(z.HeatLoad*1000) / 1000,
			TargetLoad: z.TargetLoad,
			Adjustment: math.Round(adjustments[z.ID]*1000) / 1000,
		}
	}
	tb.result.Zones = zoneStatuses

	return tb.result
}

func (tb *ThermalBalancer) GetResult() *models.ThermalBalanceResult {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	out := *tb.result
	return &out
}
