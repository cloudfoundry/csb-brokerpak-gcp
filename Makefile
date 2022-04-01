###### Help ###################################################################
.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Setup ##################################################################
IAAS=gcp
CSB_VERSION := $(or $(CSB_VERSION), $(shell grep 'github.com/cloudfoundry/cloud-service-broker' go.mod | grep -v replace | awk '{print $$NF}' | sed -e 's/v//'))
CSB_RELEASE_VERSION := CSB_VERSION # this doesnt work well if we did make latest-csb.
#CSB_RELEASE_VERSION := $(shell echo '0.10.1-0.20220330112451-7ce0dfa511c7' | awk -F'-' '{print $1}')
#$(info $$CSB_VERSION is [${CSB_VERSION}])
#$(info $$CSB_RELEASE_VERSION is [${CSB_RELEASE_VERSION}])

CSB_DOCKER_IMAGE := $(or $(CSB), cfplatformeng/csb:$(CSB_VERSION))
GO_OK :=  $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)

####### broker environment variables
PAK_CACHE=.pak-cache
SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)
GSB_COMPATIBILITY_ENABLE_BETA_SERVICES :=true
export GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS = [{"name":"small","id":"85b27a04-8695-11ea-818a-274131861b81","description":"PostgreSQL with default version, shared CPU, minumum 0.6GB ram, 10GB storage","display_name":"small","cores":0.6,"storage_gb":10},{"name":"medium","id":"b41ee300-8695-11ea-87df-cfcb8aecf3bc","description":"PostgreSQL with default version, shared CPU, minumum 1.7GB ram, 20GB storage","display_name":"medium","cores":1.7,"storage_gb":20},{"name":"large","id":"2a57527e-b025-11ea-b643-bf3bcf6d055a","description":"PostgreSQL with default version, minumum 8 cores, minumum 8GB ram, 50GB storage","display_name":"large","cores":8,"storage_gb":50}]
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"authorized_network": "$(GCP_PAS_NETWORK)"})

ifeq ($(GO_OK), 0) # use local go binary
GO=go
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
BROKER_DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak --network=host \
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

GO_DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak --network=host
GO=docker run $(GO_DOCKER_OPTS) golang:latest go

# this doesnt work well if we did make latest-csb. We should build it instead, with go inside a container.
GET_CSB="wget -O cloud-service-broker https://github.com/cloudfoundry/cloud-service-broker/releases/download/v$(CSB_RELEASE_VERSION)/cloud-service-broker.linux && chmod +x cloud-service-broker"
else
$(error either Go or Docker must be installed)
endif

###### Targets ################################################################

.PHONY: build
build: $(IAAS)-services-*.brokerpak ## build brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*/*.tf
	$(RUN_CSB) pak build

.pak-cache:
	mkdir -p $(PAK_CACHE)


.PHONY: run
run: build google_credentials google_project ## start CSB in a docker container
	$(RUN_CSB) serve

.PHONY: docs
docs: build brokerpak-user-docs.md ## build docs

brokerpak-user-docs.md: *.yml # TODO: unificar
	$(RUN_CSB) pak docs $(PAK_PATH)/$(shell ls *.brokerpak) > $@ # GO

.PHONY: examples
examples: ## display available examples  ## TODO these need have the broker already running if not using docker
	 $(RUN_CSB) client examples

PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 10000)

.PHONY: run-examples
run-examples: ## run examples against CSB on localhost (run "make run" to start it), set service_name and example_name to run specific example
	$(RUN_CSB) client run-examples --service-name="$(service_name)" --example-name="$(example_name)" -j $(PARALLEL_JOB_COUNT)

.PHONY: run-integration-tests
run-integration-tests: latest-csb  ## run integration tests for this brokerpak
	cd ./integration-tests && go run github.com/onsi/ginkgo/v2/ginkgo -r .

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

.PHONY: latest-csb
latest-csb: ## point to the very latest CSB on GitHub
	$(GO) get -d github.com/cloudfoundry/cloud-service-broker@main
	$(GO) mod tidy

.PHONY: local-csb
local-csb: ## point to a local CSB repo
	echo "replace \"github.com/cloudfoundry/cloud-service-broker\" => \"$$PWD/../cloud-service-broker\"" >>go.mod
	$(GO) mod tidy