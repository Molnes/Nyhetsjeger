#!/bin/bash
source ./.env

docker run -v ./db/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database $POSTGRESQL_URL_ROOT up