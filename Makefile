# Set the shell to bash always
SHELL := /bin/bash

# Get the operating system
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

# Look for a .env file, and if present, set make variables from it.
ifneq (,$(wildcard ./.env))
	include .env
	export $(shell sed 's/=.*//' .env)
endif

KIND_CLUSTER_NAME ?= local-dev
KUBECONFIG ?= $(HOME)/.kube/config
KIND_CONFIG=$(shell pwd)/kind-cluster.yaml
KUBECTL_VERSION ?= v1.27.3

VERSION := $(shell git describe --always --tags | sed 's/-/./2' | sed 's/-/./2')
ifndef VERSION
VERSION := 0.0.0
endif

# Tools
KIND ?= $(shell which kind)
ifeq ($(KIND),)
	KIND = $(shell go install sigs.k8s.io/kind@latest && which kind)
endif

LINT=$(shell which golangci-lint)
KUBECTL ?= $(shell which kubectl)
ifeq ($(KUBECTL),)
	KUBECTL = $(shell curl -LO "https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/$(OS)/amd64/kubectl" && chmod +x kubectl && sudo mv kubectl /usr/local/bin/ && which kubectl)
endif

SED=$(shell which sed)

.DEFAULT_GOAL := help

.PHONY: check-kind
check-kind: ## verify kind is installed
	@if [ ! -x "$(KIND)" ]; then \
		echo "kind is not installed. Installing..." && \
		go install sigs.k8s.io/kind@latest; \
	fi

.PHONY: check-kubectl
check-kubectl: ## verify kubectl is installed
	@if [ ! -x "$(KUBECTL)" ]; then \
		echo "kubectl is not installed. Installing..." && \
		curl -LO "https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/$(OS)/amd64/kubectl" && \
		chmod +x kubectl && \
		sudo mv kubectl /usr/local/bin/; \
	fi

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: generate
generate: ## generate all CRDs
generate: tidy
	go generate ./...

.PHONY: dev
dev: ## run the controller in debug mode
dev: generate 
	$(KUBECTL) apply -f package/crds/ -R
	go run cmd/provider/main.go -d

.PHONY: test
test: ## go test
	go test -v ./...

.PHONY: lint
lint: ## go lint
	$(LINT) run

.PHONY: kind-up
kind-up: check-kind check-kubectl ## starts a KinD cluster for local development
	@if ! $(KIND) get clusters | grep -q "^$(KIND_CLUSTER_NAME)$$"; then \
		echo "Creating kind cluster '$(KIND_CLUSTER_NAME)'..."; \
		$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --config=$(KIND_CONFIG); \
	else \
		echo "Kind cluster '$(KIND_CLUSTER_NAME)' already exists."; \
	fi
	@$(KIND) get kubeconfig --name $(KIND_CLUSTER_NAME) > $(KUBECONFIG)

.PHONY: kind-down
kind-down: check-kind ## shuts down the KinD cluster
	@echo "Deleting kind cluster '$(KIND_CLUSTER_NAME)'..."
	@$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: crossplane
crossplane: ## install Crossplane into the local KinD cluster
	$(KUBECTL) create namespace crossplane-system || true
	helm repo add crossplane-stable https://charts.crossplane.io/stable
	helm repo update
	helm install crossplane --namespace crossplane-system crossplane-stable/crossplane

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' ./Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'