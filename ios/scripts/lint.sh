#!/usr/bin/env bash
set -euo pipefail

# Lint Swift code using the pinned swiftlint binary
# Usage: lint.sh <swiftlint_binary> [--fix]

SWIFTLINT="${1}"

if [ "${2:-}" = "--fix" ]; then
  "${SWIFTLINT}" lint --fix
else
  "${SWIFTLINT}" lint --strict
fi
