package bet

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

type memWallet struct {
	mu     sync.Mutex
	users  map[string]int64
	events []string
}

func newMemWallet() *memWallet {
	return &memWallet{users: map[string]int64{}}
}
func (w *memWallet) Debit(userID, requestID string, cents int64, ref string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.users[userID] < cents {
		return fmt.Errorf("insufficient funds")
	}
	w.users[userID] -= cents
	w.events = append(w.events, "debit:"+requestID+":"+ref)
	return nil
}
func (w *memWallet) Credit(userID, requestID string, cents int64, ref string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.users[userID] += cents
	w.events = append(w.events, "credit:"+requestID+":"+ref)
	return nil
}
func (w *memWallet) balance(userID string) int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.users[userID]
}
func (w *memWallet) seed(userID string, cents int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.users[userID] += cents
}

type stubRG struct{ max int64 }

func (g stubRG) CheckBet(userID string, stake int64) error {
	if stake > g.max {
		return ErrStakeTooLarge
	}
	return nil
}

func openMarket(_ string) (float64, bool) { return 1.85, false }

func TestPlaceDebitsAndCreates(t *testing.T) {
	s := NewMemStore()
	w := newMemWallet()
	w.seed("user_1", 100000)
	b := &Bet{UserID: "user_1", EventID: "SP-1", Market: "FTR", Selection: "HOME", Odds: 1.85, Stake: 10000}

	err := Place(s, w, stubRG{max: 50000}, openMarket, b, "req-1")
	if err != nil {
		t.Fatal(err)
	}
	if b.ID == 0 {
		t.Fatal("bet must get an id")
	}
	if b.Status != Open {
		t.Fatalf("status = %s", b.Status)
	}
	if w.balance("user_1") != 90000 {
		t.Fatalf("balance = %d, want 90000", w.balance("user_1"))
	}
}

func TestPlaceRejectsOddsChange(t *testing.T) {
	s := NewMemStore()
	w := newMemWallet()
	w.seed("user_1", 100000)
	live := func(_ string) (float64, bool) { return 2.10, false }
	b := &Bet{UserID: "user_1", EventID: "SP-2", Odds: 1.85, Stake: 10000}
	if err := Place(s, w, stubRG{max: 50000}, live, b, "req-2"); !errors.Is(err, ErrOddsChanged) {
		t.Fatalf("got %v, want ErrOddsChanged", err)
	}
}

func TestPlaceRejectsOverRGStakeLimit(t *testing.T) {
	s := NewMemStore()
	w := newMemWallet()
	w.seed("user_1", 100000)
	b := &Bet{UserID: "user_1", EventID: "SP-3", Odds: 1.85, Stake: 100000}
	if err := Place(s, w, stubRG{max: 50000}, openMarket, b, "req-3"); !errors.Is(err, ErrStakeTooLarge) {
		t.Fatalf("got %v, want ErrStakeTooLarge", err)
	}
}

func TestSettleWinnerCreditsPayout(t *testing.T) {
	s := NewMemStore()
	w := newMemWallet()
	w.seed("user_1", 100000)
	live := func(_ string) (float64, bool) { return 2.00, false }
	b := &Bet{UserID: "user_1", EventID: "SP-4", Odds: 2.00, Stake: 10000}
	if err := Place(s, w, stubRG{max: 50000}, live, b, "req-4"); err != nil {
		t.Fatal(err)
	}
	if err := Settle(s, w, b.ID, true, "settle-4"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(b.ID)
	if got.Status != SettledW || got.Payout != 20000 {
		t.Fatalf("status=%s payout=%d", got.Status, got.Payout)
	}
	if w.balance("user_1") != 100000-10000+20000 {
		t.Fatalf("balance = %d", w.balance("user_1"))
	}
}

func TestCashoutClosesBet(t *testing.T) {
	s := NewMemStore()
	w := newMemWallet()
	w.seed("user_1", 100000)
	live := func(_ string) (float64, bool) { return 3.00, false }
	b := &Bet{UserID: "user_1", EventID: "SP-5", Odds: 3.00, Stake: 10000}
	if err := Place(s, w, stubRG{max: 50000}, live, b, "req-5"); err != nil {
		t.Fatal(err)
	}
	if err := Cashout(s, w, b.ID, 12000, "cash-5"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(b.ID)
	if got.Status != CashedOut || got.Payout != 12000 {
		t.Fatalf("status=%s payout=%d", got.Status, got.Payout)
	}
}