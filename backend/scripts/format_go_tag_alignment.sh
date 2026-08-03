#!/usr/bin/env bash
set -euo pipefail

# Format Go tag alignment
# Usage: format_go_tag_alignment.sh [max_passes]
#
# Same guard as format_go_fieldalignment.sh: an `until cmd; do true; done` around a linter
# that exits non-zero on every unfixable diagnostic is an infinite loop. tagalign happens
# to converge today, but the loop shape is the hazard, not the tool.

MAX_PASSES="${1:-5}"
# Alphabetical, which is what tagalign's own sort falls back to for anything not named
# here. Keeping the two in agreement is the point: a tag missing from this list gets
# sorted to the end by the formatter and to its alphabetical place by the linter, so
# `make format` and `make lint` would each undo the other.
TAG_ORDER="audit,env,envDefault,envPrefix,json,mapstructure,toml,yaml"

marker="$(mktemp)"
# shellcheck disable=SC2064
trap "rm -f '${marker}'" EXIT

for ((pass = 1; pass <= MAX_PASSES; pass++)); do
  touch "${marker}"

  if go tool tagalign -fix -sort -order "${TAG_ORDER}" ./... > /dev/null 2>&1; then
    break
  fi

  if [ -z "$(find . -type f -name '*.go' -not -path '*/vendor/*' -newer "${marker}" -print -quit)" ]; then
    break
  fi
done
