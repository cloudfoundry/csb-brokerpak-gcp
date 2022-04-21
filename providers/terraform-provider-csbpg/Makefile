.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: cloudfoundry.org ## build the provider

cloudfoundry.org: *.go */*.go
	mkdir -p cloudfoundry.org/cloud-service-broker/csbpg/1.0.0/linux_amd64
	mkdir -p cloudfoundry.org/cloud-service-broker/csbpg/1.0.0/darwin_amd64
	GOOS=linux go build -o cloudfoundry.org/cloud-service-broker/csbpg/1.0.0/linux_amd64/terraform-provider-csbpg_v1.0.0
	GOOS=darwin go build -o cloudfoundry.org/cloud-service-broker/csbpg/1.0.0/darwin_amd64/terraform-provider-csbpg_v1.0.0

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf cloudfoundry.org

.PHONY: test
test: ## run the tests
	go run github.com/onsi/ginkgo/v2/ginkgo -r


.PHONY: init
init: build ## perform terraform init with this provider
	terraform init --plugin-dir .