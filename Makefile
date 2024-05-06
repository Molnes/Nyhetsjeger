build:
	templ generate -path ./internal/web_server/web/
	npx --yes tailwindcss build -i assets/css/styles.css -o assets/css/tailwind.css
	go build -o bin/main cmd/server/main.go

run:
	templ generate -path ./internal/web_server/web/
	npx --yes tailwindcss build -i assets/css/styles.css -o assets/css/tailwind.css
	go run cmd/server/main.go

live-reload:
	air -c .air.toml

#!make
include .env

migrate-up:
	migrate -database $(POSTGRESQL_URL_ROOT) -path db/migrations up

migrate-down:
	migrate -database $(POSTGRESQL_URL_ROOT) -path db/migrations down
	
reset-db:
	./scripts/reset-db.sh
	./scripts/add-db-usr.sh
	go run cmd/db_populator/main.go
	./scripts/add-nickname-words.sh
	
initialize-docker:
	docker compose up -d


run-bruno:
	./scripts/bruno-test.sh

populate-usernames:
	./scripts/add-nickname-words.sh

test-unit:
	go test ./...

test-integration:
	go test --tags=integration ./...