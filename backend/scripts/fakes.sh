#!/usr/bin/env bash
set -euo pipefail

# Generate the domain fake builders from the entity declarations beside the types
# Usage: fakes.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"

go run "${PACKAGE_PREFIX}/cmd/tools/codegen/fakes"
