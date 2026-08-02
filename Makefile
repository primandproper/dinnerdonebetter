PWD           := $(shell pwd)
MYSELF        := $(shell id -u)
MY_GROUP      := $(shell id -g)

# ──────────────────────────────────────────────────────────────────────────────
# Variables
# ──────────────────────────────────────────────────────────────────────────────

CONTAINER_RUNNER      := docker
RUN_CONTAINER         := $(CONTAINER_RUNNER) run --rm --volume $(PWD):$(PWD) --workdir=$(PWD)
RUN_CONTAINER_AS_USER := $(CONTAINER_RUNNER) run --rm --volume $(PWD):$(PWD) --workdir=$(PWD) --user $(MYSELF):$(MY_GROUP)

PROTOBUF_FORMAT       := bufbuild/buf:1.70.0
FORMAT_PROTOBUFS      := $(RUN_CONTAINER) $(PROTOBUF_FORMAT)
ARTIFACTS_DIR         := artifacts

# Exclude monolithic proto/X/X.proto files (they're duplicates of the split files)
PROTO_FILES_PATH          := $(shell find proto -name "*.proto" -type f ! -regex "proto/\([^/]*\)/\1\.proto")
PROTO_GO_OUTPUT_PATH      := backend
PROTO_OUTPUT_BACKEND_PATH := backend/internal/grpc
PROTO_OUTPUT_IOS_PATH     := ios/ios/Generated
BACKEND_REPO_NAME         := github.com/primandproper/dinnerdonebetter/backend
PROTO_TS_OUTPUT_PATH      := frontend/packages/api-client/src
PROTO_TS_PLUGIN           := frontend/node_modules/.bin/protoc-gen-ts_proto

# ──────────────────────────────────────────────────────────────────────────────
# Setup & prerequisites
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: setup
setup: ensure_yamlfmt_installed
	(cd backend && $(MAKE) setup)

.PHONY: ensure_yamlfmt_installed
ensure_yamlfmt_installed:
ifeq (, $(shell which yamlfmt))
	$(shell go install github.com/google/yamlfmt/cmd/yamlfmt@latest)
endif

.PHONY: ensure_protoc_installed
ensure_protoc_installed:
ifeq (, $(shell which protoc-gen-go-grpc))
	$(shell brew install protobuf)
endif

.PHONY: ensure_protoc-gen-go_installed
ensure_protoc-gen-go_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-go-grpc))
	$(shell go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.4)
endif

.PHONY: ensure_protoc-gen-go-grpc_installed
ensure_protoc-gen-go-grpc_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-go-grpc))
	$(shell go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1)
endif

.PHONY: ensure_protoc-gen-swift_installed
ensure_protoc-gen-swift_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-swift))
	$(shell brew install swift-protobuf)
endif

.PHONY: ensure_protoc-gen-grpc-swift_installed
ensure_protoc-gen-grpc-swift_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-grpc-swift-2))
	$(shell brew install protoc-gen-grpc-swift)
endif

.PHONY: ensure_proto_ts_plugin_installed
ensure_proto_ts_plugin_installed:
	@if [ ! -f $(PROTO_TS_PLUGIN) ]; then \
		echo "Installing frontend dependencies for ts-proto..."; \
		(cd frontend && npm install); \
	fi

# ──────────────────────────────────────────────────────────────────────────────
# Formatting, linting & testing
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: format
format: format_yaml
	(cd backend && $(MAKE) format)
	(cd frontend && $(MAKE) format)
	(cd ios && $(MAKE) format)

.PHONY: format_yaml
format_yaml: ensure_yamlfmt_installed
	yamlfmt -conf .yamlfmt.yaml

.PHONY: terraformat
terraformat:
	(cd backend && $(MAKE) terraformat)

.PHONY: lint
lint:
	(cd backend && $(MAKE) lint)
	(cd frontend && $(MAKE) lint)
	(cd ios && $(MAKE) lint)

.PHONY: lint_markdown
lint_markdown:
	./scripts/lint_markdown.sh

.PHONY: test
test:
	(cd backend && $(MAKE) test)
	(cd frontend && $(MAKE) test)
	(cd ios && $(MAKE) test)

.PHONY: build
build:
	(cd backend && $(MAKE) build)
	(cd frontend && $(MAKE) build)
	(cd ios && $(MAKE) build)

.PHONY: pre_commit
pre_commit: proto build format test lint
	(cd backend && $($MAKE) generated_files)

# ──────────────────────────────────────────────────────────────────────────────
# Frontend dev
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: dev_consumer
dev_consumer:
	(cd frontend && npm run dev:consumer)

.PHONY: dev_admin
dev_admin:
	(cd frontend && npm run dev:admin)

# ──────────────────────────────────────────────────────────────────────────────
# Local deployment
# ──────────────────────────────────────────────────────────────────────────────

# Deploy to Docker Desktop Kubernetes cluster (no Helm required)
.PHONY: deploy_localdev
deploy_localdev:
	skaffold run --filename=infra/skaffold.yaml --build-concurrency 1 --profile localdev
	KO_DOCKER_REPO=ko.local skaffold run --filename=backend/skaffold.yaml --build-concurrency 1 --profile localdev

# ──────────────────────────────────────────────────────────────────────────────
# Production deployment
# ──────────────────────────────────────────────────────────────────────────────

# Deploy prod Terraform: infra first (GKE, networking), then backend. Run from repo root.
# Pass args through, e.g. make deploy_prod_infra ARGS="-auto-approve"
.PHONY: deploy_prod_infra
deploy_prod_infra:
	./infra/scripts/terraform_apply_prod.sh -auto-approve
	(cd backend && ./scripts/terraform_apply_prod.sh -auto-approve)

# Destroy prod Terraform: backend first (k8s resources), then infra (GKE, networking).
# Pass args through, e.g. make destroy_prod_infra ARGS="-auto-approve"
.PHONY: destroy_prod_infra
destroy_prod_infra:
	(cd backend && ./scripts/terraform_destroy_prod.sh -auto-approve)
	./infra/scripts/terraform_destroy_prod.sh -auto-approve

# Prod deploy + verify. Run from repo root. Requires kubectl pointed at prod, grpcurl.
.PHONY: deploy_prod_software
deploy_prod_software:
	./scripts/deploy-prod-local.sh

# Deploy only the frontend (consumer + admin webapps) to prod. Run from repo root.
.PHONY: deploy_prod_frontend
deploy_prod_frontend:
	./scripts/deploy-prod-frontend.sh

.PHONY: verify_prod
verify_prod:
	skaffold verify --filename=skaffold.yaml --profile prod

.PHONY: full_prod_deploy
full_prod_deploy: deploy_prod_infra deploy_prod_software verify_prod

# ──────────────────────────────────────────────────────────────────────────────
# Protobuf generation
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: format_proto
format_proto:
	$(FORMAT_PROTOBUFS) format proto --write

.PHONY: proto_golang
proto_golang: ensure_protoc_installed ensure_protoc-gen-go_installed ensure_protoc-gen-go-grpc_installed
	rm -rf $(ARTIFACTS_DIR)/proto_golang
	mkdir -p $(ARTIFACTS_DIR)/proto_golang
	protoc --go_out=$(ARTIFACTS_DIR)/proto_golang \
		--go-grpc_out=$(ARTIFACTS_DIR)/proto_golang \
		--go_opt=module=$(BACKEND_REPO_NAME) \
		--go-grpc_opt=module=$(BACKEND_REPO_NAME) \
		--proto_path proto/ \
		$(PROTO_FILES_PATH);
	rm -rf $(PROTO_OUTPUT_BACKEND_PATH)/generated
	mv $(ARTIFACTS_DIR)/proto_golang/internal/grpc/generated $(PROTO_OUTPUT_BACKEND_PATH)/generated
	rm -rf $(ARTIFACTS_DIR)/proto_golang
	(cd backend && $(MAKE) format_golang)

.PHONY: proto_swift
proto_swift: ensure_protoc-gen-swift_installed ensure_protoc-gen-grpc-swift_installed
	rm -rf $(ARTIFACTS_DIR)/proto_swift
	mkdir -p $(ARTIFACTS_DIR)/proto_swift
	protoc --swift_out=$(ARTIFACTS_DIR)/proto_swift \
		--grpc-swift-2_out=$(ARTIFACTS_DIR)/proto_swift \
      	--grpc-swift-2_opt=Client=true,Server=false \
      	--swift_opt=Visibility=Public \
		--proto_path proto/ \
		$(PROTO_FILES_PATH)
	rm -rf $(PROTO_OUTPUT_IOS_PATH)
	mv $(ARTIFACTS_DIR)/proto_swift $(PROTO_OUTPUT_IOS_PATH)
	(cd ios && $(MAKE) format)

# Hand-written files in the TS proto output directory that must survive regeneration
PROTO_TS_HANDWRITTEN := index.ts create-clients.ts admin-clients.ts

.PHONY: proto_typescript
proto_typescript: ensure_protoc_installed ensure_proto_ts_plugin_installed
	rm -rf $(ARTIFACTS_DIR)/proto_typescript
	mkdir -p $(ARTIFACTS_DIR)/proto_typescript
	PATH="$(PWD)/frontend/node_modules/.bin:$$PATH" protoc \
		--ts_proto_out=$(ARTIFACTS_DIR)/proto_typescript \
		--ts_proto_opt=outputServices=grpc-js \
		--ts_proto_opt=esModuleInterop=true \
		--proto_path proto/ \
		$(PROTO_FILES_PATH)
	mkdir -p $(ARTIFACTS_DIR)/proto_ts_handwritten
	for f in $(PROTO_TS_HANDWRITTEN); do \
		if [ -f $(PROTO_TS_OUTPUT_PATH)/$$f ]; then \
			cp $(PROTO_TS_OUTPUT_PATH)/$$f $(ARTIFACTS_DIR)/proto_ts_handwritten/$$f; \
		fi; \
	done
	rm -rf $(PROTO_TS_OUTPUT_PATH)
	mv $(ARTIFACTS_DIR)/proto_typescript $(PROTO_TS_OUTPUT_PATH)
	for f in $(PROTO_TS_HANDWRITTEN); do \
		if [ -f $(ARTIFACTS_DIR)/proto_ts_handwritten/$$f ]; then \
			cp $(ARTIFACTS_DIR)/proto_ts_handwritten/$$f $(PROTO_TS_OUTPUT_PATH)/$$f; \
		fi; \
	done
	rm -rf $(ARTIFACTS_DIR)/proto_ts_handwritten
	(cd frontend && $(MAKE) format)

.PHONY: proto
proto: format_proto proto_golang proto_swift proto_typescript

# ──────────────────────────────────────────────────────────────────────────────
# Utilities
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: regit
regit:
	cd ../
	git clone git@github.com:primandproper/dinnerdonebetter tempdir
	@if [ -n "$(BRANCH)" ]; then \
	  (cd tempdir && git checkout $(BRANCH)); \
	fi
	cp -rf tempdir/.git .
	rm -rf tempdir
