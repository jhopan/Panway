package controller

import (
	"crypto/subtle"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/model"
)

// Metrics serves Prometheus text metrics. It is disabled by default and can be
// protected with METRICS_TOKEN when the port is reachable from a network that
// contains more than the scraper.
func Metrics(c *gin.Context) {
	if !config.MetricsEnabled {
		c.Status(http.StatusNotFound)
		return
	}
	if config.MetricsToken != "" {
		provided := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(config.MetricsToken)) != 1 {
			c.Status(http.StatusUnauthorized)
			return
		}
	}
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	var channels int64
	if model.DB != nil {
		model.DB.Model(&model.Channel{}).
			Where("status = ?", model.ChannelStatusEnabled).
			Count(&channels)
	}
	c.String(http.StatusOK, common.MetricsText()+
		"# HELP panway_channels_enabled Enabled upstream channels.\n"+
		"# TYPE panway_channels_enabled gauge\n"+
		"panway_channels_enabled "+strconv.FormatInt(channels, 10)+"\n")
}
