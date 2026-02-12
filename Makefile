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

verify: fmt test demo
	@echo OK
