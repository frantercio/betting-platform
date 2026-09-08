#!/bin/bash
set -e
echo "Deploy Fly.io Betting Platform"

SERVICES=("admin" "wallet" "user" "betting")

for svc in "${SERVICES[@]}"; do
  if [ "$svc" = "admin" ]; then
    cd /data/data/com.termux/files/home/betting-platform
    fly deploy --app betting-admin
  else
    cd /data/data/com.termux/files/home/betting-platform/services/$svc
    fly deploy --app betting-$svc
  fi
done

echo "Deploy completo!"
