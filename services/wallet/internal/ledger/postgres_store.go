package ledger

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (p *PostgresStore) WithAccount(userID string, fn func(*Account) error) error {
	// Simplified: lock row, fetch, modify, save
	// In production use SELECT FOR UPDATE
	var main, bonus int64
	err := p.pool.QueryRow(context.Background(),
		`SELECT main_balance_cents, bonus_balance_cents FROM wallet_accounts WHERE user_id=$1`,
		userID,
	).Scan(&main, &bonus)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	acc := &Account{UserID: userID, MainCents: main, BonusCents: bonus}
	if err := fn(acc); err != nil {
		return err
	}
	_, err = p.pool.Exec(context.Background(),
		`INSERT INTO wallet_accounts(user_id, main_balance_cents, bonus_balance_cents)
		 VALUES($1,$2,$3)
		 ON CONFLICT (user_id) DO UPDATE SET main_balance_cents=$2, bonus_balance_cents=$3`,
		userID, acc.MainCents, acc.BonusCents)
	return err
}

func (p *PostgresStore) Append(tx *Transaction) error {
	// Idempotent upsert on request_id
	postingsJSON, _ := json.Marshal(tx.Postings)
	_, err := p.pool.Exec(context.Background(),
		`INSERT INTO ledger(request_id, user_id, tx_type, amount_cents, currency, status, postings)
		 VALUES($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (request_id) DO NOTHING`,
		tx.RequestID, tx.UserID, string(tx.Type), tx.Amount, tx.Currency, tx.Status, postingsJSON)
	if err != nil {
		return err
	}
	// Update balances atomically per posting
	for _, posting := range tx.Postings {
		if posting.Account == tx.UserID+":main" {
			_, err = p.pool.Exec(context.Background(),
				`INSERT INTO wallet_accounts(user_id, main_balance_cents, bonus_balance_cents)
				 VALUES($1,$2,0)
				 ON CONFLICT (user_id) DO UPDATE SET main_balance_cents = wallet_accounts.main_balance_cents + $2`,
				tx.UserID, posting.Delta)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *PostgresStore) GetByRequestID(requestID string) (*Transaction, error) {
	var tx Transaction
	var postingsJSON []byte
	err := p.pool.QueryRow(context.Background(),
		`SELECT id, request_id, user_id, tx_type, amount_cents, currency, status, postings
		 FROM ledger WHERE request_id=$1`, requestID).
		Scan(&tx.ID, &tx.RequestID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Currency, &tx.Status, &postingsJSON)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(postingsJSON, &tx.Postings)
	return &tx, nil
}

func (p *PostgresStore) List(userID string, limit int) ([]*Transaction, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT id, request_id, user_id, tx_type, amount_cents, currency, status, postings
		 FROM ledger WHERE user_id=$1 ORDER BY id DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Transaction
	for rows.Next() {
		var tx Transaction
		var postingsJSON []byte
		if err := rows.Scan(&tx.ID, &tx.RequestID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Currency, &tx.Status, &postingsJSON); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(postingsJSON, &tx.Postings)
		out = append(out, &tx)
	}
	return out, nil
}

func (p *PostgresStore) Balance(userID string) (main, bonus int64, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT COALESCE(main_balance_cents,0), COALESCE(bonus_balance_cents,0) FROM wallet_accounts WHERE user_id=$1`,
		userID).Scan(&main, &bonus)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, nil
	}
	return main, bonus, err
}
