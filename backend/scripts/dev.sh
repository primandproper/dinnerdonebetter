#!/usr/bin/env bash
set -euo pipefail

# Run local dev server
# Usage: dev.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"

go run "${PACKAGE_PREFIX}/cmd/localdev/server"
