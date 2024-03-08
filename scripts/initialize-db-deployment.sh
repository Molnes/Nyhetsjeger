#!/bin/bash

docker compose up -d db
./scripts/add-db-usr.sh