# Betting Platform Admin

## Acesso
- URL: http://localhost:8083
- Login: `admin` / senha definida em `ADMIN_PASSWORD`
- JWT secret: `JWT_SECRET`

## Docker Compose
```bash
docker compose up --build
```

Serviços:
- Postgres :5432
- Wallet   :8080
- User     :8081
- Betting  :8082
- Admin    :8083

## Endpoints Admin
- `POST /admin/login` → `{username,password}`
- `GET /admin/users` → lista usuários
- `POST /admin/users/block?id=...` → bloqueia usuário
- `GET /admin/audit` → auditoria
- `GET /admin/audit?format=csv` → export CSV
- `GET /admin/metrics` → métricas + série 7 dias

## Migrations
Banco inicializa automaticamente com `db/init/*.sql` via Docker.

## Variáveis
- `DATABASE_URL`
- `JWT_SECRET`
- `ADMIN_PASSWORD`
