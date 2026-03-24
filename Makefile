SHELL := /bin/bash

GO ?= go
NPM ?= bun
OUTPUT ?= dbmigrator

.PHONY: offline build frontend verify test clean run run-web build-frontend

build: offline

offline:
	@NPM_BIN="$(NPM)" GO_BIN="$(GO)" ./scripts/build_offline.sh "$(abspath $(OUTPUT))"

frontend:
	@cd frontend && $(NPM) run verify:fast

verify:
	@cd frontend && $(NPM) run verify:fast
	@$(GO) test ./... -count=1

test:
	@cd frontend && $(NPM) run test
	@$(GO) test ./... -count=1

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
