package common

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/jhopan/Panway/common/config"
)

func TestTokenKeyEncryptionRoundTrip(t *testing.T) {
	previousKey, previousFile := config.TokenEncryptionKey, config.TokenKeyFile
	config.TokenEncryptionKey = "unit-test-encryption-key"
	config.TokenKeyFile = filepath.Join(t.TempDir(), "token.key")
	t.Cleanup(func() {
		config.TokenEncryptionKey, config.TokenKeyFile = previousKey, previousFile
	})
	tokenKeyState.Once = sync.Once{}
	tokenKeyState.key, tokenKeyState.err = nil, nil

	plain := "sk-0123456789abcdef"
	encrypted, err := EncryptTokenKey(plain)
	if err != nil {
		t.Fatalf("gagal enkripsi: %v", err)
	}
	if !IsEncryptedTokenKey(encrypted) || encrypted == plain {
		t.Fatalf("nilai tersimpan bukan ciphertext: %q", encrypted)
	}
	decrypted, err := DecryptTokenKey(encrypted)
	if err != nil {
		t.Fatalf("gagal dekripsi: %v", err)
	}
	if decrypted != plain {
		t.Fatalf("hasil dekripsi salah: %q", decrypted)
	}

	hash, err := TokenKeyHash(plain)
	if err != nil {
		t.Fatalf("gagal hash: %v", err)
	}
	if len(hash) != 64 {
		t.Fatalf("hash harus 64 hex char, dapat %d", len(hash))
	}
	other, _ := TokenKeyHash("sk-other")
	if hash == other {
		t.Fatal("hash dua key berbeda tidak boleh sama")
	}
}

func TestDecryptPlaintextTokenKeyIsPassthrough(t *testing.T) {
	got, err := DecryptTokenKey("legacy-plain")
	if err != nil || got != "legacy-plain" {
		t.Fatalf("nilai lama harus dikembalikan apa adanya: %q %v", got, err)
	}
}

func TestTokenKeyFileIsCreated(t *testing.T) {
	previousKey, previousFile := config.TokenEncryptionKey, config.TokenKeyFile
	config.TokenEncryptionKey = ""
	config.TokenKeyFile = filepath.Join(t.TempDir(), "nested", "token.key")
	t.Cleanup(func() {
		config.TokenEncryptionKey, config.TokenKeyFile = previousKey, previousFile
	})
	tokenKeyState.Once = sync.Once{}
	tokenKeyState.key, tokenKeyState.err = nil, nil

	key, err := TokenEncryptionKeyMaterial()
	if err != nil {
		t.Fatalf("gagal membuat key file: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("key harus 32 byte, dapat %d", len(key))
	}
	data, err := os.ReadFile(config.TokenKeyFile)
	if err != nil {
		t.Fatalf("key file tidak dibuat: %v", err)
	}
	if len(data) != 32 {
		t.Fatalf("isi key file salah: %d byte", len(data))
	}
}
