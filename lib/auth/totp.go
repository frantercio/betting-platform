package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// TOTP implements RFC 6238 time-based one-time passwords (6 digits, 30s).
type TOTP struct {
	secret []byte // decoded raw key
	digits int
	step   time.Duration
}

// NewTOTP builds a TOTP from a base32 secret (same used to render the QR code).
func NewTOTP(secretBase32 string) (*TOTP, error) {
	clean := strings.ToUpper(strings.TrimSpace(secretBase32))
	raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("decode base32 secret: %w", err)
	}
	return &TOTP{secret: raw, digits: 6, step: 30 * time.Second}, nil
}

// Code returns the one-time password for the given time.
func (t *TOTP) Code(at time.Time) string {
	counter := uint64(at.Unix() / int64(t.step.Seconds()))

	mac := hmac.New(sha1.New, t.secret)
	if len(t.secret) < 16 {
		mac = hmac.New(sha256.New, t.secret)
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)

	// dynamic truncation (RFC 4226 / 6238)
	off := sum[len(sum)-1] & 0x0f
	bin := (int64(sum[off])&0x7f)<<24 |
		int64(sum[off+1])<<16 |
		int64(sum[off+2])<<8 |
		int64(sum[off+3])
	mod := int32(bin % int64(pow10(t.digits)))
	return fmt.Sprintf("%0*d", t.digits, mod)
}

// Valid checks a submitted code within a +/-1 step window (drift tolerance).
func (t *TOTP) Valid(code string, at time.Time, skew int) bool {
	for i := -skew; i <= skew; i++ {
		shifted := at.Add(t.step * time.Duration(i))
		if hmacEqual(t.Code(shifted), code) {
			return true
		}
	}
	return false
}

func hmacEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// GenerateSecret returns a random base32 secret for provisioning.
func GenerateSecret() (string, error) {
	var b [20]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b[:]), nil
}

func pow10(n int) int64 {
	p := int64(1)
	for i := 0; i < n; i++ {
		p *= 10
	}
	return p
}