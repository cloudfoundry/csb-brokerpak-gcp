#!/usr/bin/env bash

set -o pipefail
set -o nounset


if [ $# -lt 1 ]; then
    echo "Usage: ${0} <smith lock file name>"
    exit 1
fi

PASSWORD=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

PROJECT=$(cat ${1} | jq -r .project)
REGION=$(cat ${1} | jq -r .region)
VPC_NETWORK_NAME=$(cat ${1} | jq -r .service_network_name)

INSTANCE_NAME="$(cat ${1} | jq -r .name)-csb-db-$$"
TIER=db-n1-standard-1

NETWORK="https://www.googleapis.com/compute/alpha/projects/${PROJECT}/global/networks/${VPC_NETWORK_NAME}"; shift

if [ $# -gt 0 ]; then
    TIER=${1}; shift
fi

DB_NAME=csb-db

gcloud compute addresses create google-managed-services-mysql-${VPC_NETWORK_NAME} \
    --global \
    --purpose=VPC_PEERING \
    --prefix-length=24 \
    --network=${VPC_NETWORK_NAME} \
    --project=${PROJECT}  || echo "vpc peering exists..."

gcloud services vpc-peerings connect \
    --service=servicenetworking.googleapis.com \
    --ranges=google-managed-services-mysql-${VPC_NETWORK_NAME} \
    --network=${VPC_NETWORK_NAME} \
    --project=${PROJECT} || echo "service peering exists..."

gcloud beta sql instances create ${INSTANCE_NAME} --network $NETWORK --tier ${TIER} --region ${REGION}  --no-assign-ip --labels owner=csb

gcloud sql users set-password root --host=% --instance ${INSTANCE_NAME} --password ${PASSWORD} || echo "Failed to add account"

gcloud sql databases create ${DB_NAME} --instance=${INSTANCE_NAME} || echo "failed to create database"

IP_ADDRESS=$(gcloud sql instances describe ${INSTANCE_NAME} --format="json" | jq -r .ipAddresses[0].ipAddress) || echo "failed to get addres"

${SCRIPT_DIR}/cf-create-mysql-cups.sh "${IP_ADDRESS}" root "${PASSWORD}" "${DB_NAME}" || echo "failed to create cups"ech