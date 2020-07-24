MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
.DELETE_ON_ERROR:
.SUFFIXES:

DOCKER_CMD   ?= docker
DOCKER_ARGS  ?= run --rm --user "$$(id -u)" -v "$${PWD}:/component" --workdir /component

YAML_FILES      ?= $(shell find . -type f -name '*.yaml' -or -name '*.yml')
YAMLLINT_ARGS   ?= --no-warnings
YAMLLINT_CONFIG ?= .yamllint.yml
YAMLLINT_IMAGE  ?= docker.io/cytopia/yamllint:latest
YAMLLINT_DOCKER ?= $(DOCKER_CMD) $(DOCKER_ARGS) $(YAMLLINT_IMAGE)

VALE_CMD  ?= $(DOCKER_CMD) $(DOCKER_ARGS) --volume "$${PWD}"/docs/modules:/pages vshn/vale:2.1.1
VALE_ARGS ?= --minAlertLevel=error --config=/pages/ROOT/pages/.vale.ini /pages

# Project parameters
BINARY_NAME ?= boatswain
BINARY_VERSION = $(shell git describe --tags --always --dirty --match=v* || (echo "command failed $$?"; exit 1))
VERSION ?= $(BINARY_VERSION)
IMAGE_NAME ?= docker.io/projectsyn/$(BINARY_NAME):$(VERSION)

# Go parameters
GOCMD   ?= go
GOBUILD ?= $(GOCMD) build
GOCLEAN ?= $(GOCMD) clean
GOTEST  ?= $(GOCMD) test
GOGET   ?= $(GOCMD) get

.PHONY: all
all: lint test build

.PHONY: lint
lint: lint_yaml lint_adoc

.PHONY: lint_yaml
lint_yaml: $(YAML_FILES)
	$(YAMLLINT_DOCKER) -f parsable -c $(YAMLLINT_CONFIG) $(YAMLLINT_ARGS) -- $?

.PHONY: lint_adoc
lint_adoc:
	$(VALE_CMD) $(VALE_ARGS)

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v \
		-o $(BINARY_NAME) \
		-ldflags "-w -s -X main.Version=$(VERSION) -X 'main.BuildDate=$(shell date -Iseconds)'"
	@echo built '$(VERSION)'

.PHONY: test
test:
	$(GOTEST) -v -cover ./...

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(web_dir)

.PHONY: docker
docker:
	docker build --build-arg BINARY_VERSION=$(BINARY_VERSION) -t $(IMAGE_NAME) .
	@echo built image $(IMAGE_NAME)
