package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSignVerify(t *testing.T) {
	m := New([]byte("hunter2-secret"), time.Minute)
	tok, err := m.Sign(Claims{"sub": "user_abc", "kyc_level": 2})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims["sub"] != "user_abc" {
		t.Fatalf("sub = %v", claims["sub"])
	}
}

func TestVerifyRejectsTamper(t *testing.T) {
	m := New([]byte("hunter2-secret"), time.Minute)
	tok, _ := m.Sign(Claims{"sub": "user_abc"})
	if _, err := m.Verify(tok[:len(tok)-1] + "A"); err == nil {
		t.Fatal("tampered token must fail")
	}
}

func TestVerifyRejectsExpired(t *testing.T) {
	m := New([]byte("hunter2-secret"), time.Minute)
	tok, _ := m.Sign(Claims{"sub": "user_abc"})
	// wheels of time
	m.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	if _, err := m.Verify(tok); err == nil {
		t.Fatal("expired token must fail")
	}
}

func TestMiddlewareInjectsUser(t *testing.T) {
	m := New([]byte("hunter2-secret"), time.Minute)
	tok, _ := m.Sign(Claims{"sub": "user_zz9"})

	var got string
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = UserID(r)
	}))
	req := httptest.NewRequest(http.MethodGet, "/v1/wallet/balance", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got != "user_zz9" {
		t.Fatalf("user_id = %q", got)
	}
}

func TestTOTP(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTOTP(secret)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Now()
	code := tp.Code(at)
	if !tp.Valid(code, at, 1) {
		t.Fatal("generated code must validate")
	}
	if tp.Valid("000000", at, 1) {
		t.Fatal("wrong code must fail")
	}
}