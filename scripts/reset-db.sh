#!/bin/bash
source ./.env
POSTGRESQL_URL=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable

docker compose down db
docker volume rm nyhetsjeger-postgres-data
docker compose up -d db

# wait for db to start accepting connections
# code taken from https://github.com/golang-migrate/migrate/issues/366#issuecomment-1288221302
for i in {1..10}; do docker compose exec db pg_isready && break || sleep 1; done
sleep 1;

echo "Running migrate up"
migrate -database ${POSTGRESQL_URL} -path db/migrations -verbose up

go run cmd/db_populator/main.go