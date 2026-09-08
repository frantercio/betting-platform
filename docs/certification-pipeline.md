# Pipeline de Certificação - Laboratórios Acreditados (GLI/BMM/iTech)

## OBJETIVO

Garantir que a plataforma confira nos testes do laboratório acreditado e obtenha o **Certificado de Sistema de Jogos** exigido pela Portaria SPA/MF para operar legalmente no Brasil. O checklist de conformidade técnica que demonstra integridade, aleatoriedade, segurança e conformidade com jogo responsável.

---

## 1. O QUE O LABORATÓRIO TESTA (ESCopo SPA)

Os laboratórios acreditados (GLI-19, BMM, iTech Labs, etc.) auditam:

| Área | Teste | O que precisa estar pronto |
|------|-------|---------------------------|
| **RNG** | Testes estatísticos de aleatoriedade | Gerador certificado, seed, documentação |
| **RTP** | Percentual de retorno ao jogador | Relatório calculado por jogo |
| **Integridade** | Apostas não podem ser manipuladas | Event sourcing, logs imutáveis |
| **Segurança** | PCI, LGPD, criptografia | Wallet, KYC, TLS, pen-test |
| **PIX** | Conciliação e confirmação | Webhooks, idempotência |
| **Jogo Responsável** | Limites, autoexclusão, reality check | Todos com registros auditáveis |
| **Fiscal** | Reporte SPA | Dados estruturados enviados em tempo real |

---

## 2. ARQUITETURA DE LOGS PARA AUDITORIA (WORM)

O laboratório vai exigir prova de que **todas as ações são registradas de forma imutável por 5 anos**.

### 2.1 Log Estruturado (formato JSON - cada serviço)

```json
{
  "version": "1.1",
  "timestamp": "2026-09-07T14:30:00Z",
  "service": "betting-service",
  "event": "BET_PLACED",
  "request_id": "a1b2c3d4-...",
  "user_id": "u_0001",
  "payload": {
    "bet_id": "b_0001",
    "event_id": "SP-123456",
    "stake": 100.00,
    "odds": 1.85,
    "ip": "200.1.2.3",
    "device": "mobile",
    "session_id": "s_0001"
  },
  "hash_prev": "SHA256(previous_log)",
  "hash_own": "SHA256(canonical(this))"
}
```

### 2.2 Cadeia de Hash (Blockchain de logs)

```
log[0].hash = H(log[0])
log[1].hash = H(log[0].hash + log[1])
log[n].hash = H(log[n-1].hash + log[n])
```

- Cada novo log referencia o hash do anterior → **imutável**
- Ancoragem retroativa: hash periódico enviado a entidade externa (ex.: hash no blockchain público, verificação timestamp)
- Backend: S3 com Object Lock (WORM) ou banco com trigger de append-only

### 2.3 Eventos que DEVEM ser logados

```
Todos os GAMBLING events: aposta, resultado, cashout, bônus
Todas as TRANSACTIONS: depósito, saque, ajuste, reversão
Todas as ações KYC: submit, approve, reject, change_level
Todas as ações RG: limit_set, limit_changed, selfexclude, realitycheck
Todos os acessos ADMIN: login, view, edit, export
Erros e exceções do sistema
```

---

## 3. TESTES AUTOMATIZADOS (PRÉ-REQUISITO PARA O LAB)

O lab NÃO vai redesenhar seus testes. Eles querem ver **prova** de que você testou. Monte suíte robusta.

### 3.1 Camadas de teste

| Camada | Ferramenta sugerida | O que testa |
|--------|--------------------|-------------|
| Unit | Jest (JS/TS) / Go test / JUnit | Lógica pura, cálculos de odds, ledger |
| Integration | Testcontainers + SQLite/PG | Fluxos entre serviços |
| Contract | Pact / Schemathesis | Compatibilidade OpenAPI |
| E2E (UI) | Playwright / Cypress | Aposta até settlement, depósito, saque |
| API | Postman/Newman ou Karate | Sequências completas |
| Concurrency | k6 / Gatling | Idempotência, race conditions no wallet |

### 3.2 Casos de teste obrigatórios (direcionados ao lab)

```
1. Depósito PIX → confirmação webhook → saldo atualizado (com idempotência: duplicação de webhook NÃO credita 2x)
2. Aposta → odd estática → aceita; odd muda → rejeita com ODDS_CHANGED
3. Aposta com stake > limite RG → rejeitada
4. Usuário em autoexclusão → qualquer transação bloqueada
5. Saque: fraude check aprovado/revisão/rejeitado
6. Reality check: após 60 min, popup obrigatório; sem confirmação, bloqueio
7. Menor de idade (17) → bloqueio de cadastro
8. Limite de depósito diário: após atingir, 429
9. Cashout: valor correto creditado, bet fechada
10. Reversão: transação errônea estornada sem afetar saldo total
11. Integridade: mesmo request_id 2x → 1 transação
12. RNG: 1M+ tiradas com chi-square, não rejeitado
```

### 3.3 Exemplo: Teste de idempotência do wallet (Go)

```go
func TestDepositIdempotency(t *testing.T) {
    reqID := uuid.New().String()
    // envia 2x o mesmo request_id
    r1 := deposit(100.00, reqID)
    r2 := deposit(100.00, reqID)
    if r1.txID != r2.txID {
        t.Fatal("duplicate request_id must return same transaction")
    }
    balance := getBalance()
    if balance != 100.00 {
        t.Fatalf("duplicate deposit credited twice: got %f", balance)
    }
}
```

---

## 4. CI/CD (GitHub Actions / GitLab)

### 4.1 Pipeline por PR (rápido, ~10 min)
```
checkout
├─ lint (golangci-lint / eslint)
├─ unit tests (coverage > 80%)
├─ contract tests
├─ build image
└─ trivy scan (SAST + vulnerabilities)
```

### 4.2 Pipeline de staging (por merge na main)
```
checkout
├─ lint + unit
├─ integration (Testcontainers)
├─ E2E (Playwright - staging)
├─ concurrency (k6)
├─ report de cobertura
└─ deploy staging (helm)
```

### 4.3 Pipeline de certificação (manual, gate)
```
[GATE: requires approval de CTO + Compliance]
├─ full E2E suite em ambiente de homologação espelhado
├─ coleta de logs de auditoria (export para lab)
├─ gera relatório de evidências (PDF assinado digitalmente)
├─ empacota certificado + relatório RNG
└─ arquiva no S3 WORM
```

> **Ambiente de homologação** deve ser **idêntico** a produção (mesma imagem, mesmo helm chart), apenas com dados fictícios. O lab quer ver o que vai rodar de verdade.

---

## 5. EVIDÊNCIAS PARA O LAB (Checklist de entrega)

- [ ] Documento de arquitetura (aprovado por CTO)
- [ ] Build ID / hash da imagem certificada
- [ ] Relatório de testes de RNG (chi-square, KS, runs) - 1M+ amostras
- [ ] Relatório de RTP por jogo
- [ ] Dump de logs de auditoria de um dia de teste completo
- [ ] Resultado de pen-test (aprovado, sem alta criticidade)
- [ ] Certificados: TLS, ICP-Brasil, assinatura digital
- [ ] Registro de versão definitiva (imutável)
- [ ] Manifesto de compliance assinado por CTO + CCO

---

## 6. INTEGRAÇÃO CONTÍNUA COM O LAB (Processo)

```
Fase 1 - Onboarding
  │  Escolher lab, assinar NDA, receber plano de testes
  ▼
Fase 2 - Preparação (6-8 semanas)
  │  Preparar ambiente de homologação, suíte de testes, evidências
  ▼
Fase 3 - Submissão
  │  Enviar build frozen (hash) + relatórios + acesso ao ambiente
  ▼
Fase 4 - Testes do lab (2-4 semanas)
  │  Lab executa seus procedimentos, solicita ajustes
  ▼
Fase 5 - Correções/Iteração
  │  Ajustar, re-submit, até passar
  ▼
Fase 6 - Certificação emitida (válida 12 meses)
  ▼
Fase 7 - Conformidade contínua
     Renovação anual + re-test quando houver mudança significativa
```

---

## 7. FERRAMENTAS SUGERIDAS

| Uso | Ferramenta |
|-----|-----------|
| E2E | Playwright (web), Detox (mobile) |
| API/Contract | Karate, Pact, Schemathesis |
| Load/Concurrency | k6 + Grafana dashboard |
| SAST | Trivy, Semgrep, SonarQube |
| DAST | OWASP ZAP (em staging) |
| Logs WORM | S3 Object Lock + hash chain (custom ou OpenSearch) |
| Metricas RNG | Dieharder / NIST SP 800-22 (Python/Go) |
| CI | GitHub Actions / GitLab CI |
| Observabilidade | OpenTelemetry → Prometheus + Loki + Grafana |

---

## 8. MÉTRICAS DE SAÚDE PARA A AUDITORIA

- **Cobertura de testes** > 80% nos serviços core (wallet, betting, RG)
- **0 falhas críticas** em pen-test
- **Idempotência**: 100% das transações com request_id único rejeitadas/neutralizadas
- **Liveness**: 99.9% uptime em produção (SLA)
- **Logs**: 100% dos eventos de auditoria presentes na cadeia de hash
- **RG Obrigatório**: 100% de usuários com limites definidos antes do 1º depósito

---

## 9. RESPONSÁVEL / OWNERSHIP

| Papel | Responsável |
|-------|-------------|
| Owner técnico da certificação | Tech Lead / DevSecOps |
| Validação de compliance | CCO / Compliance Officer |
| Aprovação de release | CTO |
| Coordenação com lab | PMO / Delivery Manager |
| Documentação de evidências | QA Lead + Documentação |

---

## 10. IMPLEMENTAÇÃO GENÉRICA (PRÓXIMO CÓDIGO)

Quando a arquitetura de deploy for definida, criar no repositório:

- `docker-compose.yml` (dev + homologação)
- `helm/` charts
- `.github/workflows/certification.yml`
- `scripts/rng_test.go` (coleta de amostras RNG)
- `scripts/build_hash.sh` (frozen build ID)
- Documentação de evidências em `docs/certification/`

---

*Documento vivo — sincronizar com a equipe de lab e com as portarias SPA vigentes.*