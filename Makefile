###### Help ###################################################################
.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Setup ##################################################################
IAAS=gcp
CSB_VERSION := $(or $(CSB_VERSION), $(shell grep 'github.com/cloudfoundry/cloud-service-broker' go.mod | grep -v replace | awk '{print $$NF}' | sed -e 's/v//'))
CSB_RELEASE_VERSION := $(CSB_VERSION)

CSB_DOCKER_IMAGE := $(or $(CSB), cfplatformeng/csb:$(CSB_VERSION))
GO_OK :=  $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)

####### broker environment variables
PAK_CACHE=/tmp/.pak-cache
SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)
GSB_COMPATIBILITY_ENABLE_BETA_SERVICES := $(or $(GSB_COMPATIBILITY_ENABLE_BETA_SERVICES), true)
export GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS = [{"name":"small","id":"5b45de36-cb90-11ec-a755-77f8be95a49d","description":"PostgreSQL with default version, shared CPU, minimum 0.6GB ram, 10GB storage","display_name":"small","tier":"db-f1-micro","storage_gb":10},{"name":"medium","id":"a3359fa6-cb90-11ec-bcb6-cb68544eda78","description":"PostgreSQL with default version, shared CPU, minimum 1.7GB ram, 20GB storage","display_name":"medium","tier":"db-g1-small","storage_gb":20},{"name":"large","id":"cd95c5b4-cb90-11ec-a5da-df87b7fb7426","description":"PostgreSQL with default version, minimum 8 cores, minimum 8GB ram, 50GB storage","display_name":"large","tier":"db-n1-standard-8","storage_gb":50}]
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"authorized_network": "$(GCP_PAS_NETWORK)"})

ifeq ($(GO_OK), 0) # use local go binary
GO=go
GOFMT=gofmt
BROKER_GO_OPTS=PORT=8080 \
				DB_TYPE=sqlite3 \
				DB_PATH=/tmp/csb-db \
				SECURITY_USER_NAME=$(SECURITY_USER_NAME) \
				SECURITY_USER_PASSWORD=$(SECURITY_USER_PASSWORD) \
				GOOGLE_CREDENTIALS='$(GOOGLE_CREDENTIALS)' \
				GOOGLE_PROJECT=$(GOOGLE_PROJECT) \
 				PAK_BUILD_CACHE_PATH=$(PAK_CACHE) \
 				GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)' \
 				GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS='$(GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS)' \
 				GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=$(GSB_COMPATIBILITY_ENABLE_BETA_SERVICES)

PAK_PATH=$(PWD) #where the bokerpak zip resides
RUN_CSB=$(BROKER_GO_OPTS) go run github.com/cloudfoundry/cloud-service-broker
LDFLAGS="-X github.com/cloudfoundry/cloud-service-broker/utils.Version=$(CSB_VERSION)"
GET_CSB="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) github.com/cloudfoundry/cloud-service-broker"
else ifeq ($(DOCKER_OK), 0) ## running the broker and go with docker
BROKER_DOCKER_OPTS=--rm -v $(PAK_CACHE):$(PAK_CACHE) -v $(PWD):/brokerpak -w /brokerpak --network=host \
  -p 8080:8080 \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e GOOGLE_CREDENTIALS \
	-e GOOGLE_PROJECT \
	-e "DB_TYPE=sqlite3" \
	-e "DB_PATH=/tmp/csb-db" \
	-e PAK_BUILD_CACHE_PATH=$(PAK_CACHE) \
	-e GSB_PROVISION_DEFAULTS \
	-e GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS \
	-e GSB_COMPATIBILITY_ENABLE_BETA_SERVICES

RUN_CSB=docker run $(BROKER_DOCKER_OPTS) $(CSB_DOCKER_IMAGE)

#### running go inside a container, this is for integration tests and push-broker
# path inside the container
PAK_PATH=/brokerpak

GO_DOCKER_OPTS=--rm -v $(PAK_CACHE):$(PAK_CACHE) -v $(PWD):/brokerpak -w /brokerpak --network=host
GO=docker run $(GO_DOCKER_OPTS) golang:latest go
GOFMT=docker run $(GO_DOCKER_OPTS) golang:latest gofmt

# this doesnt work well if we did make latest-csb. We should build it instead, with go inside a container.
GET_CSB="wget -O cloud-service-broker https://github.com/cloudfoundry/cloud-service-broker/releases/download/v$(CSB_RELEASE_VERSION)/cloud-service-broker.linux && chmod +x cloud-service-broker"
else
$(error either Go or Docker must be installed)
endif

###### Targets ################################################################

.PHONY: build
build: $(IAAS)-services-*.brokerpak ## build brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*/*.tf ./providers/terraform-provider-csbpg/cloudfoundry.org/cloud-service-broker/csbpg | $(PAK_CACHE)
	$(RUN_CSB) pak build

.PHONY: run
run: build google_credentials google_project ## start CSB in a docker container
	$(RUN_CSB) serve

.PHONY: docs
docs: build brokerpak-user-docs.md ## build docs

brokerpak-user-docs.md: *.yml
	$(RUN_CSB) pak docs $(PAK_PATH)/$(shell ls *.brokerpak) > $@ # GO

.PHONY: examples
examples: ## display available examples
	 $(RUN_CSB) client examples

PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 10000)

.PHONY: run-examples
run-examples: ## run examples against CSB on localhost (run "make run" to start it), set service_name and example_name to run specific example
	$(RUN_CSB) client run-examples --service-name="$(service_name)" --example-name="$(example_name)" -j $(PARALLEL_JOB_COUNT)

.PHONY: test ## run the tests
test: lint provider-tests run-integration-tests

.PHONY: run-integration-tests
run-integration-tests: ## run integration tests for this brokerpak
	cd ./integration-tests && go run github.com/onsi/ginkgo/v2/ginkgo -r .

.PHONY: provider-tests
provider-tests:
	cd providers/terraform-provider-csbpg; $(MAKE) test

.PHONY: info
info: build ## use the CSB to parse the buildpak and print out contents and versions
	$(RUN_CSB) pak info $(PAK_PATH)/$(shell ls *.brokerpak)

.PHONY: validate
validate: build ## use the CSB to validate the buildpak
	$(RUN_CSB) pak validate $(PAK_PATH)/$(shell ls *.brokerpak)

# fetching bits for cf push broker
.PHONY: cloud-service-broker
cloud-service-broker: go.mod ## build or fetch CSB binary
	$(shell "$(GET_CSB)")

APP_NAME := $(or $(APP_NAME), cloud-service-broker-gcp)
DB_TLS := $(or $(DB_TLS), skip-verify)


.PHONY: push-broker
push-broker: cloud-service-broker build google_credentials google_project gcp_pas_network ## push the broker to targeted Cloud Foundry
	MANIFEST=cf-manifest.yml APP_NAME=$(APP_NAME) DB_TLS=$(DB_TLS) GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)' ./scripts/push-broker.sh

.PHONY: google_credentials
google_credentials:
ifndef GOOGLE_CREDENTIALS
	$(error variable GOOGLE_CREDENTIALS not defined)
endif

.PHONY: google_project
google_project:
ifndef GOOGLE_PROJECT
	$(error variable GOOGLE_PROJECT not defined)
endif

.PHONY: gcp_pas_network
gcp_pas_network:
ifndef GCP_PAS_NETWORK
	$(error variable GCP_PAS_NETWORK not defined - must be GCP network for PAS foundation)
endif

.PHONY: clean
clean: ## clean up build artifacts
	- rm -f $(IAAS)-services-*.brokerpak
	- rm -f ./cloud-service-broker
	- rm -f ./brokerpak-user-docs.md
	- rm -rf $(PAK_CACHE)
	- cd providers/terraform-provider-csbpg; $(MAKE) clean

$(PAK_CACHE):
	@echo "Folder $(PAK_CACHE) does not exist. Creating it..."
	mkdir -p $@
	
.PHONY: latest-csb
latest-csb: ## point to the very latest CSB on GitHub
	$(GO) get -d github.com/cloudfoundry/cloud-service-broker@main
	$(GO) mod tidy

.PHONY: local-csb
local-csb: ## point to a local CSB repo
	echo "replace \"github.com/cloudfoundry/cloud-service-broker\" => \"$$PWD/../cloud-service-broker\"" >>go.mod
	$(GO) mod tidy

.PHONY: lint
lint: checkformat checkimports vet staticcheck ## Checks format, imports and vet

checkformat: ## Checks that the code is formatted correctly
	@@if [ -n "$$(${GOFMT} -s -e -l -d .)" ]; then       \
		echo "gofmt check failed: run 'make format'"; \
		exit 1;                                       \
	fi

checkimports: ## Checks that imports are formatted correctly
	@@if [ -n "$$(${GO} run golang.org/x/tools/cmd/goimports -l -d .)" ]; then \
		echo "goimports check failed: run 'make format'";                      \
		exit 1;                                                                \
	fi

vet: ## Runs go vet
	${GO} vet ./...

staticcheck: ## Runs staticcheck
	${GO} run honnef.co/go/tools/cmd/staticcheck ./...

.PHONY: format
format: ## format the source
	${GOFMT} -s -e -l -w .
	${GO} run golang.org/x/tools/cmd/goimports -l -w .

./providers/terraform-provider-csbpg/cloudfoundry.org/cloud-service-broker/csbpg:
	cd providers/terraform-provider-csbpg; $(MAKE) build