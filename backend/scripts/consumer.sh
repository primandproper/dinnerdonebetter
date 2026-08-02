#!/usr/bin/env bash
set -euo pipefail

# Run consumer service with air
# Usage: consumer.sh <package_prefix> <artifacts_dir>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"
ARTIFACTS_DIR="${2:-artifacts}"

go tool air \
  --build.cmd "go build -o ./${ARTIFACTS_DIR}/consumer_tmp ${PACKAGE_PREFIX}/cmd/services/consumer" \
  --build.bin "./${ARTIFACTS_DIR}/consumer_tmp" \
  --proxy.app_port "8889" \
  -proxy.proxy_port "9998" \
  -proxy.enabled "true" \
  --build.log "./${ARTIFACTS_DIR}/consumer-build-errors.log"
