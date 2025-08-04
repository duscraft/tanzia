#!/bin/bash

set -euo pipefail

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.3.1
chmod a+x ./bin/golangci-lint
./bin/golangci-lint run ./apps/go/...
