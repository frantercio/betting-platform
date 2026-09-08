# Checklist Segurança Produção

## Infra
- [ ] TLS obrigatório via Let's Encrypt
- [ ] WAF Cloudflare/AWS Shield
- [ ] Secrets em Vault, nunca env
- [ ] Rede privada VPC, serviços só via internal DNS

## Auth
- [ ] JWT secret 32+ bytes rotativo
- [ ] Access token 15min, refresh 7 dias
- [ ] Rate limit login 5/min IP
- [ ] MFA obrigatório para admin

## App
- [ ] Headers de segurança ativos
- [ ] Input validation em todos handlers
- [ ] SQL queries parametrizadas
- [ ] Logs centralizados Loki + alertas

## Monitoramento
- [ ] Prometheus + Grafana
- [ ] Alerta login falhas >10/min
- [ ] Alertas saldo anormal / apostas suspeitas
- [ ] Audit trail imutável

## Compliance
- [ ] LGPD / Lei 14.790
- [ ] Retenção logs 5 anos
- [ ] Testes de penetração trimestrais
