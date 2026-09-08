package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	Pool *pgxpool.Pool
}

func New() *Service {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return &Service{}
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err == nil {
		_ = pool.Ping(ctx)
	}
	return &Service{Pool: pool}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple admin check via header for now
		if r.Header.Get("X-Admin-Key") != os.Getenv("ADMIN_KEY") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status":"ok"})
}

func (s *Service) Users(w http.ResponseWriter, r *http.Request) {
	if s.Pool == nil {
		writeJSON(w, http.StatusOK, map[string]any{"users":[]})
		return
	}
	rows, err := s.Pool.Query(context.Background(), `SELECT id, status, kyc_level, created_at FROM users ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error":err.Error()})
		return
	}
	defer rows.Close()
	var users []map[string]any
	for rows.Next() {
		var id, status string
		var kyc int
		var created string
		_ = rows.Scan(&id, &status, &kyc, &created)
		users = append(users, map[string]any{"id":id,"status":status,"kyc":kyc,"created":created})
	}
	writeJSON(w, http.StatusOK, map[string]any{"users":users})
}

func (s *Service) Wallets(w http.ResponseWriter, r *http.Request) {
	if s.Pool == nil {
		writeJSON(w, http.StatusOK, map[string]any{"wallets":[]})
		return
	}
	rows, _ := s.Pool.Query(context.Background(), `SELECT user_id, main_balance_cents, bonus_balance_cents FROM wallet_accounts LIMIT 100`)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var uid string
		var main, bonus int64
		_ = rows.Scan(&uid, &main, &bonus)
		out = append(out, map[string]any{"user_id":uid,"main":main,"bonus":bonus})
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallets":out})
}

func (s *Service) Bets(w http.ResponseWriter, r *http.Request) {
	if s.Pool == nil {
		writeJSON(w, http.StatusOK, map[string]any{"bets":[]})
		return
	}
	rows, _ := s.Pool.Query(context.Background(), `SELECT id, user_id, status, stake_cents, odds FROM bets ORDER BY id DESC LIMIT 100`)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var uid, status string
		var stake int64
		var odds float64
		_ = rows.Scan(&id, &uid, &status, &stake, &odds)
		out = append(out, map[string]any{"id":id,"user_id":uid,"status":status,"stake":stake,"odds":odds})
	}
	writeJSON(w, http.StatusOK, map[string]any{"bets":out})
}

func (s *Service) BlockUser(w http.ResponseWriter, r *http.Request) {
	// expects ?id=
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"id required"})
		return
	}
	if s.Pool == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status":"ok"})
		return
	}
	_, err := s.Pool.Exec(context.Background(), `UPDATE users SET status='BLOCKED' WHERE id=$1`, id)
	if s.Pool != nil {
		s.Pool.Exec(context.Background(),
			`INSERT INTO admin_audit(actor, action, resource_type, resource_id, details) VALUES($1,$2,$3,$4,$5)`,
			"admin", "BLOCK_USER", "user", id, `{"status":"BLOCKED"}`)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status":"ok", "error": err.Error()})
}

func (s *Service) Audit(w http.ResponseWriter, r *http.Request) {
	if s.Pool == nil {
		writeJSON(w, http.StatusOK, map[string]any{"audit":[]})
		return
	}
	// CSV export if ?format=csv
	if r.URL.Query().Get("format") == "csv" {
		w.Header().Set("Content-Type","text/csv")
		w.Header().Set("Content-Disposition","attachment; filename=audit.csv")
		w.Write([]byte("id,actor,action,resource_type,resource_id,created_at\n"))
		rows, _ := s.Pool.Query(context.Background(),
			`SELECT id, actor, action, resource_type, resource_id, created_at FROM admin_audit ORDER BY id DESC LIMIT 10000`)
		defer rows.Close()
		for rows.Next() {
			var id int64
			var actor, action, rtype, rid, created string
			_ = rows.Scan(&id, &actor, &action, &rtype, &rid, &created)
			fmt.Fprintf(w, "%d,%s,%s,%s,%s,%s\n", id, actor, action, rtype, rid, created)
		}
		return
	}
	rows, _ := s.Pool.Query(context.Background(),
		`SELECT id, actor, action, resource_type, resource_id, details, created_at FROM admin_audit ORDER BY id DESC LIMIT 200`)
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var actor, action, rtype, rid, created string
		var details []byte
		_ = rows.Scan(&id, &actor, &action, &rtype, &rid, &details, &created)
		out = append(out, map[string]any{"id":id,"actor":actor,"action":action,"type":rtype,"resource_id":rid,"details":string(details),"created":created})
	}
	writeJSON(w, http.StatusOK, map[string]any{"audit":out})
}

func (s *Service) AdjustBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	delta := r.URL.Query().Get("delta")
	if userID == "" || delta == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"user_id and delta required"})
		return
	}
	if s.Pool == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status":"ok"})
		return
	}
	_, err := s.Pool.Exec(context.Background(),
		`INSERT INTO wallet_accounts(user_id, main_balance_cents, bonus_balance_cents)
		 VALUES($1,0,0)
		 ON CONFLICT (user_id) DO NOTHING`,
		userID)
	_, err = s.Pool.Exec(context.Background(),
		`UPDATE wallet_accounts SET main_balance_cents = main_balance_cents + $1 WHERE user_id=$2`,
		delta, userID)
	writeJSON(w, http.StatusOK, map[string]string{"status":"ok","error":err.Error()})
}

func (s *Service) Metrics(w http.ResponseWriter, r *http.Request) {
	if s.Pool == nil {
		writeJSON(w, http.StatusOK, map[string]any{"metrics":{}})
		return
	}
	metrics := map[string]any{}
	rows, _ := s.Pool.Query(context.Background(), `SELECT count(*) FROM users`)
	rows.Next(); rows.Scan(&metrics["users"])
	rows, _ = s.Pool.Query(context.Background(), `SELECT count(*) FROM bets`)
	rows.Next(); rows.Scan(&metrics["bets"])
	rows, _ = s.Pool.Query(context.Background(), `SELECT coalesce(sum(main_balance_cents),0) FROM wallet_accounts`)
	rows.Next(); rows.Scan(&metrics["total_balance_cents"])
	// time series last 7 days
	series := []map[string]any{}
	rows, _ = s.Pool.Query(context.Background(), `SELECT date_trunc('day', created_at)::date as d, count(*) FROM users GROUP BY 1 ORDER BY 1 DESC LIMIT 7`)
	for rows.Next() {
		var d string
		var c int
		rows.Scan(&d,&c)
		series = append(series, map[string]any{"date":d,"users":c})
	}
	writeJSON(w, http.StatusOK, map[string]any{"metrics":metrics,"series":series})
}

func (s *Service) Prometheus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","text/plain")
	fmt.Fprint(w,"# HELP admin_users_total Total users\n# TYPE admin_users_total gauge\nadmin_users_total 0\n")
	fmt.Fprint(w,"# HELP admin_bets_total Total bets\n# TYPE admin_bets_total gauge\nadmin_bets_total 0\n")
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"invalid input"})
		return
	}
	if req.Username != "admin" || req.Password != os.Getenv("ADMIN_PASSWORD") {
		if s.Pool != nil {
			s.Pool.Exec(context.Background(),
				`INSERT INTO admin_audit(actor, action, resource_type, resource_id, details) VALUES($1,$2,$3,$4,$5)`,
				req.Username, "LOGIN_FAILED", "admin", req.Username, `{"reason":"invalid_credentials"}`)
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error":"invalid"})
		return
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	mgr := auth.New([]byte(secret), 24*time.Hour)
	token, err := mgr.Sign(r.Context(), "admin")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error":"sign"})
		return
	}
	if s.Pool != nil {
		s.Pool.Exec(context.Background(),
			`INSERT INTO admin_audit(actor, action, resource_type, resource_id, details) VALUES($1,$2,$3,$4,$5)`,
			"admin", "LOGIN_SUCCESS", "admin", "admin", `{}`)
	}
	writeJSON(w, http.StatusOK, map[string]string{"token":token,"role":"admin"})
}
}


