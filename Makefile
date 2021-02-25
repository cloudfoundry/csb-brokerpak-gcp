
IAAS=gcp
DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak --network=host
CSB=cfplatformeng/csb

.PHONY: build
build: $(IAAS)-services-*.brokerpak 

$(IAAS)-services-*.brokerpak: *.yml terraform/*.tf
	docker run $(DOCKER_OPTS) $(CSB) pak build

SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)
PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 2)

.PHONY: run
run: build google_credentials google_project
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
docs: build brokerpak-user-docs.md

brokerpak-user-docs.md: *.yml
	docker run $(DOCKER_OPTS) \
	$(CSB) pak docs /brokerpak/$(shell ls *.brokerpak) > $@

.PHONY: run-examples
run-examples: 
	docker run $(DOCKER_OPTS) \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e USER \
	$(CSB) client run-examples -j $(PARALLEL_JOB_COUNT)

.PHONY: info
info: build
	docker run $(DOCKER_OPTS) \
	$(CSB) pak info /brokerpak/$(shell ls *.brokerpak)

.PHONY: validate
validate: build
	docker run $(DOCKER_OPTS) \
	$(CSB) pak validate /brokerpak/$(shell ls *.brokerpak)

# fetching bits for cf push broker
cloud-service-broker:
	wget $(shell curl -sL https://api.github.com/repos/cloudfoundry-incubator/cloud-service-broker/releases/latest | jq -r '.assets[] | select(.name == "cloud-service-broker.linux") | .browser_download_url')
	mv ./cloud-service-broker.linux ./cloud-service-broker
	chmod +x ./cloud-service-broker


APP_NAME := $(or $(APP_NAME), cloud-service-broker-gcp)
DB_TLS := $(or $(DB_TLS), skip-verify)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"authorized_network": "$(GCP_PAS_NETWORK)"})

.PHONY: push-broker
push-broker: cloud-service-broker build google_credentials google_project gcp_pas_network
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
clean:
	- rm $(IAAS)-services-*.brokerpak
	- rm ./cloud-service-broker
	- rm ./brokerpak-user-docs.md
