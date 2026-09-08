# Estratégia de Jogo Responsável (RG) - Brasil

> **Base legal**: Lei 14.790/2023 (Art. 29), Portaria SPA/MF nº 722/2024
> **Obrigatório**: 100% implementado antes do go-live
> **Requisitos mínimos SPA**: limites de depósito, limites de tempo, autoexclusão, reality check, histórico completo de atividades.

---

## 1. MODELO DE GOVERNANÇA RG

```
                  ┌─────────────────┐
                  │  Comitê de RG   │  (Chief RG Officer + CTO + Legal + DPO)
                  │  (mensal)       │
                  └────────┬────────┘
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────────┐
│ Time RG      │  │ Data/ML Team │  │ Compliance/AML   │
│ (operações)  │  │ (detecção)   │  │ (listas, COAF)   │
└──────────────┘  └──────────────┘  └──────────────────┘
```

### Responsabilidades
- **Chief RG**: dono da estratégia, reporta ao CEO/Conselho
- **Time RG**: intervenções, reabilitação, canais de ajuda
- **Data/ML**: modelos de risco, alertas, dashboards
- **Compliance**: registro e auditoria de todas as ações RG

---

## 2. FERRAMENTAS OBRIGATÓRIAS (100% ANTES DO GO-LIVE)

### 2.1 Limites por Usuário (configuráveis pelo próprio usuário)

| Limite | Padrão | Variação | Implementação crítica |
|--------|--------|----------|----------------------|
| Depósito diário | R$ 0-250 | R$ 250 a sem limite | **Imediato** após configurar |
| Depósito semanal | R$ 0-1.000 | acima | Imediato |
| Depósito mensal | R$ 0-5.000 | acima | Imediato |
| Perda diária | - | - | Obrigatório alertar em 80% |
| Perda semanal | - | - | Obrigatório alertar em 80% |
| Perda mensal | - | - | Obrigatório alertar em 80% |
| Tempo de sessão | 60 min | +15 min ao confirmar | Popup obrigatório |

> **Regra de ouro**: *Alteração de limite para BAIXO = imediata. Alteração para CIMA = 24h de espera.* (evita impulso)

### 2.2 Autoexclusão

| Tipo | Duração | Efeito |
|------|---------|--------|
| Temporária | 24h, 7d, 30d, 90d, 6m | Bloqueio imediato, irreversível até expirar |
| Permanente | 5+ anos | Bloqueio definitivo, reativação só via suporte com identidade verificada |

- **Durante autoexclusão**: proibido apostar, depositar, receber bônus, fazer login ativo
- **Marketing**: proibido enviar promoções durante exclusão (LGPD + ética)
- **Após expiração**: cooldown de 24h antes do primeiro login ativo

### 2.3 Pause / Cool-off
- Usuário pede pausa de 24h a 30 dias
- Suspensão imediata de apostas/depósitos
- E-mail de confirmação com instruções

### 2.4 Reality Check
- **A cada 60 min** de sessão ativa
- Popup não-blocking do tipo:
  ```
  ⏱️ Você está jogando há 60 minutos.
  Tempo: 1h 00min | Apostado: R$ 250,00 | Resultado: -R$ 80,00
  [Continuar com aviso]  [Encerrar sessão]  [Definir limite]
  ```
- Sem confirmação em 30s → bloqueio automático da sessão
- Registrado em log de auditoria (RG)

### 2.5 Histórico Completo (acesso imediato do usuário)

- Todas as apostas: evento, mercado, odd, valor, resultado
- Todos os depósitos/saques/bônus
- Tempo gasto por sessão + data
- Busca por mês, exportável em CSV/PDF
- **Obrigatório SPA**: tela dedicada "Minha atividade"

---

## 3. MODELO DE DETECÇÃO DE RISCO (ML)

### 3.1 Sinais de risco (features)

```
Grupo 1 - Financeiro:
  - % do depósito mensal gasto em <7 dias
  - frequência de depósitos em horários 00h-06h
  - perseguição de perdas (aumento de stake após perda)
  - velocidade de transações (aposta <5min após depósito)
  - uso de crédito/limite para jogar

Grupo 2 - Comportamental:
  - sessões > 4h/dia
  - > 20 apostas/dia
  - tentativas de reverter autoexclusão
  - alternância de limites para cima em <48h

Grupo 3 - Contextual:
  - idade, histórico de bônus, recorrência de saques
  - reclamações prévias, autoexclusão anterior
  - sinais via chat/suporte
```

### 3.2 Modelo de pontuação

```
Risk Score = f(peso_financeiro, peso_comp, peso_contextual)
Nível 0-30: Verde (monitoramento normal)
Nível 30-60: Amarelo (alertas, revisão manual)
Nível 60+: Vermelho (intervenção obrigatória)
```

### 3.3 Intervenção escalonada

| Nível | Ação | Prazo |
|-------|------|-------|
| Amarelo leve | E-mail educativo + sugestão de limites | 48h |
| Amarelo moderado | Alerta in-app + oferta de falar com RG | 24h |
| Vermelho | Contato humano (telefone/chat dedicado) | **4h** |
| Vermelho crítico | **Recomendação de autoexclusão + redução forçada de limite** | imediato |
| Crítico + recusa | Revisão do time RG + possível bloqueio preventivo | 24h |

> Toda intervenção registrada (quem, quando, como). Se o usuário recusa apoio, documento ETHICAMENTE (LGPD: base legal interesse legítimo + saúde do jogador).

---

## 4. TECNOLOGIA RG (Integração com serviços)

### 4.1 rg-service (microserviço dedicado)

```
Fluxo de aposta com check RG:
  BetRequest
   │
   ├─ 1. rg.checkLimits(userId, stake)
   │       ├─ limite de perda? (soma do período)
   │       └─ limite de depósito? (check no wallet)
   │
   ├─ 2. rg.checkSession(userId)
   │       └─ tempo de sessão > 60min sem reality check?
   │
   ├─ 3. rg.checkExclusion(userId)
   │       └─ em autoexclusão? → bloqueia
   │
   └─ 4. risk.checkBet(userId, stake, event)
           └─ score de risco > 60? → reduz stake máx
```

### 4.2 Tabela de limites (PostgreSQL)

```sql
CREATE TABLE rg_limits (
  user_id uuid REFERENCES users(id),
  limit_type enum('DEPOSIT','LOSS','SESSION_TIME'),
  period     enum('DAILY','WEEKLY','MONTHLY'),
  amount     numeric(18,2),
  effective_from timestamptz NOT NULL,  -- para subida, min 24h
  requested_at timestamptz,             -- quando pediu
  PRIMARY KEY (user_id, limit_type, period)
);

CREATE TABLE rg_exclusions (
  user_id uuid REFERENCES users(id),
  started_at timestamptz,
  duration   enum('24H','7D','30D','90D','6M','PERMANENT'),
  reason     text,
  PRIMARY KEY (user_id, started_at)
);

CREATE TABLE rg_reality_checks (
  id bigserial PRIMARY KEY,
  user_id uuid,
  check_at timestamptz,
  session_seconds int,
  acknowledged bool,
  action_taken enum('CONTINUED','SESSION_ENDED','REALITY_COOLDOWN')
);
```

### 4.3 Eventos (Kafka)

```
rg.limit.set
rg.limit.changed_up            (com 24h delay)
rg.limit.changed_down          (imediato)
rg.limit.reached               (alert)
rg.exclusion.requested
rg.exclusion.activated
rg.exclusion.expired
rg.realitycheck.prompt_sent
rg.realitycheck.acknowledged
rg.risk_score.changed          (por usuário, para ML/alertas)
rg.intervention.taken
```

---

## 5. PORTAL RG PARA O USUÁRIO (UI)

```
Minha Página de Jogo Responsável
├─ Painel de limites (depósito, perda, sessão) - editar
├─ Status de autoexclusão (temporária/permanente) - iniciar
├─ Histórico completo de atividades (apostas, depósitos, saques, tempo)
├─ Reality check current session
├─ Ferramentas de autocontrole: alertas de gasto, notificação de tempo
└─ Links de ajuda: portal 180, CVV, Jogadores Anônimos, atendimento RG
```

---

## 6. TREINAMENTO DE EQUIPE

| Rol | Frequência | Conteúdo |
|-----|-----------|----------|
| Todos os colaboradores | Onboarding + anual | Sinais de problema, como encaminhar |
| Suporte | Trimestral | Scripts de conversa, escalonamento |
| Time RG | Mensal | Casos reais, LGPD, saúde mental |
| Executivo | Anual | Metas de RG, compliance lei |

- Certificação interna de RG para todos
- Paralelo com treino de integridade esportiva (IFMA / Sportradar)

---

## 7. MÉTRICAS RG (Dashboard para SPA)

| Métrica | Meta |
|---------|------|
| % de usuários com limite configurado | > 70% ativos em 6 meses |
| Taxa de autoexclusão ativa | 1-3% dos ativos |
| Tempo médio de resposta a alerta vermelho | < 1h |
| % de jogadores que aceitam intervenção | Reportar |
| GGR de jogadores em risco (comparativo) | Reduzir mês a mês |
| Reclamações RG resolvidas em 5 dias | 100% |

---

## 8. CAMPANHAS PROATIVAS DE EDUCAÇÃO

- Popups mensais de educação (evitar dependência)
- Campanha pós-lossstreak ("se perdeu 3 seguidas, considere parar hoje")
- E-mail com resumo mensal de atividade + dica de autocontrole
- Conteúdo educativo sobre probabilidade/odds (transparência)
- Programa de "vias de retorno": suporte a quem tinha autoexclusão vencida (readmissão gentil)

---

## 9. AUDITORIA RG (para o lab e SPA)

- Logs de: todas as alterações de limite, autoexclusões, reality checks, intervenções
- Registro de quem aprovou cada redução/expansão de limite (admin)
- Relatório mensal RG para o conselho
- **Evidência de implementação**: quantos limites setados, quantas exclusões, tempo médio de resposta da Voz RG
- Auditoria externa: verificar que as ferramentas RG existem e funcionam (não só na UI)

---

## ANEXO: LINHAS DE AJUDA (BRASIL)

- **CVV** (Centro de Valorização da Vida): 188
- **Jogadores Anônimos**: jogadoresanonimos.com (grupos regionais)
- **CAPS** (Centro de Atenção Psicossocial): rede pública SUS
- **Manual MSD / CBASP**: referenciais para avaliação
- **Reino Unido (exemplo)**: GamCare - para benchmarking

---

*Documento vivo — revisar com o time RG e com a lei a cada mudança regulatória (Portaria SPA nº 722/2024 atual).*