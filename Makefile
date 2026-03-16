SHELL := /bin/bash

GO ?= go
NPM ?= npm
OUTPUT ?= dbmigrator

.PHONY: offline build frontend verify test clean

build: offline

offline:
	@NPM_BIN="$(NPM)" GO_BIN="$(GO)" ./scripts/build_offline.sh "$(abspath $(OUTPUT))"

frontend:
	@cd frontend && $(NPM) run verify:fast

verify:
	@cd frontend && $(NPM) run verify:fast
	@$(GO) test ./... -count=1

test:
	@cd frontend && $(NPM) test
	@$(GO) test ./... -count=1

clean:
	@rm -f "$(OUTPUT)"
	@rm -rf frontend/dist
	@rm -f frontend/tsconfig.app.tsbuildinfo frontend/tsconfig.node.tsbuildinfo
