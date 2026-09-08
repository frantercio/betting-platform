#!/bin/bash
echo "Teste de carga admin login"
for i in $(seq 1 100); do
  curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8083/admin/login -H 'Content-Type: application/json' -d '{"username":"admin","password":"wrong"}'
done | sort | uniq -c
