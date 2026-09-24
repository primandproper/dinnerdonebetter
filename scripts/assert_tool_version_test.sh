#!/usr/bin/env bash
set -euo pipefail

# Tests for assert_tool_version.sh.
# Usage: assert_tool_version_test.sh

SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/assert_tool_version.sh"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "${WORKDIR}"' EXIT

FAILURES=0

# fake_tool writes an executable that prints its first argument when asked for its version.
fake_tool() {
  local name="${1}" output="${2}"
  printf '#!/usr/bin/env bash\necho "%s"\n' "${output}" > "${WORKDIR}/${name}"
  chmod +x "${WORKDIR}/${name}"
}

# expect runs the script and checks its exit status and, when given, what it printed.
expect() {
  local description="${1}" want_status="${2}" want_output="${3}"
  shift 3
  local output status=0
  output="$(PATH="${WORKDIR}:/usr/bin:/bin" "${SCRIPT}" "$@" 2>&1)" || status=$?
  if [ "${status}" != "${want_status}" ]; then
    echo "FAIL: ${description}: exit status ${status}, want ${want_status} (output: ${output})"
    FAILURES=$((FAILURES + 1))
  elif [[ "${output}" != *"${want_output}"* ]]; then
    echo "FAIL: ${description}: output '${output}' does not contain '${want_output}'"
    FAILURES=$((FAILURES + 1))
  else
    echo "ok:   ${description}"
  fi
}

version="$((RANDOM % 50)).$((RANDOM % 50))"
other_version="$((RANDOM % 50 + 50)).$((RANDOM % 50))"
fix_hint="install-hint-${RANDOM}"

fake_tool protoc "libprotoc ${version}"
expect "matching version passes silently" 0 "" \
  protoc "libprotoc ${version}" "${fix_hint}"

expect "mismatched version fails and names both versions" 1 \
  "protoc is 'libprotoc ${version}', expected 'libprotoc ${other_version}': ${fix_hint}" \
  protoc "libprotoc ${other_version}" "${fix_hint}"

expect "a version that only starts with the expected one fails" 1 "expected" \
  protoc "libprotoc ${version%.*}" "${fix_hint}"

expect "missing tool fails and says how to install it" 1 \
  "not-a-real-tool is not installed, expected 'x ${version}': ${fix_hint}" \
  not-a-real-tool "x ${version}" "${fix_hint}"

# A tool can be named by path, which is how the Makefile names the Swift plugins it builds.
fake_tool protoc-gen-swift "protoc-gen-swift ${version}"
expect "tool named by path is checked" 1 "expected 'protoc-gen-swift ${other_version}'" \
  "${WORKDIR}/protoc-gen-swift" "protoc-gen-swift ${other_version}" "${fix_hint}"

# Only the first line counts: some tools follow the version with more output.
printf '#!/usr/bin/env bash\necho "tool %s"\necho "built by someone"\n' "${version}" > "${WORKDIR}/multiline"
chmod +x "${WORKDIR}/multiline"
expect "only the first line of --version is compared" 0 "" \
  multiline "tool ${version}" "${fix_hint}"

if [ "${FAILURES}" -gt 0 ]; then
  echo "${FAILURES} failure(s)"
  exit 1
fi
