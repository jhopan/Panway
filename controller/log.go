package controller

import (
	"encoding/csv"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/ctxkey"
	"github.com/jhopan/Panway/common/logger"
	"github.com/jhopan/Panway/model"
)

func GetAllLogs(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	username := c.Query("username")
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))
	logs, err := model.GetAllLogs(logType, startTimestamp, endTimestamp, modelName, username, tokenName, p*config.ItemsPerPage, config.ItemsPerPage, channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
	return
}

func GetUserLogs(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	userId := c.GetInt(ctxkey.Id)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	logs, err := model.GetUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, p*config.ItemsPerPage, config.ItemsPerPage)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
	return
}

func SearchAllLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	logs, err := model.SearchAllLogs(keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
	return
}

func SearchUserLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	userId := c.GetInt(ctxkey.Id)
	logs, err := model.SearchUserLogs(userId, keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
	return
}

func GetLogsStat(c *gin.Context) {
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	username := c.Query("username")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))
	quotaNum := model.SumUsedQuota(logType, startTimestamp, endTimestamp, modelName, username, tokenName, channel)
	//tokenNum := model.SumUsedToken(logType, startTimestamp, endTimestamp, modelName, username, "")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"quota": quotaNum,
			//"token": tokenNum,
		},
	})
	return
}

func GetLogsSelfStat(c *gin.Context) {
	username := c.GetString(ctxkey.Username)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))
	quotaNum := model.SumUsedQuota(logType, startTimestamp, endTimestamp, modelName, username, tokenName, channel)
	//tokenNum := model.SumUsedToken(logType, startTimestamp, endTimestamp, modelName, username, tokenName)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"quota": quotaNum,
			//"token": tokenNum,
		},
	})
	return
}

// ExportLogs streams the filtered logs as CSV. The console uses it for
// bookkeeping and invoices, so the rows are the same rows the log page shows.
// The export is capped so one click cannot read an unbounded SQLite table.
const exportRowLimit = 50000

func ExportMyLogs(c *gin.Context) {
	exportLogs(c, false)
}

func ExportAllLogs(c *gin.Context) {
	exportLogs(c, true)
}

// GetReceipt renders a printable receipt for one top-up or plan purchase.
// The owner and admins may open it; anything else is rejected.
func GetReceipt(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var log model.Log
	if err := model.DB.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "receipt not found"})
		return
	}
	userId := c.GetInt(ctxkey.Id)
	if log.UserId != userId && !model.IsAdmin(userId) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "not allowed"})
		return
	}
	var user model.User
	model.DB.Select("username", "display_name").First(&user, log.UserId)

	data := struct {
		SystemName  string
		ReceiptID   int
		Date        string
		Username    string
		DisplayName string
		Content     string
		Quota       int
	}{}
	data.SystemName = config.SystemName
	data.ReceiptID = log.Id
	data.Date = time.Unix(log.CreatedAt, 0).Format("2006-01-02 15:04:05")
	data.Username = user.Username
	data.DisplayName = user.DisplayName
	data.Content = log.Content
	data.Quota = log.Quota

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	if err := receiptTemplate.Execute(c.Writer, data); err != nil {
		logger.SysError("receipt render failed: " + err.Error())
	}
}

var receiptTemplate = template.Must(template.New("receipt").Parse(`<!doctype html>
<html><head><meta charset="utf-8"><title>Receipt #{{.ReceiptID}}</title>
<style>
body{font-family:Inter,Arial,sans-serif;max-width:560px;margin:40px auto;color:#10131a;padding:0 16px}
h1{font-size:20px;margin:0 0 4px} .muted{color:#626b7c;font-size:13px}
table{width:100%;border-collapse:collapse;margin:24px 0} td{padding:10px 0;border-bottom:1px solid #e5e8ef;font-size:14px}
td:last-child{text-align:right;font-weight:600} .total{font-size:18px}
button{padding:9px 16px;border:1px solid #e5e8ef;border-radius:8px;background:#f6f7fa;cursor:pointer}
@media print{button{display:none}}
</style></head><body>
<h1>{{.SystemName}} receipt</h1>
<div class="muted">Receipt #{{.ReceiptID}} · {{.Date}}</div>
<table>
<tr><td>Customer</td><td>{{.DisplayName}} ({{.Username}})</td></tr>
<tr><td>Description</td><td>{{.Content}}</td></tr>
<tr class="total"><td>Quota credited</td><td>{{.Quota}}</td></tr>
</table>
<button onclick="window.print()">Print / save as PDF</button>
</body></html>`))

func exportLogs(c *gin.Context, allUsers bool) {
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")

	query := model.DB.Model(&model.Log{})
	if !allUsers {
		query = query.Where("user_id = ?", c.GetInt(ctxkey.Id))
	} else {
		if username := c.Query("username"); username != "" {
			query = query.Where("username = ?", username)
		}
		if channel, err := strconv.Atoi(c.Query("channel")); err == nil && channel != 0 {
			query = query.Where("channel_id = ?", channel)
		}
	}
	if logType != 0 {
		query = query.Where("type = ?", logType)
	}
	if startTimestamp > 0 {
		query = query.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp > 0 {
		query = query.Where("created_at <= ?", endTimestamp)
	}
	if tokenName != "" {
		query = query.Where("token_name = ?", tokenName)
	}
	if modelName != "" {
		query = query.Where("model_name = ?", modelName)
	}

	rows, err := query.Order("id desc").Limit(exportRowLimit).Rows()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()

	filename := "panway-usage-" + time.Now().Format("20060102") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()
	_ = writer.Write([]string{
		"time", "type", "model", "token_name", "prompt_tokens",
		"completion_tokens", "quota", "is_stream", "request_id", "content",
	})
	for rows.Next() {
		var log model.Log
		if err := model.DB.ScanRows(rows, &log); err != nil {
			continue
		}
		_ = writer.Write([]string{
			time.Unix(log.CreatedAt, 0).Format(time.RFC3339),
			exportLogType(log.Type),
			log.ModelName,
			log.TokenName,
			strconv.Itoa(log.PromptTokens),
			strconv.Itoa(log.CompletionTokens),
			strconv.Itoa(log.Quota),
			strconv.FormatBool(log.IsStream),
			log.RequestId,
			log.Content,
		})
	}
}

func exportLogType(logType int) string {
	switch logType {
	case model.LogTypeTopup:
		return "topup"
	case model.LogTypeConsume:
		return "usage"
	case model.LogTypeManage:
		return "admin"
	case model.LogTypeSystem:
		return "system"
	case model.LogTypeTest:
		return "test"
	default:
		return "unknown"
	}
}

func DeleteHistoryLogs(c *gin.Context) {
	targetTimestamp, _ := strconv.ParseInt(c.Query("target_timestamp"), 10, 64)
	if targetTimestamp == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "target timestamp is required",
		})
		return
	}
	count, err := model.DeleteOldLog(targetTimestamp)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    count,
	})
	return
}
