#!/usr/bin/env bash
set -euo pipefail

# Run admin service with air
# Usage: admin.sh <package_prefix> <artifacts_dir>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"
ARTIFACTS_DIR="${2:-artifacts}"

go tool air \
  --build.cmd "go build -o ./${ARTIFACTS_DIR}/admin_tmp ${PACKAGE_PREFIX}/cmd/services/admin" \
  --build.bin "./${ARTIFACTS_DIR}/admin_tmp" \
  --proxy.app_port "8888" \
  -proxy.proxy_port "9999" \
  -proxy.enabled "true" \
  --build.log "./${ARTIFACTS_DIR}/admin-build-errors.log"
