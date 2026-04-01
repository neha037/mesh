.PHONY: build run-api run-worker test test-integration \
       migrate-up migrate-down \
       docker-up docker-up-ai docker-down docker-logs \
       lint sqlc pull-models

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
