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

BUILD_TAG := $(shell git describe --tags 2>/dev/null)
BUILD_SHA := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u '+%Y/%m/%d:%H:%M:%S')

build:
	CGO_ENABLED=0 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher" cmd/ip-fetcher/*.go

build-all:
	GOOS=darwin  CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_darwin_amd64"  cmd/ip-fetcher/*.go
	GOOS=darwin  CGO_ENABLED=0 GOARCH=arm64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_darwin_arm64"  cmd/ip-fetcher/*.go
	GOOS=linux   CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_linux_amd64"   cmd/ip-fetcher/*.go
	GOOS=linux   CGO_ENABLED=0 GOARCH=arm   go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_linux_arm"     cmd/ip-fetcher/*.go
	GOOS=linux   CGO_ENABLED=0 GOARCH=arm64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_linux_arm64"   cmd/ip-fetcher/*.go
	GOOS=netbsd  CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_netbsd_amd64"  cmd/ip-fetcher/*.go
	GOOS=openbsd CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_openbsd_amd64" cmd/ip-fetcher/*.go
	GOOS=freebsd CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_freebsd_amd64" cmd/ip-fetcher/*.go
	GOOS=windows CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_windows_amd64.exe" cmd/ip-fetcher/*.go

clean:
	rm -rf dist

install:
	go install ./cmd/...

build-linux:
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o "dist/ip-fetcher_linux_amd64" cmd/ip-fetcher/.go

mac-install: build
	install dist/ip-fetcher /usr/local/bin/ip-fetcher

linux-install: build-linux
	sudo install dist/ip-fetcher_linux_amd64 /usr/local/bin/ip-fetcher

NAME   := ghcr.io/jonhadfield/ip-fetcher
TAG    := $(shell git rev-parse --short HEAD)
IMG    := ${NAME}:${TAG}
LATEST := ${NAME}:latest

build-docker:
	docker build --platform=linux/x86_64 --build-arg VERSION_VAR="[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC" -t ${IMG} .
	docker tag ${IMG} ${LATEST}
	docker tag ${LATEST} ip-fetcher:latest

scan-image:
	trivy image ip-fetcher:latest

release-docker: build-docker scan-image docker-push

docker-push: login
	docker --log-level debug push ${IMG}
	docker --log-level debug push ${LATEST}

find-updates:
	go list -u -m -json all | go-mod-outdated -update -direct

login:
	@echo ${CR_PAT} | docker login ghcr.io -u jonhadfield --password-stdin

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
