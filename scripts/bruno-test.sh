#!/bin/bash

# & to run in background
go run cmd/test_users/main.go &

cd bruno
npx @usebruno/cli run --env test


curl -X POST http://localhost:8089/shutdown