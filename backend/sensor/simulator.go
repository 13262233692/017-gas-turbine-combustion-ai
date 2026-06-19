package sensor

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"gas-turbine-combustion-ai/config"
	"gas-turbine-combustion-ai/models"
)

type Simulator struct {
	cfg         *config.Config
	mu          sync.RWMutex
	sensors     map[string]*models.SensorReading
	history     map[string][]*models.SensorReading
	timeOffsets map[string]float64
	running     bool
	stopCh      chan struct{}
	historySize int
	startTime   time.Time
}

func NewSimulator(cfg *config.Config) *Simulator {
	return &Simulator{
		cfg:         cfg,
		sensors:     make(map[string]*models.SensorReading),
		history:     make(map[string][]*models.SensorReading),
		timeOffsets: make(map[string]float64),
		stopCh:      make(chan struct{}),
		historySize: 50,
		startTime:   time.Now(),
	}
}

func (s *Simulator) Start() {
	s.mu.Lock()
	s.running = true
	s.initTimeOffsets()
	s.mu.Unlock()

	interval := time.Duration(1000.0/s.cfg.Sensors.SampleRateHz) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.generateReadings()
		}
	}
}

func (s *Simulator) initTimeOffsets() {
	for i := 0; i < s.cfg.Sensors.TemperatureCount; i++ {
		id := s.sensorID("T", i)
		s.timeOffsets[id] = (rand.Float64() - 0.5) * 0.05
	}
	for i := 0; i < s.cfg.Sensors.PressureCount; i++ {
		id := s.sensorID("P", i)
		s.timeOffsets[id] = (rand.Float64() - 0.5) * 0.03
	}
	for i := 0; i < s.cfg.Sensors.FlowRateCount; i++ {
		id := s.sensorID("F", i)
		s.timeOffsets[id] = (rand.Float64() - 0.5) * 0.08
	}
}

func (s *Simulator) tempSensorRadius(index int) float64 {
	total := s.cfg.Sensors.TemperatureCount
	rings := 3
	perRing := total / rings
	ring := index / perRing
	if ring >= rings {
		ring = rings - 1
	}
	innerR := 0.2
	outerR := 0.5
	if rings <= 1 {
		return 0.35
	}
	return innerR + float64(ring)*((outerR-innerR)/float64(rings-1))
}

func (s *Simulator) tempSensorAngle(index int) float64 {
	total := s.cfg.Sensors.TemperatureCount
	rings := 3
	perRing := total / rings
	ring := index / perRing
	posInRing := index % perRing
	angleOffset := float64(ring) * 0.25
	return 2.0 * math.Pi * float64(posInRing) / float64(perRing) + angleOffset
}

func (s *Simulator) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		close(s.stopCh)
		s.running = false
	}
}

func (s *Simulator) GetReadings() map[string]*models.SensorReading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*models.SensorReading)
	for k, v := range s.sensors {
		result[k] = v
	}
	return result
}

func (s *Simulator) GetAlignedReadings(targetTime time.Time) map[string]*models.SensorReading {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*models.SensorReading)
	for id, hist := range s.history {
		if len(hist) < 2 {
			if len(hist) == 1 {
				result[id] = hist[0]
			}
			continue
		}

		idx := sort.Search(len(hist), func(i int) bool {
			return hist[i].Timestamp.After(targetTime) || hist[i].Timestamp.Equal(targetTime)
		})

		if idx == 0 {
			result[id] = hist[0]
		} else if idx >= len(hist) {
			result[id] = hist[len(hist)-1]
		} else {
			t0 := hist[idx-1]
			t1 := hist[idx]
			dt := t1.Timestamp.Sub(t0.Timestamp).Seconds()
			if dt == 0 {
				result[id] = t0
			} else {
				alpha := targetTime.Sub(t0.Timestamp).Seconds() / dt
				interpolated := &models.SensorReading{
					SensorID:  t0.SensorID,
					Type:      t0.Type,
					Value:     t0.Value + alpha*(t1.Value-t0.Value),
					Unit:      t0.Unit,
					Timestamp: targetTime,
					Quality:   math.Min(t0.Quality, t1.Quality) * (1.0 - 0.1*math.Abs(alpha-0.5)),
				}
				result[id] = interpolated
			}
		}
	}
	return result
}

func (s *Simulator) generateReadings() {
	now := time.Now()

	for i := 0; i < s.cfg.Sensors.TemperatureCount; i++ {
		id := s.sensorID("T", i)
		offset := s.getTimeOffset(id)
		sensorTime := now.Add(time.Duration(offset * float64(time.Second)))

		baseTemp := 1200.0 + 200*math.Sin(float64(sensorTime.UnixNano())/30.0/1e9*2*math.Pi)
		r := s.tempSensorRadius(i)
		angle := s.tempSensorAngle(i)
		radialFactor := 1.0 - 0.45*r*r
		circFactor := 1.0 + 0.12*math.Sin(2.0*angle+float64(sensorTime.UnixNano())/12.0/1e9*2*math.Pi)
		temp := baseTemp * radialFactor * circFactor
		temp += (rand.Float64() - 0.5) * 25

		quality := 0.93 + rand.Float64()*0.07
		if rand.Float64() < 0.01 {
			quality *= 0.7
		}

		reading := &models.SensorReading{
			SensorID:  id,
			Type:      "temperature",
			Value:     temp,
			Unit:      "K",
			Timestamp: sensorTime,
			Quality:   quality,
		}

		s.addReading(id, reading)
	}

	basePressure := 1.5 + 0.3*math.Sin(float64(now.UnixNano())/20.0/1e9*2*math.Pi)
	for i := 0; i < s.cfg.Sensors.PressureCount; i++ {
		id := s.sensorID("P", i)
		offset := s.getTimeOffset(id)
		sensorTime := now.Add(time.Duration(offset * float64(time.Second)))

		phase := float64(i) / float64(s.cfg.Sensors.PressureCount) * math.Pi * 2
		pressure := basePressure + 0.15*math.Sin(float64(sensorTime.UnixNano())/0.008/1e9*2*math.Pi+phase)
		pressure += (rand.Float64() - 0.5) * 0.1

		reading := &models.SensorReading{
			SensorID:  id,
			Type:      "pressure",
			Value:     pressure,
			Unit:      "MPa",
			Timestamp: sensorTime,
			Quality:   0.92 + rand.Float64()*0.08,
		}

		s.addReading(id, reading)
	}

	baseFlow := 2.5 + 0.5*math.Sin(float64(now.UnixNano())/15.0/1e9*2*math.Pi)
	for i := 0; i < s.cfg.Sensors.FlowRateCount; i++ {
		id := s.sensorID("F", i)
		offset := s.getTimeOffset(id)
		sensorTime := now.Add(time.Duration(offset * float64(time.Second)))

		flow := baseFlow + (rand.Float64()-0.5)*0.3

		reading := &models.SensorReading{
			SensorID:  id,
			Type:      "flow_rate",
			Value:     flow,
			Unit:      "kg/s",
			Timestamp: sensorTime,
			Quality:   0.88 + rand.Float64()*0.12,
		}

		s.addReading(id, reading)
	}
}

func (s *Simulator) getTimeOffset(id string) float64 {
	if offset, ok := s.timeOffsets[id]; ok {
		return offset
	}
	return 0
}

func (s *Simulator) addReading(id string, reading *models.SensorReading) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sensors[id] = reading

	if _, ok := s.history[id]; !ok {
		s.history[id] = make([]*models.SensorReading, 0, s.historySize)
	}
	s.history[id] = append(s.history[id], reading)
	if len(s.history[id]) > s.historySize {
		s.history[id] = s.history[id][len(s.history[id])-s.historySize:]
	}
}

func (s *Simulator) sensorID(prefix string, index int) string {
	return fmt.Sprintf("%s_%02d", prefix, index)
}
