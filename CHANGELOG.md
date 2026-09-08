# Changelog

## 2026-09-07 - Admin completo
- Criado services/admin com Go, Dockerfile, UI HTML
- Login JWT com role admin, guard admin
- Endpoints: /admin/users, /admin/wallets, /admin/bets, /admin/audit, /admin/metrics
- Ações: block user, adjust balance
- Auditoria de login/logout/ações em admin_audit
- Export CSV de auditoria
- Gráfico Chart.js com série temporal 7 dias
- Docker Compose atualizado com serviço admin :8083
- README e migrations automáticas
