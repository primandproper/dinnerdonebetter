#!/usr/bin/env bash
set -euo pipefail

# Generate queries
# Usage: queries.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"

go run "${PACKAGE_PREFIX}/cmd/tools/codegen/queries"
