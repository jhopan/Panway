package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/logger"
)

// StartSQLiteBackup writes a consistent SQLite snapshot on a schedule.
//
// SQLite has no separate backup server. VACUUM INTO produces a compact copy
// while the database keeps serving requests, which is enough for a single
// node. The backup runs immediately at startup and then every
// SQLITE_BACKUP_INTERVAL_HOURS; SQLITE_BACKUP_KEEP oldest files are removed.
func StartSQLiteBackup(ctx context.Context) {
	if !common.UsingSQLite || !config.SQLiteBackupEnabled {
		return
	}
	interval := time.Duration(config.SQLiteBackupIntervalHours) * time.Hour
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	go func() {
		backupSQLite()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				backupSQLite()
			}
		}
	}()
	logger.SysLogf("sqlite backup enabled: dir %s, every %s, keep %d",
		config.SQLiteBackupDir, interval, config.SQLiteBackupKeep)
}

func backupSQLite() {
	if DB == nil {
		return
	}
	dir := config.SQLiteBackupDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		logger.SysError("sqlite backup: cannot create dir: " + err.Error())
		return
	}
	target := uniqueBackupPath(dir)
	// VACUUM INTO takes a string literal, so the path is escaped rather than
	// bound as a parameter.
	escaped := strings.ReplaceAll(target, "'", "''")
	if err := DB.Exec("VACUUM INTO '" + escaped + "'").Error; err != nil {
		logger.SysError("sqlite backup failed: " + err.Error())
		return
	}
	logger.SysLog("sqlite backup written to " + target)
	pruneSQLiteBackups(dir, config.SQLiteBackupKeep)
}

func uniqueBackupPath(dir string) string {
	stamp := time.Now().Format("20060102-150405")
	for i := 0; ; i++ {
		name := fmt.Sprintf("panway-%s.db", stamp)
		if i > 0 {
			name = fmt.Sprintf("panway-%s-%d.db", stamp, i)
		}
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
}

// pruneSQLiteBackups keeps the newest keep files, deleting the rest.
func pruneSQLiteBackups(dir string, keep int) {
	if keep <= 0 {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.SysError("sqlite backup: cannot list dir: " + err.Error())
		return
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, "panway-") || !strings.HasSuffix(name, ".db") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for len(names) > keep {
		old := names[0]
		if err := os.Remove(filepath.Join(dir, old)); err != nil {
			logger.SysError("sqlite backup: cannot remove " + old + ": " + err.Error())
		}
		names = names[1:]
	}
}
