#!/usr/bin/env bash
set -euo pipefail

# Generate the webhook event catalog from the domain service event type constants
# Usage: webhook_catalog.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"

go run "${PACKAGE_PREFIX}/cmd/tools/codegen/webhook_catalog"
