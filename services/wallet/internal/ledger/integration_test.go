package ledger_test

import (
	"context"
	"testing"

	"github.com/bets/bepping-platform/services/wallet/internal/ledger"
)

func TestPostgresAppendIdempotent(t *testing.T) {
	// TODO: start postgres via testcontainers, run migrations, then test Append twice with same request_id
	// This stub shows intent for CI integration tests.
	t.Skip("integration test requires Docker")
	_ = ledger.NewPostgres(context.Background(), "postgres://user:pass@localhost:5432/db")
}
