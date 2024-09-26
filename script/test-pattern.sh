#!/usr/bin/env bash

set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 <test_pattern>"
    exit 1
fi

pattern="$1"

docker compose exec whalebone-clients go test -race -v -run "$pattern" ./... | \
sed 's/===\s\+RUN/=== \o033[33mRUN\o033[0m/g' | \
sed 's/===/\o033[36m&\o033[0m/g' | \
sed 's/---/\o033[35m&\o033[0m/g' | \
sed '/PASS:/s/\(PASS:\s[^ ]*\s(\S*)\)/\o033[32m&\o033[0m/' | \
sed '/FAIL:/s/\(FAIL:\s[^ ]*\s(\S*)\)/\o033[31m&\o033[0m/'
