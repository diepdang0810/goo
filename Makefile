.PHONY: run build test up down migrate-create migrate-up

run:
	go run cmd/app/main.go

build:
	go build -o bin/app cmd/app/main.go

test:
	go test -v ./...

up:
	docker-compose up -d

down:
	docker-compose down

dev:
	air

migrate-create:
	@read -p "Enter migration name: " name; \
	docker run --rm -v $(shell pwd)/migrations:/migrations migrate/migrate create -ext sql -dir /migrations -seq $$name

migrate-up:
	docker run --rm -v $(shell pwd)/migrations:/migrations --network go1_default migrate/migrate -path=/migrations/ -database "postgres://postgres:postgres@go1_postgres:5432/go1_db?sslmode=disable" up

migrate-down:
	docker run --rm -v $(shell pwd)/migrations:/migrations --network go1_default migrate/migrate -path=/migrations/ -database "postgres://postgres:postgres@go1_postgres:5432/go1_db?sslmode=disable" down
