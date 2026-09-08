// Package auth implements HS256 JWT signing/verification and TOTP (RFC 6238)
// using only the standard library, so the skeleton builds offline.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrMalformed = errors.New("malformed token")
	ErrExpired   = errors.New("token expired")
	ErrSignature = errors.New("invalid signature")
)

// Claims is a JSON object map (sub, role, kyc_level, ...).
type Claims map[string]any

// Manager holds the signing key and token TTL.
type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func New(secret []byte, ttl time.Duration) *Manager {
	return &Manager{secret: secret, ttl: ttl, now: time.Now}
}

// Sign issues a compact HS256 token: b64(header).b64(payload).sig
func (m *Manager) Sign(claims Claims) (string, error) {
	if sub, _ := claims["sub"].(string); sub == "" {
		return "", errors.New("claims.sub is required")
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims["iat"] = m.now().Unix()
	claims["exp"] = m.now().Add(m.ttl).Unix()
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	seg := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(seg))
	return seg + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// Verify decodes and validates signature + exp. Returns claims on success.
func (m *Manager) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrMalformed
	}
	header, payload, sig := parts[0], parts[1], parts[2]

	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(header + "." + payload))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil || !hmac.Equal(want, got) {
		return nil, ErrSignature
	}

	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, ErrMalformed
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, ErrMalformed
	}

	if exp, ok := claims["exp"].(float64); ok && m.now().Unix() > int64(exp) {
		return nil, ErrExpired
	}
	if _, ok := claims["sub"].(string); !ok {
		return nil, ErrMalformed
	}
	return claims, nil
}