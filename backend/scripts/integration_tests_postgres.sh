#!/usr/bin/env bash
set -euo pipefail

# Run integration tests for Postgres
# Usage: integration_tests_postgres.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"

# One package at a time. Each suite stands up its own containers in an init function, and
# two of them racing for host ports and Docker's resources is a startup failure presenting
# as an unrelated test failure.
go test -v -count=1 -p 1 \
	"${PACKAGE_PREFIX}/testing/integration/apiserver" \
	"${PACKAGE_PREFIX}/testing/integration/mcpserver"
