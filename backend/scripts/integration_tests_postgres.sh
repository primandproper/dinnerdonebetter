#!/usr/bin/env bash
set -euo pipefail

# Run integration tests for Postgres
# Usage: integration_tests_postgres.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"

go test -v -count=1 "${PACKAGE_PREFIX}/testing/integration/apiserver"
