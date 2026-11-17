package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/ctxkey"
	"github.com/jhopan/Panway/model"
)

// auditf records an administrative change in the operation log. Only the
// action and its target are stored; secrets are never written here.
func auditf(c *gin.Context, format string, args ...any) {
	model.RecordLog(c.Request.Context(), c.GetInt(ctxkey.Id), model.LogTypeManage, fmt.Sprintf(format, args...))
}
