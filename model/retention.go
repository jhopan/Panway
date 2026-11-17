package model

import (
	"context"
	"time"

	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/logger"
)

// StartLogCleanup deletes operation logs older than LOG_RETENTION_DAYS.
//
// Logs grow with every relay request, and on SQLite an unbounded logs table
// makes both the database file and the log page slower. Cleanup runs once at
// startup and then daily; 0 keeps every log.
func StartLogCleanup(ctx context.Context) {
	days := config.LogRetentionDays
	if days <= 0 {
		return
	}
	go func() {
		cleanupOldLogs(days)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cleanupOldLogs(days)
			}
		}
	}()
	logger.SysLogf("log retention enabled: older than %d days are removed", days)
}

// cleanupOldLogs deletes in chunks so one large delete cannot block SQLite
// writers for a long time. It returns the number of deleted rows.
func cleanupOldLogs(days int) int64 {
	daysValue := time.Duration(days) * 24 * time.Hour
	cutoff := time.Now().Add(-daysValue).Unix()
	db := LOG_DB
	if db == nil {
		db = DB
	}
	var total int64
	for {
		res := db.Exec("DELETE FROM logs WHERE id IN (SELECT id FROM logs WHERE created_at < ? LIMIT 1000)", cutoff)
		if res.Error != nil {
			logger.SysError("log cleanup failed: " + res.Error.Error())
			break
		}
		total += res.RowsAffected
		if res.RowsAffected < 1000 {
			break
		}
	}
	if total > 0 {
		logger.SysLogf("log cleanup removed %d rows older than %d days", total, days)
	}
	return total
}
