package common

import (
	"strings"
	"testing"
	"time"
)

func TestTOTPKnownVector(t *testing.T) {
	// RFC 6238 test secret for "12345678901234567890", SHA-1.
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	code, err := totpAt(secret, time.Unix(59, 0))
	if err != nil {
		t.Fatalf("gagal menghitung TOTP: %v", err)
	}
	if code != "287082" {
		t.Fatalf("kode TOTP salah: %s", code)
	}
	if !ValidateTOTP(secret, code, time.Unix(59, 0)) {
		t.Fatal("kode pada waktu yang sama harus valid")
	}
	if ValidateTOTP(secret, "000000", time.Unix(59, 0)) {
		t.Fatal("kode salah harus ditolak")
	}
}

func TestGenerateTOTPSecretAndURL(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("gagal membuat secret: %v", err)
	}
	if len(secret) < 32 {
		t.Fatalf("secret terlalu pendek: %s", secret)
	}
	u := TOTPURL("root", secret)
	if !strings.HasPrefix(u, "otpauth://totp/") {
		t.Fatalf("URL TOTP salah: %s", u)
	}
	if ValidateTOTP(secret, "123456", time.Now()) {
		t.Fatal("kode acak tidak boleh valid")
	}
}
