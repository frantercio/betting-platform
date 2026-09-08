module github.com/bets/bepping-platform/services/admin

go 1.27

require (
	github.com/bets/bepping-platform/lib v0.0.0
	github.com/jackc/pgx/v5 v5.7.0
)

replace github.com/bets/bepping-platform/lib => ../../lib
