#!/bin/bash
source ./.env

docker compose cp data/whitelist-words.csv db:/tmp/whitelist-words.csv
docker compose cp scripts/sql/import_nickname_words.sql db:/tmp/import_nickname_words.sql

docker compose exec db psql -U postgres -d ${DB_NAME} -f /tmp/import_nickname_words.sql
