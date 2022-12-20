#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -e

if [ -z "$1" ]; then
    echo "No environment argument supplied, should be one of 'gcp', 'azure' or 'aws'"
    exit 1
fi

if [[ -z ${env} ]]; then
  echo 'Missing environment variable ($env) pointing to smith environment file'
  exit 1
fi

set -o nounset

export PCF_NETWORK=$(cat $env | jq -r .service_network_name)
GCP_PROJECT=$(cat $env | jq -r .project)
PCF_NETWORK_ID="$(https://www.googleapis.com/compute/v1/projects/$GCP_PROJECT/global/networks/$PCF_NETWORK)"
make "push-broker-${1}"

SERVICE=google-mysql
PLAN=small
NAME=mysql-test
PARAMS="{\"authorized_network_id\":\"${PCF_NETWORK_ID}\"}"

cf create-service "${SERVICE}" "${PLAN}" "${NAME}" -c "${PARAMS}"

cf service "${NAME}" | grep "create in progress"

set +e
while [ $? -eq 0 ]; do
    sleep 15
    cf service "${NAME}" | grep "create in progress"
done
set -e

APP=spring-music

cf bind-service "${APP}" "${NAME}"

cf restart "${APP}"

cf unbind-service "${APP}" "${NAME}"

cf delete-service -f "${NAME}"

