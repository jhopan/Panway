package common

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// TOTP (RFC 6238) with SHA-1, six digits and a 30 second step. That is what
// every authenticator app expects; no external dependency is needed.
const (
	totpDigits = 6
	totpPeriod = 30
	totpIssuer = "Panway"
)

// GenerateTOTPSecret returns a base32 secret for an authenticator app.
func GenerateTOTPSecret() (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw), nil
}

// TOTPURL builds the otpauth:// URI an authenticator app scans.
func TOTPURL(account, secret string) string {
	label := url.PathEscape(totpIssuer + ":" + account)
	query := url.Values{}
	query.Set("secret", secret)
	query.Set("issuer", totpIssuer)
	query.Set("algorithm", "SHA1")
	query.Set("digits", "6")
	query.Set("period", "30")
	return "otpauth://totp/" + label + "?" + query.Encode()
}

// ValidateTOTP accepts the current code and one step either side, which
// tolerates a small clock difference between the server and the phone.
func ValidateTOTP(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits || secret == "" {
		return false
	}
	for offset := int64(-1); offset <= 1; offset++ {
		want, err := totpAt(secret, now.Add(time.Duration(offset*totpPeriod)*time.Second))
		if err != nil {
			return false
		}
		if subtle.ConstantTimeCompare([]byte(want), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

func totpAt(secret string, at time.Time) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(secret))
	raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(normalized)
	if err != nil {
		return "", err
	}
	counter := uint64(at.Unix() / totpPeriod)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)
	mac := hmac.New(sha1.New, raw)
	mac.Write(buf)
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", value%1000000), nil
}
