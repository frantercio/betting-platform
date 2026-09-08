// Package bet models a sports bet: placement pre-checks (odds valid, stake
// within risk/RG limits), settlement and cashout.
package bet

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrOddsChanged   = errors.New("odds changed since offer")
	ErrEventSuspended = errors.New("event is suspended")
	ErrStakeTooLarge = errors.New("stake exceeds limit")
	ErrNotOpen       = errors.New("bet is not open")
)

type Status string

const (
	Open      Status = "OPEN"
	SettledW  Status = "SETTLED_WIN"
	SettledL  Status = "SETTLED_LOSS"
	CashedOut Status = "CASHED_OUT"
	Cancelled Status = "CANCELLED"
)

// Bet is one accepted wager (single-leg; combos extend Legs).
type Bet struct {
	ID          int64      `json:"id"`
	UserID      string     `json:"user_id"`
	EventID     string     `json:"event_id"`
	Market      string     `json:"market"`
	Selection   string     `json:"selection"`
	Odds        float64    `json:"odds"`  // accepted odds
	Stake       int64      `json:"stake"` // cents
	Status      Status     `json:"status"`
	PlacedAt    time.Time  `json:"placed_at"`
	SettledAt   *time.Time `json:"settled_at,omitempty"`
	Payout      int64      `json:"payout"` // cents credited on win/cashout
}

// Potential returns the gross payout if the bet wins (stake * odds).
func (b *Bet) Potential() int64 {
	return int64(float64(b.Stake) * b.Odds)
}

// Store persists bets (Postgres in prod).
type Store interface {
	Create(b *Bet) error
	Get(id int64) (*Bet, error)
	Update(b *Bet) error
	List(userID string, status Status, limit int) ([]*Bet, error)
}

// Wallet is the money dependency. Debits stake at placement, credits payout
// at settlement/cashout. Implemented by wallet-service (via ledger).
type Wallet interface {
	Debit(userID string, requestID string, cents int64, ref string) error
	Credit(userID string, requestID string, cents int64, ref string) error
}

// RG checks responsible-gaming gates before accepting a bet.
type RG interface {
	CheckBet(userID string, stake int64) error
}

// Place validates the offered odds against the live market and the user's
// limits, then debits the stake atomically.
func Place(s Store, w Wallet, rg RG, live func(eventID string) (odds float64, suspended bool), b *Bet, requestID string) error {
	if b.Odds < 1.01 {
		return fmt.Errorf("odds must be >= 1.01")
	}
	if b.Stake <= 0 {
		return fmt.Errorf("stake must be positive")
	}
	liveOdds, suspended := live(b.EventID)
	if suspended {
		return ErrEventSuspended
	}
	if b.Odds != liveOdds {
		return ErrOddsChanged
	}
	if err := rg.CheckBet(b.UserID, b.Stake); err != nil {
		return ErrStakeTooLarge
	}
	if err := w.Debit(b.UserID, requestID, b.Stake, "BET_DEBIT:"+b.EventID); err != nil {
		return err
	}
	b.Status = Open
	return s.Create(b)
}

// Settle resolves a bet. winner=true credits the payout.
func Settle(s Store, w Wallet, id int64, winner bool, requestID string) error {
	b, err := s.Get(id)
	if err != nil {
		return err
	}
	if b.Status != Open {
		return ErrNotOpen
	}
	if winner {
		if err := w.Credit(b.UserID, requestID, b.Potential(), fmt.Sprintf("BET_WIN:%d", id)); err != nil {
			return err
		}
		b.Payout = b.Potential()
		b.Status = SettledW
	} else {
		b.Status = SettledL
	}
	now := time.Now()
	b.SettledAt = &now
	return s.Update(b)
}

// Cashout closes an open bet early at the negotiated (discounted) value.
func Cashout(s Store, w Wallet, id int64, offer int64, requestID string) error {
	b, err := s.Get(id)
	if err != nil {
		return err
	}
	if b.Status != Open {
		return ErrNotOpen
	}
	if err := w.Credit(b.UserID, requestID, offer, fmt.Sprintf("BET_CASHOUT:%d", id)); err != nil {
		return err
	}
	b.Payout = offer
	b.Status = CashedOut
	now := time.Now()
	b.SettledAt = &now
	return s.Update(b)
}