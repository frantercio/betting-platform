# Betting Platform Brasil

![Go](https://img.shields.io/badge/Go-1.23-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Security](https://img.shields.io/badge/security-hardened-red)

Esqueleto funcional de plataforma de apostas regulada (Lei 14.790/2023).

> **Aviso legal**: usar somente com autorização SPA/MF e certificação técnica.
> Docs de compliance em `docs/`.

## 🚀 Deploy rápido

### Railway.app gratuito
```bash
# Conecte GitHub no Railway.app
# Adicione variáveis: JWT_SECRET, ADMIN_PASSWORD
# Deploy automático
```

### Fly.io
```bash
fly auth login
./deploy-fly.sh
```

### Docker Compose
```bash
docker compose -f docker-compose.prod.yml up -d
```

## Estrutura

```
betting-platform/
├── services/
│   ├── wallet/            # Go — ledger dupla-entrada, idempotência, PIX
│   │   └── internal/
│   │       ├── ledger/    # partidas duplas em centavos, MemStore swap p/ Postgres
│   │       └── handler/   # HTTP: balance, deposits, withdrawals, webhook PIX
│   └── user/              # Go — jogo responsável (limites, autoexclusão)
│       └── internal/
│           ├── rg/        # limites 24h, exclusão, reality check
│           └── handler/   # HTTP: /rg/*
├── ml-rg/                 # Python — PoC de score de risco RG (stdlib)
├── infra/                 # (placeholders) Terraform/K8s
├── .github/workflows/     # CI + pipeline de certificação
├── docs/                  # compliance, arquitetura, certificação, openapi
├── openapi-spec.json
├── docker-compose.yml
└── Makefile
```

## Quickstart

```bash
# Dev local (sem docker)
make vet && make test
make run-wallet   # :8080
make run-user     # :8081

# Full stack
docker compose up --build -d
```

## Testes de exemplo

```bash
# Idempotência: mesmo request_id 2x => mesmo saldo (crítico p/ certificação)
(cd services/wallet && go test -run TestDepositIdempotency -v)

# RG: limite para baixo imediato, para cima 24h
(cd services/user && go test -run TestSetDepositLimitUpHas24hDelay -v)

# ML: usuário normal = VERDE; chasing = VERMELHO
(cd ml-rg && python3 risk_score.py && python3 -m unittest -v)
```

## Regras de dinheiro (wallet)

- **Centavos inteiros** (`int64`) — nunca `float` para saldo.
- **Idempotência**: `request_id` único; webhook duplicado não credita 2x.
- **Dupla entrada**: toda transação tem 2 postings que somam zero.
- Em produção: Postgres append-only + Kafka events + PIX via PSP homologado.

## Próximos passos

- [ ] Auth real (JWT + MFA) e compartilhamento entre serviços
- [ ] Store Postgres (migrações SQL em `migrations/`)
- [ ] betting-service + odds + settlements
- [ ] k8s charts + Terraform (região SP, requisito SPA)
- [ ] Integração PSP PIX + Serpro (KYC)