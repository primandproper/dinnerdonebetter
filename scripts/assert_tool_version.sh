#!/usr/bin/env bash
set -euo pipefail

# Fail unless a tool reports exactly the expected version.
# Usage: assert_tool_version.sh <tool> <expected `<tool> --version` output> <how to fix>
#
# Code generators stamp or shape their output by version, so the wrong one produces a
# diff made entirely of toolchain noise. Refusing to run names the version that is wrong
# instead of leaving someone to work that out from the diff.

TOOL="${1}"
EXPECTED="${2}"
FIX="${3}"

if ! command -v "${TOOL}" &> /dev/null; then
  echo "${TOOL} is not installed, expected '${EXPECTED}': ${FIX}" >&2
  exit 1
fi

ACTUAL="$("${TOOL}" --version 2>&1 | head -n 1)"
if [ "${ACTUAL}" != "${EXPECTED}" ]; then
  echo "${TOOL} is '${ACTUAL}', expected '${EXPECTED}': ${FIX}" >&2
  exit 1
fi
