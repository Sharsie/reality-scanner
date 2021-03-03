BUILD_DIR 		:= $(shell pwd)
export GOBIN 		:= $(BUILD_DIR)/bin
export GO111MODULE 	:= on


GIT_TAG		:= $(shell git describe --tags --exact-match 2>/dev/null)
GIT_HASH	:= $(shell git rev-parse HEAD)
VERSION		:= $(or $(GIT_TAG),latest)
BUILD_TIME	:= $(shell date '+%s')

# Runtime environment variables for docker mounts
DOCKER_USER	:= $(shell id -u):$(shell id -g)
# TODO: Does not work, sets to empty string
#HOME_CACHE_DIR  := $(or $${XDG_CACHE_HOME},$$HOME/.cache)
HOME_CACHE_DIR  := $$HOME/.cache
GO_CACHE_DIR	:= $(or $(GOCACHE), $(HOME_CACHE_DIR)/go-build)

REGISTRY	:= dr.cechis.cz
APP_NAME 	:= reality-scanner

-include .env

all: clean build
.PHONY: all


clean:
	@go clean -mod readonly ./...
	@rm -rf $(GOBIN)/*
.PHONY: clean

test:
	@go test -mod readonly ./...
.PHONY: test


build/%:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod readonly -ldflags "-s -w \
		-X github.com/Sharsie/$(APP_NAME)/cmd/$*/version.Tag=$(VERSION)" -o ./bin/linux/amd64/$* ./cmd/$*
build-arm32v7/%:
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -mod readonly -ldflags "-s -w \
		-X github.com/Sharsie/$(APP_NAME)/cmd/$*/version.Tag=$(VERSION)" -o ./bin/linux/arm32v7/$* ./cmd/$*

build: clean build/scan
build-arm32v7: clean build-arm32v7/scan
.PHONY: build build-arm32v7


release/%:
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_HASH=$(GIT_HASH) \
		--build-arg REPO_URL=https://github.com/Sharsie/$(APP_NAME) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg COMMAND=$* \
		--build-arg APP_NAME=$(APP_NAME) \
		-t $(REGISTRY)/$(APP_NAME)-$*:$(VERSION) \
		-f $(BUILD_DIR)/docker/linux/amd64/Dockerfile \
		.
	@docker tag $(REGISTRY)/$(APP_NAME)-$*:$(VERSION) $(REGISTRY)/$(APP_NAME)-$*:latest
	@docker push $(REGISTRY)/$(APP_NAME)-$*:$(VERSION)
	@docker push $(REGISTRY)/$(APP_NAME)-$*:latest

release-arm32v7/%:
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMAND=$* \
		--build-arg APP_NAME=$(APP_NAME) \
		-t $(REGISTRY)/$(APP_NAME)-$*:$(VERSION)-arm32v7 \
		-f $(BUILD_DIR)/docker/linux/arm32v7/Dockerfile \
		.
	@docker tag $(REGISTRY)/$(APP_NAME)-$*:$(VERSION)-arm32v7 $(REGISTRY)/$(APP_NAME)-$*:latest-arm32v7
	@docker push $(REGISTRY)/$(APP_NAME)-$*:$(VERSION)-arm32v7
	@docker push $(REGISTRY)/$(APP_NAME)-$*:latest-arm32v7


release: release/scan
release-arm32v7: release-arm32v7/scan
.PHONY: release release-arm32v7


develop/%:
	@go run --tags dev -race ./cmd/$*

develop: develop/scan
.PHONY: develop

scp: test
	@scp -rp ../reality-scanner punctuator:/tmp
.PHONY: scp
