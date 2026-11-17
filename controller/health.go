package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/model"
)

// Health is the unauthenticated health check used by the reverse proxy and
// monitoring. It reports only coarse status: no keys, quotas or customer data.
func Health(c *gin.Context) {
	database := "ok"
	channels := 0

	if model.DB == nil {
		database = "unavailable"
	} else {
		sqlDB, err := model.DB.DB()
		if err != nil {
			database = "unavailable"
		} else {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err != nil {
				database = "error"
			} else {
				channels = countEnabledChannels()
			}
		}
	}

	status := http.StatusOK
	state := "ok"
	if database != "ok" {
		status = http.StatusServiceUnavailable
		state = "degraded"
	}

	c.JSON(status, gin.H{
		"status":   state,
		"database": database,
		"channels": channels,
		"version":  common.Version,
	})
}

func countEnabledChannels() int {
	var count int64
	if model.DB == nil {
		return 0
	}
	model.DB.Model(&model.Channel{}).
		Where("status = ?", model.ChannelStatusEnabled).
		Count(&count)
	return int(count)
}
