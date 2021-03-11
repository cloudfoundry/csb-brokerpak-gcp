###### Help ###################################################################

.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Targets ################################################################

IAAS=gcp
DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak --network=host
CSB := $(or $(CSB), cfplatformeng/csb)

.PHONY: build
build: $(IAAS)-services-*.brokerpak 

$(IAAS)-services-*.brokerpak: *.yml terraform/*.tf
	docker run $(DOCKER_OPTS) $(CSB) pak build

SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)
PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 1000)

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

.PHONY: info
info: build ## use the CSB to parse the buildpak and print out contents and versions
	docker run $(DOCKER_OPTS) \
	$(CSB) pak info /brokerpak/$(shell ls *.brokerpak)

.PHONY: validate
validate: build ## use the CSB to validate the buildpak
	docker run $(DOCKER_OPTS) \
	$(CSB) pak validate /brokerpak/$(shell ls *.brokerpak)

# fetching bits for cf push broker
cloud-service-broker: ## fetch CSB latest release from GitHub
	wget $(shell curl -sL https://api.github.com/repos/cloudfoundry-incubator/cloud-service-broker/releases/latest | jq -r '.assets[] | select(.name == "cloud-service-broker.linux") | .browser_download_url')
	mv ./cloud-service-broker.linux ./cloud-service-broker
	chmod +x ./cloud-service-broker


APP_NAME := $(or $(APP_NAME), cloud-service-broker-gcp)
DB_TLS := $(or $(DB_TLS), skip-verify)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"authorized_network": "$(GCP_PAS_NETWORK)"})

.PHONY: push-broker
push-broker: cloud-service-broker build google_credentials google_project gcp_pas_network ## push the broker to targetted Cloud Foundry
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

.PHONY: 

.PHONY: clean
clean: ## clean up build artifacts
	- rm $(IAAS)-services-*.brokerpak
	- rm ./cloud-service-broker
	- rm ./brokerpak-user-docs.md
