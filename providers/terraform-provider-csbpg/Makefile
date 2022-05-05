.DEFAULT_GOAL = help

GO-VERSION = 1.18.1
GO_OK :=  $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)
ifeq ($(GO_OK), 0)
  GO=go
else ifeq ($(DOCKER_OK), 0)
  GO=docker run --rm -v $(PWD)/../..:/src -w /src/providers/terraform-provider-csbpg -e GOARCH -e GOOS golang:$(GO-VERSION) go
else
  $(error either Go or Docker must be installed)
endif

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: cloudfoundry.org ## build the provider

cloudfoundry.org: *.go */*.go
	mkdir -p cloudfoundry.org/cloud-service-broker/csbpg/1.0.0/linux_amd64
	mkdir -p cloudfoundry.org/cloud-service-broker/csbpg/1.0.0/darwin_amd64
	GOOS=linux $(GO) build -o cloudfoundry.org/cloud-service-broker/csbpg/1.0.0/linux_amd64/terraform-provider-csbpg_v1.0.0
	GOOS=darwin $(GO) build -o cloudfoundry.org/cloud-service-broker/csbpg/1.0.0/darwin_amd64/terraform-provider-csbpg_v1.0.0

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf cloudfoundry.org

.PHONY: test
test: ## run the tests
	## runs docker, so tricky to make it work inside docker
	go run github.com/onsi/ginkgo/v2/ginkgo -r


.PHONY: init
init: build ## perform terraform init with this provider
	terraform init --plugin-dir .