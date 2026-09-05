# iOS Project Makefile
# Provides convenient commands to run tests and build the iOS project

# Project configuration
PROJECT := ios.xcodeproj
SCHEME := ios
# DESTINATION can be overridden via environment variable (e.g., in CI)
# Default to iPhone 16 Pro with OS 18.2 for local development
# Note: Don't include quotes in the value - they're added in the command
DESTINATION ?= platform=iOS Simulator,name=iPhone 16 Pro,OS=18.2
TEST_TARGET := iosTests
UI_TEST_TARGET := iosUITests

# xcodebuild configuration
XCODEBUILD := xcodebuild
BUILD_DIR := build
DERIVED_DATA := $(BUILD_DIR)/DerivedData
SCRIPTS_DIR := scripts

# Container versions
SHELLCHECK_IMAGE := koalaman/shellcheck:stable
CONTAINER_RUNNER := docker

# SwiftLint is pinned so CI and local runs apply the same rule set. Bump deliberately: new releases
# add rules, and an unpinned install broke CI on code that had not changed.
SWIFTLINT_VERSION := 0.65.1
SWIFTLINT_DIR := $(BUILD_DIR)/tools/swiftlint-$(SWIFTLINT_VERSION)
SWIFTLINT := $(SWIFTLINT_DIR)/swiftlint

.PHONY: ensure_swift-format_installed
ensure_swift-format_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh swift-format "brew install swift-format"

.PHONY: ensure_swiftlint_installed
ensure_swiftlint_installed:
	@$(SCRIPTS_DIR)/ensure_swiftlint.sh $(SWIFTLINT_VERSION) $(SWIFTLINT_DIR)

.PHONY: test
test: unit-test ui-test

.PHONY: unit-test
unit-test:
	$(SCRIPTS_DIR)/unit_test.sh $(XCODEBUILD) $(PROJECT) $(SCHEME) '$(DESTINATION)' $(TEST_TARGET) $(DERIVED_DATA)

.PHONY: ui-test
ui-test:
	$(SCRIPTS_DIR)/ui_test.sh $(XCODEBUILD) $(PROJECT) $(SCHEME) '$(DESTINATION)' $(UI_TEST_TARGET) $(DERIVED_DATA)

.PHONY: build
build:
	$(SCRIPTS_DIR)/build.sh $(XCODEBUILD) $(PROJECT) $(SCHEME) '$(DESTINATION)' $(DERIVED_DATA)

.PHONY: clean
clean:
	$(SCRIPTS_DIR)/clean.sh $(BUILD_DIR) $(XCODEBUILD) $(PROJECT) $(SCHEME) $(DERIVED_DATA)

.PHONY: lint
lint: ensure_swiftlint_installed
	$(SCRIPTS_DIR)/lint.sh $(SWIFTLINT)

.PHONY: lint-autocorrect
lint-autocorrect: ensure_swiftlint_installed
	$(SCRIPTS_DIR)/lint.sh $(SWIFTLINT) --fix

.PHONY: list-destinations
list-destinations: ## List available iOS Simulator destinations
	$(SCRIPTS_DIR)/list_destinations.sh

.PHONY: format
format: ensure_swift-format_installed
	$(SCRIPTS_DIR)/format.sh ios

.PHONY: format-check
format-check: ensure_swift-format_installed
	$(SCRIPTS_DIR)/format.sh ios --check

.PHONY: shellcheck
shellcheck:
	@$(SCRIPTS_DIR)/shellcheck.sh $(CONTAINER_RUNNER) $(SHELLCHECK_IMAGE) $(SCRIPTS_DIR)
