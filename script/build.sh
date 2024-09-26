#!/usr/bin/env bash

set -euo pipefail

GO_LDFLAGS=' -w -extldflags "-static"'
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

cd "$(dirname "$0")/.."
rm -rf target
mkdir -p target

echo "Building whalebone-clients..."

go build -ldflags "$GO_LDFLAGS" -o "target/whalebone-clients" -buildvcs=false "/go/src/github.com/jamm3e3333/whalebone-clients/cmd"
echo "Built: $(ls target/*)"
