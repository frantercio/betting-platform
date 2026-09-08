package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Ledger struct {
	ID         int64
	RequestID  string
	UserID     string
	TxType     string
	AmountCents int64
	Currency   string
	Status     string
	Postings   []Posting
}

type Posting struct {
	Account string
	Delta   int64
}

type PostgresStore struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &PostgresStore{db: pool}, nil
}

func (s *PostgresStore) Append(ctx context.Context, l *Ledger) error {
	// idempotent upsert on request_id
	postingsJSON, _ := json.Marshal(l.Postings)
	const q = `
		INSERT INTO ledger (request_id, user_id, tx_type, amount_cents, currency, status, postings)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (request_id) DO NOTHING
		RETURNING id
	`
	var id int64
	err := s.db.QueryRow(ctx, q, l.RequestID, l.UserID, l.TxType, l.AmountCents, l.Currency, l.Status, postingsJSON).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// duplicate -> treat as success (idempotent)
			return nil
		}
		return err
	}
	l.ID = id
	// atomic balance adjustments in same tx
	return s.applyPostings(ctx, l.UserID, l.Postings)
}

func (s *PostgresStore) applyPostings(ctx context.Context, userID string, p []Posting) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, po := range p {
		// only main balance for now; schema can be extended later
		if po.Account == fmt.Sprintf("%s:main", userID) {
			_, err = tx.Exec(ctx, `
				UPDATE wallet_accounts SET main_balance_cents = main_balance_cents + $1, version = version + 1
				WHERE user_id = $2
			`, po.Delta, userID)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) Balance(ctx context.Context, userID string) (mainCents int64, err error) {
	const q = `SELECT main_balance_cents FROM wallet_accounts WHERE user_id = $1`
	err = s.db.QueryRow(ctx, q, userID).Scan(&mainCents)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return mainCents, nil
}

func (s *PostgresStore) GetByRequestID(ctx context.Context, requestID string) (*Ledger, error) {
	const q = `SELECT id, request_id, user_id, tx_type, amount_cents, currency, status, postings FROM ledger WHERE request_id = $1`
	var l Ledger
	var postingsJSON []byte
	if err := s.db.QueryRow(ctx, q, requestID).Scan(&l.ID, &l.RequestID, &l.UserID, &l.TxType, &l.AmountCents, &l.Currency, &l.Status, &postingsJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(postingsJSON, &l.Postings)
	return &l, nil
}
