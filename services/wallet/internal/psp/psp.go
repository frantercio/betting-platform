package psp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type PSPClient interface {
	Deposit(userID string, amountCents int64, requestID string) (*DepositResponse, error)
	ConfirmWebhook(reqID string, txID string) error
	VerifySignature(payload, signature, secret string) bool
}

type DepositResponse struct {
	QRCode     string
	CopiaECola string
	ExpiresAt  string
}

type MockPSP struct {
	Secret string
	Store  map[string]DepositResponse
}

func NewMockPSP(secret string) *MockPSP {
	return &MockPSP{Secret: secret, Store: map[string]DepositResponse{}}
}

func (m *MockPSP) Deposit(userID string, amountCents int64, requestID string) (*DepositResponse, error) {
	resp := DepositResponse{
		QRCode:     "data:image/png;base64,FAKE",
		CopiaECola: "00020126580014BR.GOV.BCB.PIX0136" + requestID,
		ExpiresAt:  "2026-01-01T00:00:00Z",
	}
	m.Store[requestID] = resp
	return &resp, nil
}

func (m *MockPSP) ConfirmWebhook(reqID string, txID string) error {
	if _, ok := m.Store[reqID]; !ok {
		return fmt.Errorf("unknown request_id")
	}
	// in real implementation would call PSP API
	return nil
}

func (m *MockPSP) VerifySignature(payload, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
