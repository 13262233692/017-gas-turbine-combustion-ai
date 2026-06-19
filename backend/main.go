package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gas-turbine-combustion-ai/alarm"
	"gas-turbine-combustion-ai/config"
	"gas-turbine-combustion-ai/control"
	"gas-turbine-combustion-ai/fusion"
	"gas-turbine-combustion-ai/handler"
	"gas-turbine-combustion-ai/sensor"
	"gas-turbine-combustion-ai/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Default()

	hub := ws.NewHub()
	go hub.Run()

	simulator := sensor.NewSimulator(cfg)
	go simulator.Start()

	fusionEngine := fusion.NewFusionEngine(cfg)
	alarmManager := alarm.NewManager(cfg)

	combustionCtrl := control.NewCombustionController()
	effOptimizer := control.NewEfficiencyOptimizer()
	emissionModel := control.NewEmissionModel()
	thermalBalancer := control.NewThermalBalancer()
	aiStabilityCtrl := control.NewStabilityController()

	h := handler.NewHandler(
		simulator, fusionEngine, alarmManager, hub,
		combustionCtrl, effOptimizer, emissionModel, thermalBalancer, aiStabilityCtrl,
	)

	go dataPipeline(cfg, simulator, fusionEngine, alarmManager, hub,
		combustionCtrl, effOptimizer, emissionModel, thermalBalancer, aiStabilityCtrl)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(corsMiddleware())
	h.RegisterRoutes(r)
	r.GET("/ws", h.HandleWebSocket)
	r.Static("/assets", "./frontend/dist/assets")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Gas Turbine Combustion AI Server starting on %s", addr)
	log.Printf("   WebSocket endpoint: ws://%s/ws", addr)
	log.Printf("   API endpoint: http://%s/api", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func dataPipeline(
	cfg *config.Config,
	sim *sensor.Simulator,
	fusionEng *fusion.FusionEngine,
	alarmMgr *alarm.Manager,
	hub *ws.Hub,
	ctrl *control.CombustionController,
	optimizer *control.EfficiencyOptimizer,
	emission *control.EmissionModel,
	balancer *control.ThermalBalancer,
	aiStability *control.StabilityController,
) {
	interval := time.Duration(1000.0/cfg.AI.PredictionHz) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		alignedReadings := sim.GetAlignedReadings(now)
		fusionEng.UpdateReadings(alignedReadings)

		field := fusionEng.ReconstructTemperatureField()
		hub.BroadcastMessage("temperature_field", field)

		state := fusionEng.DetectInstability()
		hub.BroadcastMessage("combustion_state", state)

		efficiency := fusionEng.AnalyzeEfficiency()
		hub.BroadcastMessage("efficiency", efficiency)

		controlOutput := ctrl.Update(alignedReadings, field, state)
		hub.BroadcastMessage("control_output", controlOutput)

		optimResult := optimizer.Update(efficiency, state, field)
		hub.BroadcastMessage("optimization", optimResult)

		emissionResult := emission.Update(alignedReadings, field, efficiency)
		hub.BroadcastMessage("emission", emissionResult)

		balanceResult := balancer.Update(field)
		hub.BroadcastMessage("thermal_balance", balanceResult)

		aiStabilityResult := aiStability.Update(state, field, alignedReadings)
		hub.BroadcastMessage("ai_stability", aiStabilityResult)

		newAlarms := alarmMgr.Check(alignedReadings, state, efficiency)
		for _, a := range newAlarms {
			hub.BroadcastMessage("alarm", a)
		}

		hub.BroadcastMessage("sensors", alignedReadings)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
