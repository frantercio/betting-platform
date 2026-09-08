# Arquitetura de Microserviços - Plataforma de Apostas Brasil

## 1. VISÃO GERAL

```
                                      ┌─────────────────┐
                                      │   CDN / WAF     │
                                      └────────┬────────┘
                                               │
┌───────────────────┐     ┌──────────────────┐▼──────────────────┐
│  Usuário (Web)    │     │  API Gateway     │                   │
│  Usuário (Mobile) │────▶│  (Kong / Envoy)  │  Auth/JWT/MFA     │
│  Admin (Back)     │     │  Rate Limit/LB   │  Observabilidade  │
└───────────────────┘     └───────┬──────────┘                   │
                                  │                              │
                    ┌─────────────┼─────────────────────────────┐
                    │             │                             │
              ┌─────▼─────┐ ┌─────▼─────┐ ┌──────────┐ ┌───────▼──────┐
              │ Core Svc  │ │ User Svc  │ │ Wallet    │ │ Betting Svc  │
              └─────┬─────┘ └─────┬─────┘ │ Svc      │ └───────┬──────┘
                    │             │       └────┬─────┘         │
              ┌─────▼─────┐ ┌─────▼─────┐      │        ┌──────▼──────┐
              │ Casino    │ │ Odds Svc  │      │        │ Risk Svc    │
              │ (3rd)     │ └─────┬─────┘      │        │ Fraud       │
              └─────┬─────┘       │            │        └─────────────┘
                    │             │            │
              ┌─────▼─────┐ ┌─────▼─────┐ ┌────▼─────┐
              │ RG Svc    │ │ KYC/AML   │ │ Reporting│
              │ Limits    │ │ Svc       │ │ Svc      │
              │ Self-excl │ └─────┬─────┘ └────┬─────┘
              └─────┬─────┘       │            │
                    └─────────────┼────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │  Message Bus (Kafka)      │
                    │  Event Streaming          │
                    └───────────────────────────┘
```

---

## 2. TECH STACK RECOMENDADO

### 2.1 Linguagens / Runtimes
| Camada | Tecnologia | Justificativa |
|--------|-----------|---------------|
| Backend (services) | **Go** (alta concorrência, apostas/trading) | Sprint/Runtime baixo, perfeito para sportsbook |
| Backend (negócio) | **Java 21 / Spring Boot** ou **Node.js/NestJS** | Equipe madura com Spring é comum no mercado BR |
| Frontend Web | **React** + TypeScript + Next.js SSR | SEO parcial, cache de odds |
| Mobile | **React Native** ou **Flutter** | Um codebase multiplataforma |
| Admin Panel | React + TypeScript | Painel interno |
| Python | **ml_only** (fraude, RG, recomendações) | Serviços de ML isolados |

> **Recomendação**: Go para os 3 serviços mais críticos (Wallet, Betting, Risk) e Java/Node para o restante. Times menores: tudo Node/TypeScript ou tudo Go.

### 2.2 Dados
| Tipo | Tecnologia |
|------|-----------|
| Banco relacional principal | **PostgreSQL 16+** (wallet, user, bet core) |
| Cache | **Redis** (sessões, odds recentes, rate limit) |
| Streaming/Eventos | **Apache Kafka** (event sourcing de bets, transações) |
| Busca | **OpenSearch / Elasticsearch** (histórico, relatórios) |
| Analytics | **ClickHouse** (reporting SPA, clickstream) |
| ML Features | S3 + Feast (feature store) |

### 2.3 Infraestrutura
| Item | Tecnologia |
|------|-----------|
| Orquestração | **Kubernetes (EKS/GKE/AKS)** ou **Docker Compose** para MVP |
| IaC | Terraform |
| CI/CD | GitHub Actions ou GitLab CI |
| Observabilidade | Prometheus + Grafana + Loki + Tempo + OpenTelemetry |
| Alerting | Alertmanager + PagerDuty/OnCall |
| Segredos | HashiCorp Vault ou AWS Secrets Manager |
| Service Mesh (opcional grande escala) | Istio/Linkerd |
| Cloud | AWS/Google Cloud com **região no Brasil** (São Paulo, requisito SPA) |

---

## 3. CATÁLOGO DE SERVIÇOS

### 3.1 Core
| Serviço | Idioma | DB | Descrição |
|---------|--------|-----|-----------|
| **api-gateway** | Go/Nginx | - | Roteamento, auth, rate limit, mTLS |
| **user-service** | Java/Node | PostgreSQL | Cadastro, perfil, sessões, MFA |
| **wallet-service** | Go | PostgreSQL + Kafka | Saldo, transações, PIX, bônus. **Single source of truth do dinheiro** |
| **betting-service** | Go | PostgreSQL + Kafka | Aceitação de apostas, resolução, cashout |
| **odds-service** | Java/Node | Redis + PG | Ingestion de odds, mercado, limite por odd |
| **casino-gateway** | Java/Node | Redis | Proxy para provedores terceiros de cassino |
| **game-service** | Java/Node | PG | Catálogo, provedores, spin wagers |
| **promo-service** | Java/Node | PG + Redis | Bônus, freebets, fidelidade, rollover |
| **kyc-service** | Python/Node | PG + S3 | Upload docs, validação CPF/Serpro, liveness, screening |
| **risk-service** | Go | Redis + PG + ML | Fraude, apostas suspeitas, limites dinâmicos |
| **aml-service** | Python | PG + Elastic | PLD/FT, listas restritivas, ROS/COAF |
| **rg-service** | Java/Node | PG | Jogo responsável: limites, autoexclusão, realty checks |
| **reporting-service** | Go | ClickHouse | Relatórios SPA, financeiros, auditoria |
| **audit-service** | Go | S3 (WORM) | Logs imutáveis com hash encadeado |
| **notification-service** | Node | Kafka + Redis | Email/SMS/Push/WhatsApp |
| **cms-service** | Node | PG | Conteúdo, promo pages, i18n |
| **admin-service** | Java/Node | PG | Backoffice, RBAC, gestão de usuários |
| **support-service** | Node | PG | Tickets, chat, ouvidoria, consumidor.gov |

### 3.2 Serviços transversais
| Serviço | Descrição |
|---------|-----------|
| **auth-service** | Emissão/validação JWT + refresh, OAuth2, MFA |
| **event-bus** | Kafka topologias (bets, wallet, users, risk) |
| **feature-flag** | Unleash / LaunchDarkly |
| **secrets** | Vault |
| **schema-registry** | Confluent / Karapace |

---

## 4. PADRÕES CRÍTICOS

### 4.1 Wallet Service (Proteção de Dinheiro)
- **Idempotência obrigatória** em cada transação (request_id único)
- **Event sourcing**: cada transação vira evento imutável no Kafka
- **Ledger**: debit/credit em partidas duplas (contas principais / sub-contas)
- **Concorrência**: transações atômicas no PostgreSQL, `SELECT ... FOR UPDATE` + otimistic locking
- **Nunca** usar saldo em memória; saldo = soma do ledger
- **Conciliação** diária com PSP/PIX
- Transações: DEPOSIT, WITHDRAW, BET_DEBIT, BET_CREDIT, CASHOUT, BONUS, ADJUST, REVERSAL

### 4.2 Betting Service
- **Odds com 3 estados**: OPEN → SUSPENDED → CLOSED
- **Aceitação**: validar odd vigente + limite do usuário (risk) antes de aceitar
- **Resolução**: job idempotente, publica evento `BET_SETTLED`
- **Cashout**: valor negociado, transação atômica wallet
- **Dedupe**: mesmo bet_id não pode ser processado 2x

### 4.3 Event Sourcing (Kafka Toipcs)
```
bets.created
bets.accepted
bets.settled.win
bets.settled.loss
bets.cashout
wallet.deposit.completed
wallet.withdrawal.requested
wallet.withdrawal.completed
wallet.bet.debited
wallet.bet.credited
user.registered
user.kyc.approved
rg.limit.changed
rg.selfexcluded
risk.alert.raised
```

---

## 5. SEGURANÇA TRANSVERSAL

- **TLS 1.3** em todos os serviços
- mTLS entre serviços (Service Mesh ou Vault PKI)
- **JWT** curto (15 min) + refresh token rotativo + MFA (TOTP/WebAuthn)
- **RBAC** centralizado (admin-service), escopos OAuth2 por serviço
- Rate limiting por IP, por usuário, por endpoint (gateway + Redis)
- Nenhum serviço expõe dados do ledger na rede pública
- **Segregação**: wallet-service acessa o DB financeiro exclusivamente (dono do schema)

---

## 6. MVP vs PRODUÇÃO

### 6.1 MVP (para validar + certificação)
- Monolito modular (Spring Boot / NestJS) **OU** 6 serviços: gateway, user, wallet, betting, kyc, rg
- 1 DB PostgreSQL + Redis + Kafka
- Docker Compose (dev) + ECS (prod)
- Provedores de cassino integrados via API

### 6.2 Produção (escala + 500k usuários)
- Full microservices conforme catálogo acima
- Kubernetes multi-AZ em São Paulo
- ClickHouse para reporting
- Serviços de ML (fraude, RG)
- Service Mesh + observabilidade completa

**Regra**: Não comece com 20 microserviços. Comece enxuto, evolua por necessidade e por limites de time/ownership. A certificação SPA exige segurança, não microserviços.

---

## 7. DIAGRAMA DE FLUXO PIX (KA-CHING! Transação crítica)

```
User clica "Depositar"
    │
    ▼
wallet-service.initializeDeposit(userId, amount, pixKey)
    │
    ├─ Cria transação PENDING (idempotente)
    ├─ Gera QR Code dinâmico (via PSP)
    └─ Publica wallet.deposit.requested
    │
    ▼
PSP envia webhook: "PIX CREDITADO" (confirmado ou TID)
    │
    ▼
wallet-service.validateWebhook (assinatura PSP, dedupe TID)
    │
    ├─ Marca transação COMPLETED
    ├─ Credita saldo (event sourcing)
    ├─ Notifica user (PUSH/WhatsApp)
    └─ Publica wallet.deposit.completed
    │
    ▼
User pode apostar (betting checa saldo no wallet)
```

---

## 8. DESENHO DE DADOS PRINCIPAIS

### 8.1 usuarios
```sql
users(id uuid PK, cpf_hash, email, phone, name, birth_date,
      status enum, kyc_level int, mfa_enabled bool, created_at)

user_wallets(user_id FK, balance numeric(18,2), bonus numeric(18,2),
             version int)  -- otimistic lock
ledger(id bigserial PK, wallet_id FK, txn_type, amount, currency,
       request_id UNIQUE, status, created_at)  -- append-only
```

### 8.2 apostas
```sql
bets(id uuid PK, user_id FK, event_id, market, selection, odds,
     stake numeric, potential numeric, status enum,
     placed_at, settled_at)
bet_events(bet_id FK, type, payload jsonb, created_at)  -- dreamhe
```

---

## 9. PLANO DE MIGRAÇÃO PARA PRODUÇÃO

1. **Fase 0 (2-3 semanas)**: Setup infra (Terraform, K8s, Vault), CI/CD, observabilidade
2. **Fase 1 (4-6 semanas)**: Serviços core (gateway, user, wallet, betting) + PIX
3. **Fase 2 (4 semanas)**: KYC/AML, jogo responsável
4. **Fase 3 (4 semanas)**: Cassino/provedores, promoções
5. **Fase 4 (2 semanas)**: Reporting SPA, auditoria, logs WORM
6. **Fase 5 (2 semanas)**: Hardening, pen-test, load test, DR test
7. **Go-Live** após aprovação jogos responsáveis + certificação válida

---

## 10. RESPONSABILIDADES POR SERVIÇO (OWNERSHIP)

| Serviço | Dono | SLA uptime |
|---------|------|-----------|
| wallet, betting, ledger | Core Payments | 99.99% |
| api-gateway | Platform | 99.99% |
| user, kyc, rg | Compliance/Trust | 99.9% |
| odds, casino, game | Product | 99.9% |
| reporting, audit | Data/Fin | 99.5% |

---

*Arquitetura viva — revisar a cada release de produção e a cada mudança regulatória.*