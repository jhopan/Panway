package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jhopan/Panway/common/config"
)

const tokenKeyCipherPrefix = "enc:v1:"

var tokenKeyState struct {
	sync.Once
	key []byte
	err error
}

// TokenEncryptionKeyMaterial loads the AES/HMAC key. A configured
// TOKEN_ENCRYPTION_KEY is hashed to 32 bytes; otherwise a key file is created
// on first use with 0600 permissions.
func TokenEncryptionKeyMaterial() ([]byte, error) {
	tokenKeyState.Do(func() {
		if config.TokenEncryptionKey != "" {
			sum := sha256.Sum256([]byte(config.TokenEncryptionKey))
			tokenKeyState.key = sum[:]
			return
		}
		path := config.TokenKeyFile
		if data, err := os.ReadFile(path); err == nil {
			if len(data) != 32 {
				tokenKeyState.err = errors.New("token key file must contain exactly 32 bytes")
				return
			}
			tokenKeyState.key = data
			return
		} else if !os.IsNotExist(err) {
			tokenKeyState.err = err
			return
		}
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			tokenKeyState.err = err
			return
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			tokenKeyState.err = err
			return
		}
		if err := os.WriteFile(path, key, 0o600); err != nil {
			tokenKeyState.err = err
			return
		}
		tokenKeyState.key = key
	})
	return tokenKeyState.key, tokenKeyState.err
}

func IsEncryptedTokenKey(value string) bool {
	return strings.HasPrefix(value, tokenKeyCipherPrefix)
}

// TokenKeyHash is a keyed hash used to look a token up without storing its
// plaintext. HMAC with the same secret means a leaked hash cannot be reversed
// without the key.
func TokenKeyHash(plain string) (string, error) {
	key, err := TokenEncryptionKeyMaterial()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("panway-token:" + plain))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// EncryptTokenKey seals a plaintext API key for storage.
func EncryptTokenKey(plain string) (string, error) {
	key, err := TokenEncryptionKeyMaterial()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return tokenKeyCipherPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptTokenKey opens a stored key. Values written before encryption was
// enabled are returned unchanged so the console can still show them.
func DecryptTokenKey(stored string) (string, error) {
	if !IsEncryptedTokenKey(stored) {
		return stored, nil
	}
	key, err := TokenEncryptionKeyMaterial()
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, tokenKeyCipherPrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("encrypted token key is truncated")
	}
	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
