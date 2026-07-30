# ENVIRONMENT
PWD      := $(shell pwd)
MYSELF   := $(shell id -u)
MY_GROUP := $(shell id -g)

# PATHS
THIS          := github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend
ARTIFACTS_DIR := artifacts
SCRIPTS_DIR   := scripts
COVERAGE_OUT  := $(ARTIFACTS_DIR)/coverage.out

# COMPUTED
TOTAL_PACKAGE_LIST := $(shell go list $(THIS)/...)

# CONTAINER VERSIONS
CONTAINER_LINTER_IMAGE := openpolicyagent/conftest:v0.54.0
SQL_GENERATOR_IMAGE    := sqlc/sqlc:1.26.0
LINTER_IMAGE           := golangci/golangci-lint:v2.10.1
SHELLCHECK_IMAGE       := koalaman/shellcheck:stable
GO_IMAGE               := golang:1.26-trixie

# COMMANDS
CONTAINER_RUNNER      := docker
RUN_CONTAINER         := $(CONTAINER_RUNNER) run --rm --volume $(PWD):$(PWD) --workdir=$(PWD) --network=host
RUN_CONTAINER_AS_USER := $(RUN_CONTAINER) --user $(MYSELF):$(MY_GROUP)

# The linter runs with --rm, so without a named volume golangci-lint's analysis cache dies with
# the container and every run is a cold full type-check of the module — including reading all of
# vendor/ to resolve imports, which is most of the wall clock. The volume is keyed to the image
# tag so bumping the linter starts from a clean cache rather than a stale one.
LINTER_CACHE_VOLUME   := golangci-lint-cache-$(subst :,-,$(subst /,-,$(LINTER_IMAGE)))
LINTER                := $(RUN_CONTAINER) --volume $(LINTER_CACHE_VOLUME):/root/.cache $(LINTER_IMAGE) golangci-lint
SQL_GENERATOR         := $(RUN_CONTAINER_AS_USER) $(SQL_GENERATOR_IMAGE)

## non-PHONY folders/files

clean:
	$(SCRIPTS_DIR)/clean.sh $(ARTIFACTS_DIR)

$(ARTIFACTS_DIR):
	@mkdir -p $(ARTIFACTS_DIR)

clean-$(ARTIFACTS_DIR):
	@rm -rf $(ARTIFACTS_DIR)

.PHONY: setup
setup: $(ARTIFACTS_DIR) revendor configs

.PHONY: clean_vendor
clean_vendor:
	$(SCRIPTS_DIR)/clean_vendor.sh

vendor:
	$(SCRIPTS_DIR)/vendor.sh

.PHONY: revendor
revendor: clean_vendor vendor

## FORMATTING

.PHONY: format_imports
format_imports:
	$(SCRIPTS_DIR)/format_imports.sh $(THIS) $(PWD)

.PHONY: format_go_fieldalignment
format_go_fieldalignment:
	@$(SCRIPTS_DIR)/format_go_fieldalignment.sh

.PHONY: format_go_tag_alignment
format_go_tag_alignment:
	@$(SCRIPTS_DIR)/format_go_tag_alignment.sh

.PHONY: go_fix
go_fix:
	go fix ./...

.PHONY: format_golang
format_golang: go_fix goimports format_imports format_go_fieldalignment format_go_tag_alignment
	@$(SCRIPTS_DIR)/format_golang.sh $(PWD)

.PHONY: terraformat
terraformat:
	@$(SCRIPTS_DIR)/terraformat.sh

.PHONY: format
format: format_golang terraformat

.PHONY: fmt
fmt: format

.PHONY: goimports
goimports:
	$(SCRIPTS_DIR)/goimports.sh

## LINTING

.PHONY: lint_containers
lint_containers:
	@$(SCRIPTS_DIR)/lint_containers.sh $(CONTAINER_RUNNER) $(CONTAINER_LINTER_IMAGE) "$(RUN_CONTAINER_AS_USER)" $(PWD)

.PHONY: queries_lint
queries_lint:
	@$(SCRIPTS_DIR)/queries_lint.sh $(CONTAINER_RUNNER) $(SQL_GENERATOR_IMAGE) "$(SQL_GENERATOR)"

.PHONY: golang_lint
golang_lint: struct_usage_check
	@$(SCRIPTS_DIR)/golang_lint.sh $(CONTAINER_RUNNER) $(LINTER_IMAGE) "$(LINTER)"

.PHONY: lint_terraform
lint_terraform: terraformat
	@$(SCRIPTS_DIR)/lint_terraform.sh deploy/environments/prod/terraform

.PHONY: shellcheck
shellcheck:
	@$(SCRIPTS_DIR)/shellcheck.sh $(CONTAINER_RUNNER) $(SHELLCHECK_IMAGE) $(SCRIPTS_DIR)

.PHONY: lint
lint: lint_containers queries_lint golang_lint shellcheck

## GENERATED FILES

.PHONY: configs
configs:
	$(SCRIPTS_DIR)/configs.sh $(THIS)

.PHONY: queries
queries:
	$(SCRIPTS_DIR)/queries.sh $(THIS)

.PHONY: querier
querier: queries queries_lint
	@$(SCRIPTS_DIR)/querier.sh "$(SQL_GENERATOR)"

.PHONY: struct_usage_check
struct_usage_check:
	$(SCRIPTS_DIR)/struct_usage_check.sh

.PHONY: check_struct_usage
check_struct_usage: struct_usage_check

.PHONY: env_vars
env_vars:
	$(SCRIPTS_DIR)/env_vars.sh $(THIS)
	$(MAKE) format_golang

.PHONY: mocks
mocks:
	go generate ./...
	$(MAKE) format_golang

.PHONY: generated_files
generated_files: configs queries env_vars mocks format_golang

## INTEGRATION TESTING

.PHONY: integration-tests
integration-tests: integration_tests

.PHONY: integration_test
integration_test: integration_tests

.PHONY: integration_tests
integration_tests: integration_tests_postgres

.PHONY: integration_tests_postgres
integration_tests_postgres:
	$(SCRIPTS_DIR)/integration_tests_postgres.sh $(THIS)

## EXECUTION

.PHONY: build
build:
	$(SCRIPTS_DIR)/build.sh $(TOTAL_PACKAGE_LIST)

.PHONY: build-api
build-api: $(ARTIFACTS_DIR)
	$(SCRIPTS_DIR)/build.sh -o $(ARTIFACTS_DIR)/api $(THIS)/cmd/services/api

.PHONY: build-bootstrap
build-bootstrap: $(ARTIFACTS_DIR)
	$(SCRIPTS_DIR)/build.sh -o $(ARTIFACTS_DIR)/bootstrap $(THIS)/cmd/tools/bootstrap

.PHONY: build-migrate
build-migrate: $(ARTIFACTS_DIR)
	$(SCRIPTS_DIR)/build.sh -o $(ARTIFACTS_DIR)/migrate $(THIS)/cmd/tools/migrate

.PHONY: test
test: $(ARTIFACTS_DIR) vendor
	$(SCRIPTS_DIR)/test.sh

.PHONY: dev
dev: $(ARTIFACTS_DIR)
	$(SCRIPTS_DIR)/dev.sh $(THIS)

AIR_PORT := 9999

.PHONY: admin
admin: $(ARTIFACTS_DIR)
	DINNER_DONE_BETTER_COOKIES_LIFETIME=24h \
	DINNER_DONE_BETTER_API_SERVICE_GRPC_API_SERVER_URL=localhost:8001 \
	DINNER_DONE_BETTER_API_SERVICE_HTTP_API_SERVER_URL=http://localhost:8000 \
	DINNER_DONE_BETTER_API_SERVICE_OAUTH2_API_CLIENT_ID=AAAAAAAAAAAAAAAA \
	DINNER_DONE_BETTER_API_SERVICE_OAUTH2_API_CLIENT_SECRET=AAAAAAAAAAAAAAAA \
	DEVELOPING_LOCALLY=true \
	$(SCRIPTS_DIR)/admin.sh $(THIS) $(ARTIFACTS_DIR)

.PHONY: consumer
consumer: $(ARTIFACTS_DIR)
	DINNER_DONE_BETTER_COOKIES_LIFETIME=24h \
	DINNER_DONE_BETTER_API_SERVICE_GRPC_API_SERVER_URL=localhost:8001 \
	DINNER_DONE_BETTER_API_SERVICE_HTTP_API_SERVER_URL=http://localhost:8000 \
	DINNER_DONE_BETTER_API_SERVICE_OAUTH2_API_CLIENT_ID=AAAAAAAAAAAAAAAA \
	DINNER_DONE_BETTER_API_SERVICE_OAUTH2_API_CLIENT_SECRET=AAAAAAAAAAAAAAAA \
	DEVELOPING_LOCALLY=true \
	$(SCRIPTS_DIR)/consumer.sh $(THIS) $(ARTIFACTS_DIR)

dry_run:
	DRY_RUN=true go run $(THIS)/cmd/localdev/server

.PHONY: mcp
mcp: $(ARTIFACTS_DIR)
	$(SCRIPTS_DIR)/mcp.sh $(THIS) $(ARTIFACTS_DIR)

.PHONY: mcp_inspector
mcp_inspector:
	$(SCRIPTS_DIR)/mcp_inspector.sh

.PHONY: agent
agent:
	$(SCRIPTS_DIR)/agent.sh $(THIS)

# OPERATIONS

GCP_PROJECT_NAME := dinner-done-better-prod

# https://cloud.google.com/sql/docs/postgres/connect-admin-proxy#connect-tcp
.PHONY: proxy_db
proxy_db:
	$(SCRIPTS_DIR)/proxy_db.sh $(GCP_PROJECT_NAME) prod

# Deploy backend to Docker Desktop Kubernetes (localdev profile)
.PHONY: deploy_localdev
deploy_localdev:
	KO_DOCKER_REPO=ko.local skaffold run --filename=skaffold.yaml --build-concurrency 1 --profile localdev
