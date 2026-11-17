package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
)

func TestMain(m *testing.M) {
	config.TokenEncryptionKey = "model-test-encryption-key"
	common.RedisEnabled = false
	os.Exit(m.Run())
}

func setupTokenKeyTestDB(t *testing.T) {
	t.Helper()
	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "token.db")) + "?_busy_timeout=5000"
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
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := DB.AutoMigrate(&Token{}); err != nil {
		t.Fatalf("auto migrate gagal: %v", err)
	}
}

func TestTokenInsertStoresCiphertextAndValidatesByHash(t *testing.T) {
	setupTokenKeyTestDB(t)
	plain := "sk-insert-test-key"
	token := Token{UserId: 1, Key: plain, Name: "t", Status: TokenStatusEnabled, RemainQuota: 1000, ExpiredTime: -1}
	if err := token.Insert(); err != nil {
		t.Fatalf("gagal insert token: %v", err)
	}
	var stored Token
	if err := DB.First(&stored, token.Id).Error; err != nil {
		t.Fatalf("gagal membaca token: %v", err)
	}
	if !common.IsEncryptedTokenKey(stored.Key) {
		t.Fatalf("key di DB harus ciphertext, dapat %q", stored.Key)
	}
	if stored.KeyHash == "" {
		t.Fatal("key_hash harus terisi")
	}
	if stored.DecryptedKey() != plain {
		t.Fatalf("dekripsi salah: %q", stored.DecryptedKey())
	}
	validated, err := ValidateUserToken(plain)
	if err != nil {
		t.Fatalf("token valid ditolak: %v", err)
	}
	if validated.Id != token.Id {
		t.Fatalf("token yang ditemukan salah: %d", validated.Id)
	}
	if _, err := ValidateUserToken("sk-wrong"); err == nil {
		t.Fatal("token salah harus ditolak")
	}
}

func TestValidateLegacyPlaintextTokenMigratesIt(t *testing.T) {
	setupTokenKeyTestDB(t)
	legacy := Token{UserId: 1, Key: "legacy-plain-key", Name: "legacy", Status: TokenStatusEnabled, RemainQuota: 1000, ExpiredTime: -1}
	if err := DB.Create(&legacy).Error; err != nil {
		t.Fatalf("gagal membuat token lama: %v", err)
	}
	if _, err := ValidateUserToken("legacy-plain-key"); err != nil {
		t.Fatalf("token lama valid ditolak: %v", err)
	}
	var stored Token
	DB.First(&stored, legacy.Id)
	if !common.IsEncryptedTokenKey(stored.Key) || stored.KeyHash == "" {
		t.Fatalf("token lama tidak dimigrasi: key=%q hash=%q", stored.Key, stored.KeyHash)
	}
	if !strings.HasPrefix(stored.Key, "enc:v1:") {
		t.Fatalf("format ciphertext tidak dikenal: %q", stored.Key)
	}
	if _, err := ValidateUserToken("legacy-plain-key"); err != nil {
		t.Fatalf("token setelah migrasi harus tetap valid: %v", err)
	}
}

func TestMigrateTokenKeysScansAllRows(t *testing.T) {
	setupTokenKeyTestDB(t)
	for _, key := range []string{"legacy-one", "legacy-two"} {
		if err := DB.Create(&Token{UserId: 1, Key: key, Name: key, Status: TokenStatusEnabled, RemainQuota: 1000, ExpiredTime: -1}).Error; err != nil {
			t.Fatalf("gagal membuat token: %v", err)
		}
	}
	MigrateTokenKeys()
	var tokens []Token
	DB.Find(&tokens)
	if len(tokens) != 2 {
		t.Fatalf("jumlah token salah: %d", len(tokens))
	}
	for _, token := range tokens {
		if !common.IsEncryptedTokenKey(token.Key) || token.KeyHash == "" {
			t.Fatalf("token belum dimigrasi: %+v", token)
		}
	}
}
