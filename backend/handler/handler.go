package handler

import (
	"net/http"
	"strconv"
	"time"

	"gas-turbine-combustion-ai/alarm"
	"gas-turbine-combustion-ai/control"
	"gas-turbine-combustion-ai/fusion"
	"gas-turbine-combustion-ai/models"
	"gas-turbine-combustion-ai/sensor"
	"gas-turbine-combustion-ai/ws"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	simulator  *sensor.Simulator
	fusion     *fusion.FusionEngine
	alarmMgr   *alarm.Manager
	hub        *ws.Hub
	ctrl       *control.CombustionController
	optimizer  *control.EfficiencyOptimizer
	emission   *control.EmissionModel
	balancer   *control.ThermalBalancer
	aiStability *control.StabilityController
	startTime  time.Time
}

func NewHandler(
	sim *sensor.Simulator,
	fusionEng *fusion.FusionEngine,
	alarmMgr *alarm.Manager,
	hub *ws.Hub,
	ctrl *control.CombustionController,
	optimizer *control.EfficiencyOptimizer,
	emission *control.EmissionModel,
	balancer *control.ThermalBalancer,
	aiStability *control.StabilityController,
) *Handler {
	return &Handler{
		simulator:   sim,
		fusion:      fusionEng,
		alarmMgr:    alarmMgr,
		hub:         hub,
		ctrl:        ctrl,
		optimizer:   optimizer,
		emission:    emission,
		balancer:    balancer,
		aiStability: aiStability,
		startTime:   time.Now(),
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/sensors", h.GetSensors)
		api.GET("/temperature-field", h.GetTemperatureField)
		api.GET("/combustion-state", h.GetCombustionState)
		api.GET("/efficiency", h.GetEfficiency)
		api.GET("/alarms", h.GetAlarms)
		api.POST("/alarms/:id/acknowledge", h.AcknowledgeAlarm)
		api.GET("/system/status", h.GetSystemStatus)
		api.GET("/history/temperature", h.GetTemperatureHistory)
		api.GET("/history/efficiency", h.GetEfficiencyHistory)

		api.GET("/control/output", h.GetControlOutput)
		api.POST("/control/enable", h.EnableControl)
		api.POST("/control/disable", h.DisableControl)
		api.POST("/control/targets", h.SetControlTargets)
		api.POST("/control/mode", h.SetControlMode)

		api.GET("/optimization", h.GetOptimization)
		api.POST("/optimization/mode", h.SetOperatingMode)

		api.GET("/emission", h.GetEmission)
		api.POST("/emission/limits", h.SetEmissionLimits)

		api.GET("/thermal-balance", h.GetThermalBalance)

		api.GET("/ai-stability", h.GetAIStability)
		api.POST("/ai-stability/enable", h.EnableAIStability)
		api.POST("/ai-stability/disable", h.DisableAIStability)
	}
}

func (h *Handler) GetSensors(c *gin.Context) {
	readings := h.simulator.GetAlignedReadings(time.Now())
	result := make([]*models.SensorReading, 0, len(readings))
	for _, r := range readings {
		result = append(result, r)
	}
	c.JSON(http.StatusOK, gin.H{
		"sensors": result,
		"count":   len(result),
	})
}

func (h *Handler) GetTemperatureField(c *gin.Context) {
	field := h.fusion.ReconstructTemperatureField()
	c.JSON(http.StatusOK, field)
}

func (h *Handler) GetCombustionState(c *gin.Context) {
	state := h.fusion.DetectInstability()
	c.JSON(http.StatusOK, state)
}

func (h *Handler) GetEfficiency(c *gin.Context) {
	eff := h.fusion.AnalyzeEfficiency()
	c.JSON(http.StatusOK, eff)
}

func (h *Handler) GetAlarms(c *gin.Context) {
	activeOnly := c.Query("active") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if activeOnly {
		c.JSON(http.StatusOK, gin.H{
			"alarms": h.alarmMgr.GetActive(),
			"count":  h.alarmMgr.ActiveCount(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alarms": h.alarmMgr.GetAll(limit),
		"count":  h.alarmMgr.Count(),
	})
}

func (h *Handler) AcknowledgeAlarm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alarm id"})
		return
	}
	if h.alarmMgr.Acknowledge(id) {
		c.JSON(http.StatusOK, gin.H{"message": "alarm acknowledged"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "alarm not found"})
	}
}

func (h *Handler) GetSystemStatus(c *gin.Context) {
	status := models.SystemStatus{
		Online:          true,
		SensorCount:     len(h.simulator.GetReadings()),
		ActiveAlarms:    h.alarmMgr.ActiveCount(),
		AIServiceOnline: true,
		Uptime:          time.Since(h.startTime).Seconds(),
		Timestamp:       time.Now(),
	}
	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetTemperatureHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"history": []interface{}{},
		"message": "historical data endpoint - connect to database for full history",
	})
}

func (h *Handler) GetEfficiencyHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"history": []interface{}{},
		"message": "historical data endpoint - connect to database for full history",
	})
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	ws.ServeWs(h.hub, c.Writer, c.Request)
}

func (h *Handler) GetControlOutput(c *gin.Context) {
	c.JSON(http.StatusOK, h.ctrl.GetOutput())
}

func (h *Handler) EnableControl(c *gin.Context) {
	h.ctrl.SetEnabled(true)
	c.JSON(http.StatusOK, gin.H{"message": "combustion control enabled", "enabled": true})
}

func (h *Handler) DisableControl(c *gin.Context) {
	h.ctrl.SetEnabled(false)
	c.JSON(http.StatusOK, gin.H{"message": "combustion control disabled", "enabled": false})
}

func (h *Handler) SetControlTargets(c *gin.Context) {
	var req struct {
		TargetTemp     float64 `json:"target_temp"`
		TargetFuelAir  float64 `json:"target_fuel_air"`
		TargetPressure float64 `json:"target_pressure"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	h.ctrl.SetTargets(req.TargetTemp, req.TargetFuelAir, req.TargetPressure)
	c.JSON(http.StatusOK, gin.H{"message": "targets updated"})
}

func (h *Handler) SetControlMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	h.ctrl.SetMode(req.Mode)
	c.JSON(http.StatusOK, gin.H{"message": "mode updated", "mode": req.Mode})
}

func (h *Handler) GetOptimization(c *gin.Context) {
	c.JSON(http.StatusOK, h.optimizer.GetResult())
}

func (h *Handler) SetOperatingMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	validModes := map[string]bool{
		"startup": true, "normal": true, "peak_load": true,
		"low_load": true, "shutdown": true, "emergency": true,
	}
	if !validModes[req.Mode] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode", "valid_modes": []string{"startup", "normal", "peak_load", "low_load", "shutdown", "emergency"}})
		return
	}
	h.optimizer.SetMode(req.Mode)
	c.JSON(http.StatusOK, gin.H{"message": "operating mode updated", "mode": req.Mode})
}

func (h *Handler) GetEmission(c *gin.Context) {
	c.JSON(http.StatusOK, h.emission.GetResult())
}

func (h *Handler) SetEmissionLimits(c *gin.Context) {
	var req struct {
		NOxLimit float64 `json:"nox_limit"`
		COLimit  float64 `json:"co_limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	h.emission.SetLimits(req.NOxLimit, req.COLimit)
	c.JSON(http.StatusOK, gin.H{"message": "emission limits updated"})
}

func (h *Handler) GetThermalBalance(c *gin.Context) {
	c.JSON(http.StatusOK, h.balancer.GetResult())
}

func (h *Handler) GetAIStability(c *gin.Context) {
	c.JSON(http.StatusOK, h.aiStability.GetResult())
}

func (h *Handler) EnableAIStability(c *gin.Context) {
	h.aiStability.SetEnabled(true)
	c.JSON(http.StatusOK, gin.H{"message": "AI stability control enabled", "enabled": true})
}

func (h *Handler) DisableAIStability(c *gin.Context) {
	h.aiStability.SetEnabled(false)
	c.JSON(http.StatusOK, gin.H{"message": "AI stability control disabled", "enabled": false})
}
