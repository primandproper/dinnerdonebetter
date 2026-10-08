PWD           := $(shell pwd)
MYSELF        := $(shell id -u)
MY_GROUP      := $(shell id -g)

# ──────────────────────────────────────────────────────────────────────────────
# Variables
# ──────────────────────────────────────────────────────────────────────────────

CONTAINER_RUNNER      := docker
RUN_CONTAINER         := $(CONTAINER_RUNNER) run --rm --volume $(PWD):$(PWD) --workdir=$(PWD)
RUN_CONTAINER_AS_USER := $(CONTAINER_RUNNER) run --rm --volume $(PWD):$(PWD) --workdir=$(PWD) --user $(MYSELF):$(MY_GROUP)

PROTOBUF_FORMAT       := yoheimuta/protolint:0.57.0
FORMAT_PROTOBUFS      := $(RUN_CONTAINER_AS_USER) $(PROTOBUF_FORMAT)
ARTIFACTS_DIR         := artifacts

PROTO_FILES_PATH          := $(shell find proto -name "*.proto" -type f)

# filtering's schema is not ours to keep a copy of: primitives-go ships the .proto
# inside the published module, so go.mod already pins which version we build
# against. Putting the module's proto directory on protoc's path is how a
# consumer imports it, exactly as google/protobuf/timestamp.proto already works.
PLATFORM_PROTO_PATH       := $(shell cd backend && go list -m -f '{{.Dir}}' github.com/primandproper/primitives-go/v2)/filtering/proto
PLATFORM_FILTERING_PROTO  := primandproper/platform/filtering/v1/filtering.proto
# Go links against the bindings platform already generated rather than making a
# second copy, because the page-size clamp and the default are server-side rules
# and a second copy of one can be wrong in a way nothing reports. Swift and
# TypeScript import it from their client libraries, which ship it generated from
# the same schema (see PROTO_SWIFT_MODULE_MAPPINGS and PROTO_TS_IMPORT_MAPPINGS).
PROTO_GO_FILTERING_MAP    := M$(PLATFORM_FILTERING_PROTO)=github.com/primandproper/primitives-go/v2/filtering/filteringpb

# mediaregistry's schema too: mealplanning's media are the registry's objects, and its uploads
# are the registry's upload stream. The client libraries ship the stubs, so Swift and TypeScript
# map the file to them (see PROTO_SWIFT_MODULE_MAPPINGS and PROTO_TS_IMPORT_MAPPINGS) and Go
# links against platform's bindings.
PLATFORM_MEDIAREGISTRY_PROTO_PATH := $(shell cd backend && go list -m -f '{{.Dir}}' github.com/primandproper/platform-go/v15)/mediaregistry/proto
PLATFORM_MEDIAREGISTRY_PROTO      := primandproper/platform/mediaregistry/v1/mediaregistry.proto
PROTO_GO_MEDIAREGISTRY_MAP        := M$(PLATFORM_MEDIAREGISTRY_PROTO)=github.com/primandproper/platform-go/v15/mediaregistry/mediaregistrypb
PROTO_GO_OUTPUT_PATH      := backend
PROTO_OUTPUT_BACKEND_PATH := backend/internal/grpc
PROTO_OUTPUT_IOS_PATH     := ios/ios/Generated
BACKEND_REPO_NAME         := github.com/primandproper/dinnerdonebetter/backend
PROTO_TS_OUTPUT_PATH      := frontend/packages/api-client/src
PROTO_TS_PLUGIN           := frontend/node_modules/.bin/protoc-gen-ts_proto

# The codegen toolchain is pinned, not just the schema: every generator here stamps or
# shapes its output by version, so an unpinned one turns a regeneration with no schema
# change into a diff nobody can review. protoc's version is written into the header of
# every generated Go and TypeScript file; the Swift plugins changed their output between
# releases (and brew's build of protoc-gen-grpc-swift 2.1.1 is not the same generator as
# the 2.1.1 tag). Each generation target asserts these before it runs, so a mismatch is
# reported as a version rather than discovered as a four-thousand-line diff.
#
# protoc cannot be installed by version from brew, so it comes from mise (mise.toml pins
# it) or any other source with the right version on PATH. The Swift plugins are built
# from the manifests in scripts/protoc-plugins, which pin them exactly and match
# platform-client-swift's, so the two generate identical code from identical schemas.
PROTOC_VERSION                := 33.1
PROTOC_GEN_GO_VERSION         := 1.36.4
PROTOC_GEN_GO_GRPC_VERSION    := 1.5.1
PROTOC_GEN_SWIFT_VERSION      := 1.33.3
PROTOC_GEN_GRPC_SWIFT_VERSION := 2.1.1
ASSERT_TOOL_VERSION           := ./scripts/assert_tool_version.sh
PROTOC_PLUGINS_DIR            := scripts/protoc-plugins
PROTOC_GEN_SWIFT              := $(PROTOC_PLUGINS_DIR)/protoc-gen-swift/.build/release/protoc-gen-swift
PROTOC_GEN_GRPC_SWIFT         := $(PROTOC_PLUGINS_DIR)/protoc-gen-grpc-swift-2/.build/release/protoc-gen-grpc-swift-2

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
	@$(ASSERT_TOOL_VERSION) protoc "libprotoc $(PROTOC_VERSION)" \
		"install protoc $(PROTOC_VERSION) (mise install protoc), or put it first on PATH"

.PHONY: ensure_protoc-gen-go_installed
ensure_protoc-gen-go_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-go))
	$(shell go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC_GEN_GO_VERSION))
endif
	@$(ASSERT_TOOL_VERSION) protoc-gen-go "protoc-gen-go v$(PROTOC_GEN_GO_VERSION)" \
		"go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC_GEN_GO_VERSION)"

.PHONY: ensure_protoc-gen-go-grpc_installed
ensure_protoc-gen-go-grpc_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-go-grpc))
	$(shell go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION))
endif
	@$(ASSERT_TOOL_VERSION) protoc-gen-go-grpc "protoc-gen-go-grpc $(PROTOC_GEN_GO_GRPC_VERSION)" \
		"go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)"

# The Swift plugins are built rather than looked up on PATH, so there is no guard to get
# wrong: swift build is a no-op when the pinned binary is already current.
.PHONY: ensure_protoc-gen-swift_installed
ensure_protoc-gen-swift_installed: ensure_protoc_installed
	swift build -c release --package-path $(PROTOC_PLUGINS_DIR)/protoc-gen-swift --product protoc-gen-swift
	@$(ASSERT_TOOL_VERSION) $(PROTOC_GEN_SWIFT) "protoc-gen-swift $(PROTOC_GEN_SWIFT_VERSION)" \
		"bump $(PROTOC_PLUGINS_DIR)/protoc-gen-swift/Package.swift and PROTOC_GEN_SWIFT_VERSION together"

.PHONY: ensure_protoc-gen-grpc-swift_installed
ensure_protoc-gen-grpc-swift_installed: ensure_protoc_installed
	swift build -c release --package-path $(PROTOC_PLUGINS_DIR)/protoc-gen-grpc-swift-2 --product protoc-gen-grpc-swift-2
	@$(ASSERT_TOOL_VERSION) $(PROTOC_GEN_GRPC_SWIFT) "protoc-gen-grpc-swift-2 $(PROTOC_GEN_GRPC_SWIFT_VERSION)" \
		"bump $(PROTOC_PLUGINS_DIR)/protoc-gen-grpc-swift-2/Package.swift and PROTOC_GEN_GRPC_SWIFT_VERSION together"

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
test: test_scripts
	(cd backend && $(MAKE) test)
	(cd frontend && $(MAKE) test)
	(cd ios && $(MAKE) test)

.PHONY: test_scripts
test_scripts:
	./scripts/assert_tool_version_test.sh

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
	$(FORMAT_PROTOBUFS) lint -fix proto/

# check_proto_format is format_proto without the fixer: it reports what would change
# and exits non-zero, which is what CI runs.
.PHONY: check_proto_format
check_proto_format:
	$(FORMAT_PROTOBUFS) lint proto/

.PHONY: proto_golang
proto_golang: ensure_protoc_installed ensure_protoc-gen-go_installed ensure_protoc-gen-go-grpc_installed
	rm -rf $(ARTIFACTS_DIR)/proto_golang
	mkdir -p $(ARTIFACTS_DIR)/proto_golang
	protoc --go_out=$(ARTIFACTS_DIR)/proto_golang \
		--go-grpc_out=$(ARTIFACTS_DIR)/proto_golang \
		--go_opt=module=$(BACKEND_REPO_NAME) \
		--go-grpc_opt=module=$(BACKEND_REPO_NAME) \
		--go_opt=$(PROTO_GO_FILTERING_MAP) \
		--go-grpc_opt=$(PROTO_GO_FILTERING_MAP) \
		--go_opt=$(PROTO_GO_MEDIAREGISTRY_MAP) \
		--go-grpc_opt=$(PROTO_GO_MEDIAREGISTRY_MAP) \
		--proto_path proto/ \
		--proto_path $(PLATFORM_PROTO_PATH) \
		--proto_path $(PLATFORM_MEDIAREGISTRY_PROTO_PATH) \
		$(PROTO_FILES_PATH);
	rm -rf $(PROTO_OUTPUT_BACKEND_PATH)/generated
	mv $(ARTIFACTS_DIR)/proto_golang/internal/grpc/generated $(PROTO_OUTPUT_BACKEND_PATH)/generated
	rm -rf $(ARTIFACTS_DIR)/proto_golang
	(cd backend && $(MAKE) format_golang)

.PHONY: proto_swift
proto_swift: ensure_protoc-gen-swift_installed ensure_protoc-gen-grpc-swift_installed
	rm -rf $(ARTIFACTS_DIR)/proto_swift
	mkdir -p $(ARTIFACTS_DIR)/proto_swift
	protoc --plugin=protoc-gen-swift=$(PROTOC_GEN_SWIFT) \
		--plugin=protoc-gen-grpc-swift-2=$(PROTOC_GEN_GRPC_SWIFT) \
		--swift_out=$(ARTIFACTS_DIR)/proto_swift \
		--grpc-swift-2_out=$(ARTIFACTS_DIR)/proto_swift \
		--grpc-swift-2_opt=Client=true,Server=false,ProtoPathModuleMappings=$(PROTO_SWIFT_MODULE_MAPPINGS) \
		--swift_opt=Visibility=Public,ProtoPathModuleMappings=$(PROTO_SWIFT_MODULE_MAPPINGS) \
		--proto_path proto/ \
		--proto_path $(PLATFORM_PROTO_PATH) \
		--proto_path $(PLATFORM_MEDIAREGISTRY_PROTO_PATH) \
		$(PROTO_FILES_PATH)
	rm -rf $(PROTO_OUTPUT_IOS_PATH)
	mv $(ARTIFACTS_DIR)/proto_swift $(PROTO_OUTPUT_IOS_PATH)
	(cd ios && $(MAKE) format)

# Hand-written files in the TS proto output directory that must survive regeneration
PROTO_TS_HANDWRITTEN := index.ts platform.ts

# The web apps call platform's services through @primandproper/platform-client, whose
# stubs are generated from the platform-go tag it pins, so nothing platform owns is
# generated here. ts-proto's import mappings send every platform proto this repository's
# own protos import to the library's export for it, and ts-proto generates no file for a
# mapped proto, so the apps and these stubs hold one copy of each platform type, the
# library's. A platform proto newly imported here needs a mapping, or it is generated
# into the tree again.
PROTO_TS_IMPORT_MAPPINGS := \
	--ts_proto_opt=M$(PLATFORM_FILTERING_PROTO)=@primandproper/platform-client/filtering/v1 \
	--ts_proto_opt=M$(PLATFORM_MEDIAREGISTRY_PROTO)=@primandproper/platform-client/mediaregistry/v1

# The iOS app calls platform's services through platform-client-swift, which generates their
# stubs from the platform-go tag it pins, so nothing platform owns is generated here. Unlike
# ts-proto, the Swift generators can point an import at a module: the mappings file sends
# every platform proto this repository's own protos import to PlatformClient, so the app holds
# one copy of each platform type, the library's.
PROTO_SWIFT_MODULE_MAPPINGS := proto/swift_module_mappings.asciipb

.PHONY: proto_typescript
proto_typescript: ensure_protoc_installed ensure_proto_ts_plugin_installed
	rm -rf $(ARTIFACTS_DIR)/proto_typescript
	mkdir -p $(ARTIFACTS_DIR)/proto_typescript
	PATH="$(PWD)/frontend/node_modules/.bin:$$PATH" protoc \
		--ts_proto_out=$(ARTIFACTS_DIR)/proto_typescript \
		--ts_proto_opt=outputServices=grpc-js \
		--ts_proto_opt=esModuleInterop=true \
		$(PROTO_TS_IMPORT_MAPPINGS) \
		--proto_path proto/ \
		--proto_path $(PLATFORM_PROTO_PATH) \
		--proto_path $(PLATFORM_MEDIAREGISTRY_PROTO_PATH) \
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
proto: format_proto proto_golang proto_typescript proto_swift

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
