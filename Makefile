###### Help ###################################################################
.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Setup ##################################################################
IAAS=gcp
CSB_VERSION := $(or $(CSB_VERSION), $(shell grep 'github.com/cloudfoundry/cloud-service-broker' go.mod | grep -v replace | awk '{print $$NF}' | sed -e 's/v//'))
CSB := $(or $(CSB), cfplatformeng/csb:$(CSB_VERSION))
GO_OK := $(shell which go 1>/dev/null 2>/dev/null; echo $$?)
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)
ifeq ($(GO_OK), 0)
GO=go
BUILDER=go run github.com/cloudfoundry/cloud-service-broker
LDFLAGS="-X github.com/cloudfoundry/cloud-service-broker/utils.Version=$(CSB_VERSION)"
GET_CSB="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) github.com/cloudfoundry/cloud-service-broker"
else ifeq ($(DOCKER_OK), 0)
DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak --network=host
GO=docker run $(DOCKER_OPTS) golang:$(GOVERSION) go
BUILDER=docker run $(DOCKER_OPTS) $(CSB)
GET_CSB="wget -O cloud-service-broker https://github.com/cloudfoundry/cloud-service-broker/releases/download/v$(CSB_VERSION)/cloud-service-broker.linux && chmod +x cloud-service-broker"
else
$(error either Go or Docker must be installed)
endif

###### Targets ################################################################

.PHONY: build
build: $(IAAS)-services-*.brokerpak ## build brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*/*.tf
	$(BUILDER) pak build

SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)
PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 10000)

export GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS = [{"name":"small","id":"85b27a04-8695-11ea-818a-274131861b81","description":"PostgreSQL v11, shared CPU, minumum 0.6GB ram, 10GB storage","display_name":"small","cores":0.6,"postgres_version":"POSTGRES_11","storage_gb":10},{"name":"medium","id":"b41ee300-8695-11ea-87df-cfcb8aecf3bc","description":"PostgreSQL v11, shared CPU, minumum 1.7GB ram, 20GB storage","display_name":"medium","cores":1.7,"postgres_version":"POSTGRES_11","storage_gb":20},{"name":"large","id":"2a57527e-b025-11ea-b643-bf3bcf6d055a","description":"PostgreSQL v11, minumum 8 cores, minumum 8GB ram, 50GB storage","display_name":"large","cores":8,"postgres_version":"POSTGRES_11","storage_gb":50}]

.PHONY: run
run: build google_credentials google_project ## start CSB in a docker container
	docker run $(DOCKER_OPTS) \
	-p 8080:8080 \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e GOOGLE_CREDENTIALS \
	-e GOOGLE_PROJECT \
	-e "DB_TYPE=sqlite3" \
	-e "DB_PATH=/tmp/csb-db" \
	-e GSB_PROVISION_DEFAULTS \
	-e GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS \
	$(CSB) serve

.PHONY: docs
docs: build brokerpak-user-docs.md ## build docs

brokerpak-user-docs.md: *.yml
	docker run $(DOCKER_OPTS) \
	$(CSB) pak docs /brokerpak/$(shell ls *.brokerpak) > $@

.PHONY: examples
examples: ## display available examples
	docker run $(DOCKER_OPTS) \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e USER \
	$(CSB) client examples

.PHONY: run-examples
run-examples: ## run examples against CSB on localhost (run "make run" to start it), set service_name and example_name to run specific example
	docker run $(DOCKER_OPTS) \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e USER \
	$(CSB) client run-examples --service-name="$(service_name)" --example-name="$(example_name)" -j $(PARALLEL_JOB_COUNT)

.PHONY: run-integration-tests
run-integration-tests: latest-csb
	cd ./integration-tests && go run github.com/onsi/ginkgo/v2/ginkgo -r .

.PHONY: info
info: build ## use the CSB to parse the buildpak and print out contents and versions
	docker run $(DOCKER_OPTS) \
	$(CSB) pak info /brokerpak/$(shell ls *.brokerpak)

.PHONY: validate
validate: build ## use the CSB to validate the buildpak
	docker run $(DOCKER_OPTS) \
	$(CSB) pak validate /brokerpak/$(shell ls *.brokerpak)

# fetching bits for cf push broker
cloud-service-broker: go.mod ## build or fetch CSB binary
	$(shell "$(GET_CSB)")

APP_NAME := $(or $(APP_NAME), cloud-service-broker-gcp)
DB_TLS := $(or $(DB_TLS), skip-verify)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"authorized_network": "$(GCP_PAS_NETWORK)"})

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

.PHONY: latest-csb
latest-csb: ## point to the very latest CSB on GitHub
	$(GO) get -d github.com/cloudfoundry/cloud-service-broker@main
	$(GO) mod tidy

.PHONY: local-csb
local-csb: ## point to a local CSB repo
	echo "replace \"github.com/cloudfoundry/cloud-service-broker\" => \"$$PWD/../cloud-service-broker\"" >>go.mod
	$(GO) mod tidy