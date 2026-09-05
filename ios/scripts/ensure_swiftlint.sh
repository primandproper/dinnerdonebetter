#!/usr/bin/env bash
set -euo pipefail

# Install the pinned SwiftLint release into the build directory, if it is not there already.
# Usage: ensure_swiftlint.sh <version> <install_dir>
#
# The binary is fetched from the GitHub release for exactly that version, so CI and local runs lint
# with the same rule set regardless of what Homebrew currently ships.

VERSION="${1}"
INSTALL_DIR="${2}"
BINARY="${INSTALL_DIR}/swiftlint"

if [ -x "${BINARY}" ] && [ "$("${BINARY}" version)" = "${VERSION}" ]; then
  exit 0
fi

echo "Installing SwiftLint ${VERSION} into ${INSTALL_DIR}"
rm -rf "${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}"
ARCHIVE="$(mktemp -t swiftlint.XXXXXX).zip"
trap 'rm -f "${ARCHIVE}"' EXIT
curl --fail --silent --show-error --location \
  "https://github.com/realm/SwiftLint/releases/download/${VERSION}/portable_swiftlint.zip" \
  --output "${ARCHIVE}"
unzip -q -o "${ARCHIVE}" -d "${INSTALL_DIR}"
chmod +x "${BINARY}"

INSTALLED="$("${BINARY}" version)"
if [ "${INSTALLED}" != "${VERSION}" ]; then
  echo "expected SwiftLint ${VERSION}, got ${INSTALLED}" >&2
  exit 1
fi
