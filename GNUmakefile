default: testacc

# VERSION is the provider version used for local install paths. Bump this
# when cutting a release so `make install` writes to the new directory and
# existing Terraform configs that pin `version = ">= X.Y"` pick it up.
# CI/release builds inject the git tag via `-ldflags "-X main.version=..."`
# and do not use this Makefile install target.
VERSION ?= 0.1.0
OS_ARCH ?= linux_amd64

build:
	go build -o terraform-provider-sporkops

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/sporkops/sporkops/$(VERSION)/$(OS_ARCH)
	cp terraform-provider-sporkops ~/.terraform.d/plugins/registry.terraform.io/sporkops/sporkops/$(VERSION)/$(OS_ARCH)/

testacc:
	TF_ACC=1 go test ./internal/provider/ -v -tags=acceptance $(TESTARGS) -timeout 120m

generate:
	go generate ./...

lint:
	golangci-lint run ./...

.PHONY: default build install testacc generate lint
