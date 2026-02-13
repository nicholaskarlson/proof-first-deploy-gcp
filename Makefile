SHELL := /usr/bin/env bash

.PHONY: fmt test demo verify

fmt:
	gofmt -w cmd internal

test:
	go test -count=1 ./...

demo:
	rm -rf ./out
	mkdir -p ./out
	go run ./cmd/pfdeploy demo --out ./out

# Proof gate: non-mutating (does not run gofmt -w).
verify: test demo
	@echo OK
