#!/usr/bin/env bash
set -euo pipefail

# Generate the domain converters from the declarations in cmd/tools/codegen/converters
# Usage: converters.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"

go run "${PACKAGE_PREFIX}/cmd/tools/codegen/converters"
