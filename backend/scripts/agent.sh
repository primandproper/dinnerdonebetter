#!/usr/bin/env bash
set -euo pipefail

# Run AI agent
# Usage: agent.sh <package_prefix> [args...]

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"
shift || true

# Default args if none provided
if [ $# -eq 0 ]; then
  set -- web api webui
fi

go run "${PACKAGE_PREFIX}/cmd/tools/aiagent" "$@"
