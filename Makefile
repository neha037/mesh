-include .env
export

DATABASE_URL ?= postgres://mesh:$(PG_PASSWORD)@localhost:5432/mesh?sslmode=disable

.PHONY: build run-api run-worker test test-integration \
       migrate-up migrate-down \
       docker-up docker-up-ai docker-down docker-logs \
       lint lint-sql sqlc check-sqlc pull-models install uninstall \
       fmt tidy coverage validate-migrations

fmt:
	gofmt -w .

tidy:
	go mod tidy

coverage:
	go test -coverprofile=coverage.out -race ./...
	go tool cover -func=coverage.out

setup:
	@test -f .env || (cp .env.example .env && echo "Created .env from .env.example — edit passwords before use")

build: sqlc
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

test:
	go test ./... -v -race -count=1

test-integration:
	TESTCONTAINERS_RYUK_DISABLED=true go test ./... -v -race -tags=integration -count=1

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

docker-up:
	cd deploy && docker-compose --env-file ../.env up -d

docker-up-ai:
	cd deploy && docker-compose --env-file ../.env --profile ai up -d

docker-down:
	cd deploy && docker-compose --env-file ../.env --profile ai down

docker-logs:
	cd deploy && docker-compose --env-file ../.env logs -f

lint:
	golangci-lint run ./...

lint-sql:
	@command -v squawk >/dev/null 2>&1 || { echo "Install squawk: npm i -g squawk-cli"; exit 1; }
	squawk migrations/*.up.sql

check-sqlc:
	sqlc generate
	@git diff --quiet internal/storage/models.go internal/storage/db.go internal/storage/*.sql.go || \
		(echo "ERROR: sqlc generated code is stale. Run 'make sqlc' and commit the changes."; \
		git diff --stat internal/storage/models.go internal/storage/db.go internal/storage/*.sql.go; exit 1)
	@echo "sqlc code is up-to-date"

validate-migrations: lint-sql check-sqlc test-integration
	@echo "All migration checks passed"

sqlc:
	sqlc generate

pull-models:
	docker exec mesh-ollama ollama pull embeddinggemma:300m-qat-q8_0
	docker exec mesh-ollama ollama pull gemma4:e4b

install:
	bash scripts/install.sh

uninstall:
	systemctl --user stop mesh 2>/dev/null || true
	systemctl --user disable mesh 2>/dev/null || true
	rm -f ~/.config/systemd/user/mesh.service
	rm -f ~/.local/share/applications/mesh.desktop
	rm -f ~/.config/autostart/mesh.desktop
	systemctl --user daemon-reload
	update-desktop-database ~/.local/share/applications 2>/dev/null || true
	@echo "Mesh uninstalled. Docker containers may still be running — use 'make docker-down' to stop them."
