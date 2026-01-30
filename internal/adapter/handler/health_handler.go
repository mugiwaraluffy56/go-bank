package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gobank/internal/infrastructure/database"
)

type HealthHandler struct {
	db        *database.PostgresDB
	redis     *database.RedisDB
	startTime time.Time
}

func NewHealthHandler(db *database.PostgresDB, redis *database.RedisDB) *HealthHandler {
	return &HealthHandler{
		db:        db,
		redis:     redis,
		startTime: time.Now(),
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	checks := make(map[string]string)
	healthy := true

	if err := h.db.Ping(c.Request.Context()); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		healthy = false
	} else {
		checks["database"] = "healthy"
	}

	if err := h.redis.Ping(c.Request.Context()); err != nil {
		checks["redis"] = "unhealthy: " + err.Error()
		healthy = false
	} else {
		checks["redis"] = "healthy"
	}

	status := http.StatusOK
	statusText := "ready"
	if !healthy {
		status = http.StatusServiceUnavailable
		statusText = "not ready"
	}

	c.JSON(status, gin.H{
		"status":    statusText,
		"checks":    checks,
		"timestamp": time.Now().UTC(),
	})
}

func (h *HealthHandler) Info(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"service":    "gobank",
		"version":    "1.0.0",
		"go_version": runtime.Version(),
		"uptime":     time.Since(h.startTime).String(),
		"memory": gin.H{
			"alloc_mb":       m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"num_gc":         m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"timestamp":  time.Now().UTC(),
	})
}
