package sensor

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"gas-turbine-combustion-ai/config"
	"gas-turbine-combustion-ai/models"
)

type Simulator struct {
	cfg     *config.Config
	mu      sync.RWMutex
	sensors map[string]*models.SensorReading
	running bool
	stopCh  chan struct{}
}

func NewSimulator(cfg *config.Config) *Simulator {
	return &Simulator{
		cfg:     cfg,
		sensors: make(map[string]*models.SensorReading),
		stopCh:  make(chan struct{}),
	}
}

func (s *Simulator) Start() {
	s.mu.Lock()
	s.running = true
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

func (s *Simulator) generateReadings() {
	now := time.Now()
	baseTemp := 1200.0 + 200*math.Sin(float64(now.Unix())/30.0)

	for i := 0; i < s.cfg.Sensors.TemperatureCount; i++ {
		id := s.sensorID("T", i)
		angle := float64(i) / float64(s.cfg.Sensors.TemperatureCount) * 2 * math.Pi
		radialFactor := 0.7 + 0.3*math.Sin(angle*2+float64(now.Unix())/10.0)
		temp := baseTemp * radialFactor
		temp += (rand.Float64() - 0.5) * 30
		s.mu.Lock()
		s.sensors[id] = &models.SensorReading{
			SensorID:  id,
			Type:      "temperature",
			Value:     temp,
			Unit:      "K",
			Timestamp: now,
			Quality:   0.95 + rand.Float64()*0.05,
		}
		s.mu.Unlock()
	}

	basePressure := 1.5 + 0.3*math.Sin(float64(now.Unix())/20.0)
	for i := 0; i < s.cfg.Sensors.PressureCount; i++ {
		id := s.sensorID("P", i)
		pressure := basePressure + (rand.Float64()-0.5)*0.2
		s.mu.Lock()
		s.sensors[id] = &models.SensorReading{
			SensorID:  id,
			Type:      "pressure",
			Value:     pressure,
			Unit:      "MPa",
			Timestamp: now,
			Quality:   0.93 + rand.Float64()*0.07,
		}
		s.mu.Unlock()
	}

	baseFlow := 2.5 + 0.5*math.Sin(float64(now.Unix())/15.0)
	for i := 0; i < s.cfg.Sensors.FlowRateCount; i++ {
		id := s.sensorID("F", i)
		flow := baseFlow + (rand.Float64()-0.5)*0.3
		s.mu.Lock()
		s.sensors[id] = &models.SensorReading{
			SensorID:  id,
			Type:      "flow_rate",
			Value:     flow,
			Unit:      "kg/s",
			Timestamp: now,
			Quality:   0.90 + rand.Float64()*0.10,
		}
		s.mu.Unlock()
	}
}

func (s *Simulator) sensorID(prefix string, index int) string {
	return prefix + "_" + padZero(index, 2)
}

func padZero(n, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		if n < 10 {
			s += "0"
		}
		n /= 10
	}
	return s
}
