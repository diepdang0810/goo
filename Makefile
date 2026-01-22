.PHONY: run build test up down migrate-create migrate-up run-worker build-worker build-all

run:
	go run cmd/app/main.go

run-worker:
	go run cmd/worker/main.go

build:
	go build -o bin/app cmd/app/main.go

build-worker:
	go build -o bin/worker cmd/worker/main.go

build-all:
	go build -o bin/app cmd/app/main.go
	go build -o bin/worker cmd/worker/main.go

dev:
	$(shell go env GOPATH)/bin/air

dev-worker:
	$(shell go env GOPATH)/bin/air -c .air.worker.toml

test:
	go test -v ./...

up:
	docker-compose up -d

down:
	docker-compose down

migrate-create:
	@read -p "Enter migration name: " name; \
	docker run --rm -v $(shell pwd)/migrations:/migrations migrate/migrate create -ext sql -dir /migrations -seq $$name

migrate-up:
	docker-compose run --rm migrate -path=/migrations/ -database "postgres://postgres:postgres@postgres:5432/go1_db?sslmode=disable" up

migrate-down:
	docker-compose run --rm migrate -path=/migrations/ -database "postgres://postgres:postgres@postgres:5432/go1_db?sslmode=disable" down
