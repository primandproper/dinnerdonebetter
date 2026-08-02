#!/usr/bin/env bash
set -euo pipefail

# Format Go imports using gci
# Usage: format_imports.sh <package_prefix> <project_root>

PACKAGE_PREFIX="${1:-github.com/primandproper/dinnerdonebetter/backend}"
PROJECT_ROOT="${2:-$(pwd)}"

# Find all Go files and pass them to gci
go_files=()
while IFS= read -r -d '' file; do
  go_files+=("${file}")
done < <(find "${PROJECT_ROOT}" -type f -not -path '*/vendor/*' -name "*.go" -print0)

if [ ${#go_files[@]} -gt 0 ]; then
  # Sections must stay in sync with the gci block in .golangci.yml, or `make format`
  # and `make lint` will disagree about import order.
  go tool gci write \
    --section standard \
    --section "prefix(${PACKAGE_PREFIX})" \
    --section "prefix(github.com/primandproper)" \
    --section default \
    --custom-order \
    "${go_files[@]}"
fi
