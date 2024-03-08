#!/bin/bash
source ./.env

docker compose pull
docker compose stop server

docker run -v ./db/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database $POSTGRESQL_URL_ROOT up

docker compose up -d