SOURCE_FILES?=$$(go list ./...)
TEST_PATTERN?=.
TEST_OPTIONS?=-race -v

setup:
	go get -u github.com/go-critic/go-critic/...
	go get -u github.com/psampaz/go-mod-outdated
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.50.1
	go get -u golang.org/x/tools/cmd/cover
	go install mvdan.cc/gofumpt@latest

test:
	echo 'mode: atomic' > coverage.txt && go list ./... | xargs -n1 -I{} sh -c 'go test -v -timeout=600s -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.txt' && rm coverage.tmp

cover: test
	go tool cover -html=coverage.txt

fmt:
	find . -name '*.go' | while read -r file; do gofumpt -w "$$file"; done

lint:
	golangci-lint run ./...

ci: lint test

critic:
	gocritic check ./...

find-updates:
	go list -u -m -json all | go-mod-outdated -update -direct

login:
	@echo ${CR_PAT} | docker login ghcr.io -u jonhadfield --password-stdin

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
