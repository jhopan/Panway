package model

import (
	"path/filepath"
	"sync"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupQuotaTestDB(t *testing.T) (User, Token) {
	t.Helper()
	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "quota.db")) + "?_busy_timeout=5000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gagal membuka db uji: %v", err)
	}
	previous := DB
	DB = db
	t.Cleanup(func() { DB = previous })
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("gagal mengambil koneksi sql: %v", err)
	}
	// Close before t.TempDir removes the folder; Windows refuses to delete an
	// open database file.
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := DB.AutoMigrate(&User{}, &Token{}); err != nil {
		t.Fatalf("auto migrate gagal: %v", err)
	}
	user := User{
		Username: "quota-user", Password: "x", DisplayName: "Quota User",
		Role: RoleCommonUser, Status: UserStatusEnabled, Quota: 1_000_000,
	}
	if err := DB.Create(&user).Error; err != nil {
		t.Fatalf("gagal membuat user: %v", err)
	}
	token := Token{
		UserId: user.Id, Key: "quota-test-key", Status: TokenStatusEnabled,
		Name: "quota-test", RemainQuota: 10, ExpiredTime: -1,
	}
	if err := DB.Create(&token).Error; err != nil {
		t.Fatalf("gagal membuat token: %v", err)
	}
	return user, token
}

func readToken(t *testing.T, id int) Token {
	t.Helper()
	var token Token
	if err := DB.First(&token, id).Error; err != nil {
		t.Fatalf("gagal membaca token: %v", err)
	}
	return token
}

func TestPreConsumeTokenQuotaRejectsSpendAboveBalance(t *testing.T) {
	_, token := setupQuotaTestDB(t)

	err := PreConsumeTokenQuota(token.Id, 25)
	if err == nil {
		t.Fatal("saldo 10 tidak boleh lolos pre-consume 25")
	}
	got := readToken(t, token.Id)
	if got.RemainQuota != 10 || got.UsedQuota != 0 {
		t.Fatalf("saldo tidak boleh berubah: remain %d used %d", got.RemainQuota, got.UsedQuota)
	}
}

func TestPreConsumeTokenQuotaHoldsBalance(t *testing.T) {
	user, token := setupQuotaTestDB(t)

	if err := PreConsumeTokenQuota(token.Id, 5); err != nil {
		t.Fatalf("pre-consume 5 dari saldo 10 harus berhasil: %v", err)
	}
	got := readToken(t, token.Id)
	if got.RemainQuota != 5 || got.UsedQuota != 5 {
		t.Fatalf("saldo salah: remain %d used %d", got.RemainQuota, got.UsedQuota)
	}
	var reloaded User
	if err := DB.First(&reloaded, user.Id).Error; err != nil {
		t.Fatalf("gagal membaca user: %v", err)
	}
	if reloaded.Quota != 999_995 {
		t.Fatalf("kuota user harus ikut dipotong: %d", reloaded.Quota)
	}
}

func TestPostConsumeTokenQuotaChargesAtMostRemaining(t *testing.T) {
	_, token := setupQuotaTestDB(t)
	if err := DB.Model(&token).Updates(map[string]interface{}{"remain_quota": 5}).Error; err != nil {
		t.Fatalf("gagal menyetel saldo: %v", err)
	}

	if err := PostConsumeTokenQuota(token.Id, 12); err != nil {
		t.Fatalf("post-consume harus berhasil: %v", err)
	}
	got := readToken(t, token.Id)
	if got.RemainQuota != 0 {
		t.Fatalf("saldo tidak boleh minus: %d", got.RemainQuota)
	}
	if got.UsedQuota != 5 {
		t.Fatalf("hanya sisa saldo yang boleh ditagih: %d", got.UsedQuota)
	}
}

func TestPostConsumeTokenQuotaConcurrentStaysNonNegative(t *testing.T) {
	_, token := setupQuotaTestDB(t)
	if err := DB.Model(&token).Updates(map[string]interface{}{"remain_quota": 5}).Error; err != nil {
		t.Fatalf("gagal menyetel saldo: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = PostConsumeTokenQuota(token.Id, 3)
		}()
	}
	wg.Wait()

	got := readToken(t, token.Id)
	if got.RemainQuota < 0 {
		t.Fatalf("saldo minus setelah request paralel: %d", got.RemainQuota)
	}
	if got.RemainQuota+got.UsedQuota != 5 {
		t.Fatalf("saldo tidak seimbang: remain %d used %d", got.RemainQuota, got.UsedQuota)
	}
}
