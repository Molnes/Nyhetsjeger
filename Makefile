build:
	templ generate -path ./api/web/
	npx tailwindcss -o assets/css/tailwind.css
	go build -o bin/main cmd/server/main.go

run: reset-db
	templ generate -path ./api/web/
	npx tailwindcss -o assets/css/tailwind.css
	go run cmd/server/main.go

live_reload: reset-db
	air -c .air.toml

#!make
include .env

migrate-up:
	migrate -database $(POSTGRESQL_URL) -path db/migrations up

migrate-down:
	migrate -database $(POSTGRESQL_URL) -path db/migrations down
	

reset-db:
	docker compose down db
	docker volume rm nyhetsjeger-postgres-data
	docker compose up -d db
	migrate -database $(POSTGRESQL_URL) -path db/migrations up # This is broken in make for some reason
	go run cmd/db_populator/main.go