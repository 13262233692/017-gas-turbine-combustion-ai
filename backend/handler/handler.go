package handler

import (
	"net/http"
	"strconv"
	"time"

	"gas-turbine-combustion-ai/alarm"
	"gas-turbine-combustion-ai/fusion"
	"gas-turbine-combustion-ai/models"
	"gas-turbine-combustion-ai/sensor"
	"gas-turbine-combustion-ai/ws"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	simulator *sensor.Simulator
	fusion    *fusion.FusionEngine
	alarmMgr  *alarm.Manager
	hub       *ws.Hub
	startTime time.Time
}

func NewHandler(sim *sensor.Simulator, fusionEng *fusion.FusionEngine, alarmMgr *alarm.Manager, hub *ws.Hub) *Handler {
	return &Handler{
		simulator: sim,
		fusion:    fusionEng,
		alarmMgr:  alarmMgr,
		hub:       hub,
		startTime: time.Now(),
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
	}
}

func (h *Handler) GetSensors(c *gin.Context) {
	readings := h.simulator.GetReadings()
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
