SHELL := /bin/bash

GO ?= go
NPM ?= bun
OUTPUT ?= dbmigrator
GO_CACHE_DIR ?= $(CURDIR)/.cache/go-build
GO_TEST_ENV = GOCACHE="$(GO_CACHE_DIR)"
COVERAGE_MIN ?= 71.5

.PHONY: offline build frontend verify test clean run run-web build-frontend coverage coverage-check

build: offline

offline:
	@NPM_BIN="$(NPM)" GO_BIN="$(GO)" ./scripts/build_offline.sh "$(abspath $(OUTPUT))"

frontend:
	@cd frontend && $(NPM) run verify:fast

verify:
	@cd frontend && $(NPM) run verify:fast
	@mkdir -p "$(GO_CACHE_DIR)"
	@$(GO_TEST_ENV) $(GO) test ./... -count=1

test:
	@cd frontend && $(NPM) run test
	@mkdir -p "$(GO_CACHE_DIR)"
	@$(GO_TEST_ENV) $(GO) test ./... -count=1

coverage:
	@mkdir -p "$(GO_CACHE_DIR)"
	@$(GO_TEST_ENV) $(GO) test ./... -count=1 -coverprofile=coverage.out
	@$(GO_TEST_ENV) $(GO) tool cover -func=coverage.out | tail -n 1

coverage-check:
	@mkdir -p "$(GO_CACHE_DIR)"
	@$(GO_TEST_ENV) $(GO) test ./... -count=1 -coverprofile=coverage.out
	@coverage="$$( $(GO_TEST_ENV) $(GO) tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/, "", $$3); print $$3}' )"; \
	awk -v coverage="$$coverage" -v min="$(COVERAGE_MIN)" 'BEGIN { if (coverage + 0 < min + 0) { printf("coverage %.1f%% is below minimum %.1f%%\n", coverage + 0, min + 0) > "/dev/stderr"; exit 1 } }'

run:
	@$(GO) run . $(ARGS)

run-web: build-frontend
	@$(GO) run . -web $(ARGS)

build-frontend:
	@cd frontend && $(NPM) run build

clean:
	@rm -f "$(OUTPUT)"
	@rm -rf frontend/dist
	@rm -f frontend/tsconfig.app.tsbuildinfo frontend/tsconfig.node.tsbuildinfo
	@rm -rf .cache/go-build
