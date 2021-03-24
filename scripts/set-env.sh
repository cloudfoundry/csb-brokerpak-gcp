#!/usr/bin/env bash

set +x # Hide secrets
set -e

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "You must source this script\nsource ${0}" && exit 1

export GCP_SERVICE_ACCOUNT_JSON=$(lpass show --notes "Shared-CF Platform Engineering/pks cluster management gcp service account")
export ROOT_SERVICE_ACCOUNT_JSON="${GCP_SERVICE_ACCOUNT_JSON}"
export GOOGLE_CREDENTIALS="${GCP_SERVICE_ACCOUNT_JSON}"
export GOOGLE_PROJECT=$(echo ${GOOGLE_CREDENTIALS} | jq -r .project_id)

export GCP_PAS_NETWORK=$(lpass show "Shared-CF Platform Engineering/pe-cloud-service-broker/cloud service broker pipeline secrets.yml" | grep gcp-network | cut -d ' ' -f 2)

export SECURITY_USER_NAME=brokeruser
export SECURITY_USER_PASSWORD=brokeruserpassword
export DB_HOST=localhost
export DB_USERNAME=broker
export DB_PASSWORD=brokerpass
export DB_NAME=brokerdb
export PORT=8080