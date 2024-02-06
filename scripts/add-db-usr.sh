#!/bin/bash
source ./.env

echo "CREATE USER ${DB_USR_APP} WITH PASSWORD '${DB_PASSWORD_APP}';" | docker compose exec -T db psql -U postgres -d ${DB_NAME}
echo "GRANT SELECT, UPDATE, INSERT, DELETE ON ALL TABLES IN SCHEMA public TO ${DB_USR_APP};" | docker compose exec -T db psql -U postgres -d ${DB_NAME}
echo "GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO ${DB_USR_APP};" | docker compose exec -T db psql -U postgres -d ${DB_NAME}