#!/usr/bin/env bash
set -euo pipefail

# Run tests
# Usage: test.sh
#        RUN_CONTAINER_TESTS=false test.sh   # skip the container-backed tests

# The platform's container helpers gate on RUN_CONTAINER_TESTS=true (see
# platform-go/testutils/containers), so container-backed tests skip themselves unless it
# is set. Default it on here to keep `make test` running the postgres suites, while still
# letting a caller opt out on a machine without a Docker daemon.
export RUN_CONTAINER_TESTS="${RUN_CONTAINER_TESTS:-true}"

# shellcheck disable=SC2086,SC2046
CGO_ENABLED=1 go test -shuffle=on -race -vet=all -failfast $(go list github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/... | grep -Ev '(cmd|integration|mock|fakes|converters|utils|generated)')
