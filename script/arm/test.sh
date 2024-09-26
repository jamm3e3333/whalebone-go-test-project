#!/usr/bin/env bash

set -euo pipefail

docker compose exec whalebone-clients go test -race -v -timeout 30s -coverpkg=./... -coverprofile=./tmp/coverage ./... | \
sed -E 's/===\s+RUN/=== \x1B[33mRUN\x1B[0m/g' | \
sed -E 's/===/\x1B[36m&\x1B[0m/g' | \
sed -E 's/---/\x1B[35m&\x1B[0m/g' | \
sed -E $'/PASS:/s/(PASS:\\s[^ ]*\\s(\\S*))/\x1B[32m&\x1B[0m/' | \
sed -E $'/FAIL:/s/(FAIL:\\s[^ ]*\\s(\\S*))/\x1B[31m&\x1B[0m/'

docker compose exec whalebone-clients go tool cover -html=./tmp/coverage -o ./tmp/coverage.html
