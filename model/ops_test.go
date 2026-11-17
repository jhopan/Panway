package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jhopan/Panway/common/config"
)

func setupOpsTestDB(t *testing.T) {
	t.Helper()
	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "ops.db")) + "?_busy_timeout=5000&_journal_mode=WAL"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gagal membuka db uji: %v", err)
	}
	previousDB, previousLogDB := DB, LOG_DB
	DB, LOG_DB = db, db
	t.Cleanup(func() { DB, LOG_DB = previousDB, previousLogDB })
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("gagal mengambil koneksi sql: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := DB.AutoMigrate(&Log{}, &Channel{}); err != nil {
		t.Fatalf("auto migrate gagal: %v", err)
	}
}

func TestBackupSQLiteWritesAndPrunesSnapshots(t *testing.T) {
	setupOpsTestDB(t)
	dir := t.TempDir()
	previousDir, previousKeep := config.SQLiteBackupDir, config.SQLiteBackupKeep
	config.SQLiteBackupDir, config.SQLiteBackupKeep = dir, 2
	t.Cleanup(func() {
		config.SQLiteBackupDir, config.SQLiteBackupKeep = previousDir, previousKeep
	})

	for i := 0; i < 3; i++ {
		backupSQLite()
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("gagal membaca folder backup: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("harus menyisakan 2 backup, dapat %d", len(entries))
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			t.Fatalf("gagal membaca info backup: %v", err)
		}
		if info.Size() == 0 {
			t.Fatalf("backup %s kosong", entry.Name())
		}
	}
}

func TestCleanupOldLogs(t *testing.T) {
	setupOpsTestDB(t)
	old := Log{CreatedAt: time.Now().Add(-40 * 24 * time.Hour).Unix(), Type: LogTypeConsume, Content: "old"}
	recent := Log{CreatedAt: time.Now().Unix(), Type: LogTypeConsume, Content: "recent"}
	if err := DB.Create(&old).Error; err != nil {
		t.Fatalf("gagal membuat log lama: %v", err)
	}
	if err := DB.Create(&recent).Error; err != nil {
		t.Fatalf("gagal membuat log baru: %v", err)
	}

	deleted := cleanupOldLogs(30)
	if deleted != 1 {
		t.Fatalf("harus menghapus 1 log, dapat %d", deleted)
	}
	var count int64
	DB.Model(&Log{}).Count(&count)
	if count != 1 {
		t.Fatalf("harus tersisa 1 log, dapat %d", count)
	}
	var remaining Log
	DB.First(&remaining)
	if remaining.Content != "recent" {
		t.Fatalf("log yang tersisa salah: %+v", remaining)
	}
	if again := cleanupOldLogs(30); again != 0 {
		t.Fatalf("cleanup kedua tidak boleh menghapus apa pun, dapat %d", again)
	}
}
