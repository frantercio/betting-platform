#!/bin/bash
# Inicia Docker daemon
dockerd > /var/log/dockerd.log 2>&1 &
sleep 5

cd /data/data/com.termux/files/home/betting-platform

# Subir stack
docker compose up --build -d

echo "Serviços iniciados"
docker ps
