.PHONY: all build test vet run-wallet run-user up down logs clean cert-build

SERVICES = wallet user

# -race is unsupported on android/arm64; enable via RACE=1 (CI runs it).
RACE ?=

all: vet test build

build:
	@for s in $(SERVICES); do \
		echo "==> building $$s"; \
		(cd services/$$s && go build ./...); \
	done

test:
	@for s in $(SERVICES); do \
		echo "==> testing $$s"; \
		(cd services/$$s && go test $(RACE) -cover ./...); \
	done

vet:
	@for s in $(SERVICES); do \
		echo "==> vet $$s"; \
		(cd services/$$s && go vet ./...); \
	done

run-wallet:
	cd services/wallet && go run ./cmd/server

run-user:
	cd services/user && go run ./cmd/server

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=100

clean:
	find . -name '*.test' -delete
	rm -rf services/*/bin

# Freeze a build so the cert lab sees an immutable artifact.
cert-build:
	@echo "=== frozen build ==="
	@git rev-parse HEAD 2>/dev/null || echo "no-git"
	@for s in $(SERVICES); do \
		echo "--> $$s"; \
		(cd services/$$s && go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server); \
	done
	@find services -name server -type f -exec sha256sum {} \;