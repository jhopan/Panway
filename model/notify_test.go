package model

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
)

func TestNotifyLowBalanceSendsOncePerInterval(t *testing.T) {
	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "notify.db")) + "?_busy_timeout=5000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gagal membuka db uji: %v", err)
	}
	previousDB := DB
	DB = db
	t.Cleanup(func() { DB = previousDB })
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := DB.AutoMigrate(&User{}); err != nil {
		t.Fatalf("auto migrate gagal: %v", err)
	}

	user := User{
		Username: "low-user", Password: "x", DisplayName: "Low",
		Role: RoleCommonUser, Status: UserStatusEnabled, Quota: 10,
		TelegramChatId: "12345", TelegramVerified: true,
	}
	if err := DB.Create(&user).Error; err != nil {
		t.Fatalf("gagal membuat user: %v", err)
	}

	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		if payload["chat_id"] != "12345" {
			t.Errorf("chat_id salah: %v", payload["chat_id"])
		}
		if text, _ := payload["text"].(string); !strings.Contains(text, "balance is low") {
			t.Errorf("pesan salah: %q", text)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()

	previousThreshold := config.LowBalanceThreshold
	config.LowBalanceThreshold = 100
	t.Cleanup(func() { config.LowBalanceThreshold = previousThreshold })
	lowBalanceLastSent.Lock()
	lowBalanceLastSent.byUser = map[int]time.Time{}
	lowBalanceLastSent.Unlock()

	client := common.NewTelegramClient("test-token", server.URL)
	notifyLowBalance(client, time.Hour)
	notifyLowBalance(client, time.Hour)

	if calls.Load() != 1 {
		t.Fatalf("reminder harus dikirim sekali per interval, dapat %d", calls.Load())
	}
}
