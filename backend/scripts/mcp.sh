#!/usr/bin/env bash
set -euo pipefail

# Run MCP service with air
# Usage: mcp.sh <package_prefix> <artifacts_dir>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"
ARTIFACTS_DIR="${2:-artifacts}"

go tool air \
  --build.cmd "go build -o ./${ARTIFACTS_DIR}/mcp_tmp ${PACKAGE_PREFIX}/cmd/ddb" \
  --build.bin "./${ARTIFACTS_DIR}/mcp_tmp serve mcp" \
  --proxy.app_port "8888" \
  -proxy.proxy_port "9999" \
  -proxy.enabled "true" \
  --build.log "./${ARTIFACTS_DIR}/mcp-build-errors.log"
