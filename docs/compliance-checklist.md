# Checklist de Compliance - Plataforma de Apostas Brasil (Lei 14.790/2023)

> **Status**: Obrigatório antes de qualquer desenvolvimento
> **Responsável**: Legal + Compliance + CTO + CISO

---

## 1. LICENCIAMENTO SPA/MF (CRÍTICO)

### 1.1 Requisitos Societários
- [ ] Sociedade empresária constituída no Brasil (S.A. ou Ltda.)
- [ ] Capital social mínimo: **R$ 30.000.000,00** (apostas esportivas) / R$ 5.000.000,00 (jogos online)
- [ ] Sede e administração no Brasil
- [ ] Representante legal residente no Brasil
- [ ] Composição acionária transparente (beneficiários finais identificados)
- [ ] Ausência de impedidos (políticos, servidores públicos, menores, etc.)

### 1.2 Documentação para Autorização
- [ ] Requerimento assinado por representante legal
- [ ] Contrato social / Estatuto social atualizado
- [ ] Comprovação de capital social integralizado
- [ ] Certidões negativas (federal, estadual, municipal, trabalhista)
- [ ] Comprovação de idoneidade dos sócios/controladores (antecedentes criminais, certidões)
- [ ] Plano de negócios detalhado (5 anos)
- [ ] Demonstrativos financeiros projetados
- [ ] Política de compliance assinada
- [ ] Manual de procedimentos de PLD/FT
- [ ] Política de jogo responsável
- [ ] Política de privacidade (LGPD)
- [ ] Termos e condições de uso
- [ ] Contrato de prestação de serviços com provedor de tecnologia (se houver)

### 1.3 Certificação Técnica (Lab Acreditado)
- [ ] Laboratório escolhido: GLI, BMM, iTech Labs, ou outro acreditado pelo Inmetro
- [ ] Plano de testes submetido e aprovado
- [ ] Ambiente de homologação espelhando produção
- [ ] Testes de: RNG, RTP, segurança, integridade, auditoria, PIX, KYC, jogo responsável
- [ ] Relatório de certificação emitido
- [ ] Renovação anual agendada

---

## 2. SISTEMA DE PAGAMENTOS (PIX OBRIGATÓRIO)

### 2.1 Integração PIX
- [ ] PSP homologado no Banco Central (lista oficial)
- [ ] Chaves PIX exclusivas por usuário (CPF como chave preferencial)
- [ ] QR Code dinâmico por transação
- [ ] Webhook de confirmação em tempo real (< 3s)
- [ ] Conciliação automática diária
- [ ] Estorno/cancelamento conforme normas BCB

### 2.2 Prevenção à Lavagem de Dinheiro (PLD/FT)
- [ ] Cadastro completo (KYC) antes de qualquer depósito
- [ ] Verificação de identidade: documento + selfie + prova de vida (liveness)
- [ ] Consulta a listas restritivas (ONU, OFAC, Bacen, Receita Federal)
- [ ] Monitoramento de transações suspeitas (estruturação, valores atípicos, velocidade)
- [ ] Relatórios de Operações Suspeitas (ROS) ao COAF - prazo 24h
- [ ] Relatório de Operações em Espécie (ROE) se aplicável
- [ ] Relatório de Altas Valor (RAV) > R$ 100.000
- [ ] Guarda de registros por 5 anos (mínimo)
- [ ] Treinamento contínuo da equipe PLD
- [ ] Oficial de PLD nomeado e registrado no COAF

---

## 3. KYC / ONBOARDING (OBRIGATÓRIO ANTES DO 1º DEPÓSITO)

### 3.1 Identificação Pessoa Física
- [ ] CPF válido e regular na Receita Federal
- [ ] Documento oficial com foto (RG, CNH, Passaporte)
- [ ] Selfie com prova de vida (liveness detection certificado)
- [ ] Comprovante de residência (últimos 90 dias)
- [ ] Validação de titularidade da conta bancária/PIX (mesmo CPF)
- [ ] Verificação de idade: **18+** (bloqueio imediato se menor)

### 3.2 Verificações Contínuas
- [ ] Revalidação periódica (12 meses ou gatilhos de risco)
- [ ] Monitoramento de PEP (Pessoas Expostas Politicamente)
- [ ] Screening de mídia adversa
- [ ] Atualização cadastral obrigatória em mudanças

---

## 4. JOGO RESPONSÁVEL (OBRIGATÓRIO LEI 14.790 ART. 29)

### 4.1 Ferramentas Obrigatórias (Todas Implementadas)
- [ ] **Limite de depósito**: diário, semanal, mensal (definido pelo usuário, irrevogável por 24h)
- [ ] **Limite de perda**: diário, semanal, mensal
- [ ] **Limite de tempo de sessão**: alerta + bloqueio automático
- [ ] **Autoexclusão**: temporária (24h a 6 meses) e permanente (mín. 5 anos)
- [ ] **Pause/Cool-off**: 24h a 30 dias
- [ ] **Realidade check**: popup a cada 60 min com tempo gasto e P/L
- [ ] **Histórico completo**: apostas, depósitos, saques, bônus (acesso imediato)

### 4.2 Monitoramento Comportamental
- [ ] Algoritmo de detecção de risco (frequência, valores, horários, perseguição de perdas)
- [ ] Alertas automáticos para equipe de jogo responsável
- [ ] Intervenção proativa: contato, redução de limites, suspensão temporária
- [ ] Encaminhamento para ajuda profissional (CVV, Jogadores Anônimos, CAPS)
- [ ] Dashboard de métricas RG para auditoria SPA

### 4.3 Comunicação
- [ ] Mensagens de jogo responsável em todas as telas
- [ ] Links diretos para ajuda visíveis
- [ ] Campanhas educativas periódicas
- [ ] Proibição de comunicação que incentive jogo excessivo

---

## 5. LGPD / PROTEÇÃO DE DADOS

### 5.1 Base Legal e Governança
- [ ] DPO (Encarregado) nomeado e registrado na ANPD
- [ ] ROPA (Registro de Operações de Tratamento) completo
- [ ] DPIA (Relatório de Impacto) para tratamento de alto risco
- [ ] Base legal mapeada por finalidade (consentimento, contrato, obrigação legal, interesse legítimo)

### 5.2 Direitos dos Titulares (Atendimento em 15 dias)
- [ ] Confirmação e acesso
- [ ] Correção
- [ ] Anonimização/bloqueio/eliminação desnecessários
- [ ] Portabilidade
- [ ] Eliminação (com exceções legais - guarda 5 anos PLD)
- [ ] Informação sobre compartilhamento
- [ ] Revogação do consentimento
- [ ] Oposição a tratamento automatizado

### 5.3 Segurança Técnica
- [ ] Criptografia em trânsito (TLS 1.3) e em repouso (AES-256)
- [ ] Pseudonimização de dados sensíveis em logs/analytics
- [ ] Controle de acesso por menor privilégio (RBAC)
- [ ] Logs de auditoria imutáveis (WORM)
- [ ] Plano de resposta a incidentes (notificação ANPD em 48h)
- [ ] Testes de invasão anuais + após mudanças relevantes

### 5.4 Fornecedores (Data Processing Agreement)
- [ ] DPA assinado com todos subprocessadores
- [ ] Auditoria de segurança de fornecedores críticos
- [ ] Transferência internacional: cláusulas contratuais padrão ou decisão de adequação

---

## 6. SEGURANÇA DA INFORMAÇÃO

### 6.1 Infraestrutura
- [ ] Cloud/hospedagem no Brasil (requisito SPA)
- [ ] Certificação ISO 27001 (desejável) ou controles equivalentes
- [ ] WAF, DDoS protection, rate limiting
- [ ] Segmentação de rede (DMZ, app, data, mgmt)
- [ ] Backup criptografado, testado, geograficamente distribuído
- [ ] RPO < 1h, RTO < 4h

### 6.2 Aplicação
- [ ] OWASP Top 10 coberto (testes SAST/DAST/IAST no pipeline)
- [ ] Autenticação forte: MFA obrigatório (TOTP/WebAuthn)
- [ ] Gestão de segredos (Vault, AWS Secrets Manager, HashiCorp)
- [ ] Rotação de chaves/credenciais automatizada
- [ ] Headers de segurança (CSP, HSTS, X-Frame-Options, etc.)

### 6.3 Operacional
- [ ] SOC 24/7 ou MDR (Monitored Detection & Response)
- [ ] Plano de continuidade de negócios testado
- [ ] Gestão de vulnerabilidades (SLAs por severidade)
- [ ] Treinamento de segurança para todos (onboarding + anual)

---

## 7. AUDITORIA E REPORTING SPA/MF

### 7.1 Relatórios Obrigatórios
- [ ] **Diário**: Volume de apostas, GGR, usuários ativos, transações PIX
- [ ] **Mensal**: Demonstrativo financeiro, tributação, jogo responsável
- [ ] **Trimestral**: Auditoria independente (Big 4 ou equivalente)
- [ ] **Anual**: Relatório de compliance, certificação técnica renovada

### 7.2 Logs de Auditoria (Imutáveis, 5 anos)
- [ ] Todas as apostas (ID, user, evento, odd, valor, resultado, timestamp)
- [ ] Todas as transações financeiras (depósito, saque, bônus, estorno)
- [ ] Todas as ações de KYC/RG (limites alterados, autoexclusão, alertas)
- [ ] Acessos administrativos (quem, o que, quando, IP)
- [ ] Erros e exceções do sistema
- [ ] Assinatura digital / hash encadeado para integridade

### 7.3 Integração SPA (Tecnologia)
- [ ] API de envio de dados em tempo real (padrão SPA)
- [ ] Ambiente de testes SPA homologado
- [ ] Certificado digital ICP-Brasil para assinatura

---

## 8. TRIBUTAÇÃO

### 8.1 Tributos Diretos
- [ ] IRPJ/CSLL (Lucro Real obrigatório)
- [ ] Retenção na fonte (30% prêmios > R$ 2.112 - atualizar anualmente)

### 8.2 Tributos Indiretos
- [ ] ISS (município da sede) - alíquota 2% a 5%
- [ ] PIS/COFINS sobre receita bruta

### 8.3 Obrigações Acessórias
- [ ] ECF, ECD, DCTF, EFD-Contribuições, DIRF
- [ ] e-Fiscal / SPED Fiscal

---

## 9. PUBLICIDADE E MARKETING (CONAR + SPA)

- [ ] Proibido: menores em publicidade, promessa de ganho fácil, "grátis" com depósito
- [ ] Obrigatório: "18+", "Jogue com responsabilidade", link para autoexclusão
- [ ] Proibido patrocínio de eventos majoritariamente infantojuvenis
- [ ] Influenciadores: contrato com cláusulas de compliance, divulgação de risco
- [ ] Bônus: termos claros, rollover justo, não obrigatório para saque de saldo real

---

## 10. SUPORTE E OUVIDORIA

- [ ] Canais: chat 24/7, e-mail, telefone 0800
- [ ] SLA: resposta inicial < 5 min (chat), < 2h (e-mail)
- [ ] Resolução reclamações: 5 dias úteis
- [ ] Ouvidoria independente (terceirizada ou conselho)
- [ ] Registro de todas as interações (5 anos)
- [ ] Integração com consumidor.gov.br e Procon

---

## 11. CHECKLIST DE PRÉ-LANÇAMENTO (GO/NO-GO)

| Item | Status | Responsável | Evidência |
|------|--------|-------------|-----------|
| Autorização SPA/MF publicada no DOU | ☐ | Legal | Diário Oficial |
| Certificação técnica válida | ☐ | CTO/Tech | Relatório Lab |
| PIX homologado em produção | ☐ | Payments | Logs transações teste |
| KYC/AML funcionando end-to-end | ☐ | Compliance | Casos de teste |
| Jogo responsável 100% operacional | ☐ | RG Lead | Demo + testes |
| LGPD/DPIA/ROPA completos | ☐ | DPO | Documentos assinados |
| Penetration test aprovado | ☐ | CISO | Relatório + correções |
| Load test (10x pico esperado) | ☐ | SRE | Relatório + métricas |
| Disaster recovery testado | ☐ | SRE | RTO/RPO comprovados |
| Auditor independete contratado | ☐ | CFO | Contrato assinado |
| Seguro cyber + D&O contratado | ☐ | Risk | Apólices |
| Equipe 24/7 escalonada | ☐ | Ops | Escala + runbooks |

---

## 12. PÓS-LANÇAMENTO (CONTÍNUO)

- [ ] Monitoramento regulatório (mudanças SPA, BCB, ANPD, Receita)
- [ ] Auditoria interna trimestral
- [ ] Penetration test semestral
- [ ] Revisão de risco PLD anual
- [ ] Treinamento contínuo (compliance, segurança, RG)
- [ ] Simulação de incidente (tabletop) trimestral
- [ ] Renovação certificação técnica anual
- [ ] Relatório transparência anual (publicado no site)

---

## ANEXOS REFERÊNCIA

- Lei 14.790/2023 (Apostas de quota fixa)
- Portaria SPA/MF nº 615/2024 (Regulamento técnico)
- Portaria SPA/MF nº 722/2024 (Jogo responsável)
- Portaria SPA/MF nº 1.231/2024 (Certificação)
- Circular BCB nº 3.978/2020 (PLD/FT)
- LGPD (Lei 13.709/2018)
- Resolução CMN 4.658/2018 (Segurança cibernética)
- Normas COAF (PLD/FT)
- Código de Defesa do Consumidor
- Marco Civil da Internet

---

**Assinaturas de Aprovação:**

| Área | Nome | Cargo | Assinatura | Data |
|------|------|-------|------------|------|
| Legal | | Head Legal | | |
| Compliance | | CCO | | |
| Tecnologia | | CTO | | |
| Segurança | | CISO | | |
| Financeiro | | CFO | | |
| Operações | | COO | | |
| CEO | | CEO | | |

---

*Documento vivo - revisar a cada mudança regulatória ou trimestralmente, o que ocorrer primeiro.*