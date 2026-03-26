default: testacc

build:
	go build -o terraform-provider-sporkops

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/sporkops/sporkops/0.0.1/linux_amd64
	cp terraform-provider-sporkops ~/.terraform.d/plugins/registry.terraform.io/sporkops/sporkops/0.0.1/linux_amd64/

testacc:
	TF_ACC=1 go test ./internal/provider/ -v -tags=acceptance $(TESTARGS) -timeout 120m

generate:
	go generate ./...

lint:
	golangci-lint run ./...

.PHONY: default build install testacc generate lint
