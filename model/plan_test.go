package model

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPlanTestDB(t *testing.T) User {
	t.Helper()
	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "plan.db")) + "?_busy_timeout=5000"
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
	if err := DB.AutoMigrate(&User{}, &Plan{}, &Log{}); err != nil {
		t.Fatalf("auto migrate gagal: %v", err)
	}
	user := User{
		Username: "plan-user", Password: "x", DisplayName: "Plan User",
		Role: RoleCommonUser, Status: UserStatusEnabled, Quota: 1000,
	}
	if err := DB.Create(&user).Error; err != nil {
		t.Fatalf("gagal membuat user: %v", err)
	}
	return user
}

func TestBuyPlanChargesAndExtends(t *testing.T) {
	user := setupPlanTestDB(t)
	plan := Plan{Name: "Starter", PriceIDR: 500, Quota: 500, DurationDays: 30, Status: PlanStatusEnabled}
	if err := plan.Insert(); err != nil {
		t.Fatalf("gagal membuat plan: %v", err)
	}

	quota, expiresAt, err := BuyPlan(user.Id, &plan)
	if err != nil {
		t.Fatalf("buy plan gagal: %v", err)
	}
	if quota != 1000 {
		t.Fatalf("kuota akhir salah: %d", quota)
	}
	if expiresAt < time.Now().Add(29*24*time.Hour).Unix() {
		t.Fatalf("expiry terlalu pendek: %d", expiresAt)
	}
	active, err := GetActiveUserPlan(user.Id)
	if err != nil || active == nil || active.Id != plan.Id {
		t.Fatalf("plan tidak aktif: plan=%+v err=%v", active, err)
	}
	var logs int64
	DB.Model(&Log{}).Where("user_id = ? AND type = ?", user.Id, LogTypeTopup).Count(&logs)
	if logs != 1 {
		t.Fatalf("harus ada 1 log pembelian, dapat %d", logs)
	}
}

func TestBuyPlanRejectsInsufficientQuota(t *testing.T) {
	user := setupPlanTestDB(t)
	plan := Plan{Name: "Pro", PriceIDR: 2000, Quota: 2000, DurationDays: 30, Status: PlanStatusEnabled}
	if err := plan.Insert(); err != nil {
		t.Fatalf("gagal membuat plan: %v", err)
	}
	if _, _, err := BuyPlan(user.Id, &plan); err == nil {
		t.Fatal("pembelian melebihi saldo harus ditolak")
	}
	var reloaded User
	DB.First(&reloaded, user.Id)
	if reloaded.Quota != 1000 {
		t.Fatalf("kuota tidak boleh berubah: %d", reloaded.Quota)
	}
}

func TestExpirePlansClearsExpiredPlan(t *testing.T) {
	user := setupPlanTestDB(t)
	plan := Plan{Name: "Expired", PriceIDR: 100, Quota: 100, DurationDays: 1, Status: PlanStatusEnabled}
	if err := plan.Insert(); err != nil {
		t.Fatalf("gagal membuat plan: %v", err)
	}
	DB.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]interface{}{
		"plan_id": plan.Id, "plan_expires_at": time.Now().Add(-time.Hour).Unix(),
	})
	active, err := GetActiveUserPlan(user.Id)
	if err != nil {
		t.Fatalf("gagal membaca plan: %v", err)
	}
	if active != nil {
		t.Fatal("plan kedaluwarsa tidak boleh dianggap aktif")
	}
	if removed := expirePlans(); removed != 1 {
		t.Fatalf("harus membersihkan 1 user, dapat %d", removed)
	}
	var reloaded User
	DB.First(&reloaded, user.Id)
	if reloaded.PlanId != 0 {
		t.Fatalf("plan_id harus 0, dapat %d", reloaded.PlanId)
	}
}
