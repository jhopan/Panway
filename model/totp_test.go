package model

import (
	"strings"
	"testing"
)

func TestTotpSecretStoredEncryptedAndReadable(t *testing.T) {
	user := setupPlanTestDB(t)
	secret := "JBSWY3DPEHPK3PXP"
	if err := SetUserTotpSecret(user.Id, secret); err != nil {
		t.Fatalf("gagal menyimpan secret: %v", err)
	}
	var raw string
	DB.Model(&User{}).Where("id = ?", user.Id).Pluck("totp_secret", &raw)
	if !strings.HasPrefix(raw, "enc:v1:") {
		t.Fatalf("secret harus tersimpan terenkripsi, dapat %q", raw)
	}
	got, err := GetUserTotpSecret(user.Id)
	if err != nil || got != secret {
		t.Fatalf("secret tidak bisa dibaca kembali: %q %v", got, err)
	}
}

func TestTotpSecretPlaintextFallback(t *testing.T) {
	user := setupPlanTestDB(t)
	DB.Model(&User{}).Where("id = ?", user.Id).Update("totp_secret", "PLAINSECRET")
	got, err := GetUserTotpSecret(user.Id)
	if err != nil || got != "PLAINSECRET" {
		t.Fatalf("secret lama harus dibaca apa adanya: %q %v", got, err)
	}
}
