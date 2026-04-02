-include .env
export

DATABASE_URL ?= postgres://mesh:$(PG_PASSWORD)@localhost:5432/mesh?sslmode=disable

.PHONY: build run-api run-worker test test-integration \
       migrate-up migrate-down \
       docker-up docker-up-ai docker-down docker-logs \
       lint sqlc pull-models install uninstall

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

test:
	go test ./... -v -race -count=1

test-integration:
	go test ./... -v -race -tags=integration

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

sqlc:
	sqlc generate

pull-models:
	docker exec mesh-ollama ollama pull nomic-embed-text
	docker exec mesh-ollama ollama pull mistral:7b-instruct-q4_0

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
