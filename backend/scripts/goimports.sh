#!/usr/bin/env bash
set -euo pipefail

# Format Go imports using goimports
# Usage: goimports.sh [project_root]

PROJECT_ROOT="${1:-$(pwd)}"

# goimports walks with a plain filepath.Walk and has no notion of vendor, so `-w .` would
# rewrite every vendored dependency: ~18k files against ~1.4k of our own. Feed it an
# explicit file list instead, excluding vendor the way format_imports.sh and
# format_golang.sh already do.
go_files=()
while IFS= read -r -d '' file; do
  go_files+=("${file}")
done < <(find "${PROJECT_ROOT}" -type f -not -path '*/vendor/*' -name "*.go" -print0)

if [ ${#go_files[@]} -gt 0 ]; then
  go tool goimports -w "${go_files[@]}"
fi
