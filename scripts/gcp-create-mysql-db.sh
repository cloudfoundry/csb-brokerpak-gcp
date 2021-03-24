#!/usr/bin/env bash

# derived from https://cloud.google.com/sql/docs/mysql/create-instance

set -o pipefail
set -o nounset
set -e

if [ $# -lt 4 ]; then
    echo "Usage: ${0} <instance name> <region> <project> <network> [tier - default db-n1-standard-1]"
    exit 1
fi

INSTANCE_NAME=${1}; shift
REGION=${1}; shift
PROJECT=${1}; shift
VPC_NETWORK_NAME=${1}; shift
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
    --project=${PROJECT}

gcloud services vpc-peerings connect \
    --service=servicenetworking.googleapis.com \
    --ranges=google-managed-services-mysql-${VPC_NETWORK_NAME} \
    --network=${VPC_NETWORK_NAME} \
    --project=${PROJECT}

gcloud beta sql instances create ${INSTANCE_NAME} --network $NETWORK --tier ${TIER} --region ${REGION}  --no-assign-ip --labels owner=csb

#USERNAME=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-zA-Z' | fold -w 16 | head -n 1)
PASSWORD=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

gcloud sql users set-password root --host=% --instance ${INSTANCE_NAME} --password ${PASSWORD}

gcloud sql databases create ${DB_NAME} --instance=${INSTANCE_NAME}

IP_ADDRESS=$(gcloud sql instances describe ${INSTANCE_NAME} --format="json" | jq -r .ipAddresses[0].ipAddress)
echo Server Details
echo     IP Address: ${IP_ADDRESS}
echo Admin Username: root
echo Admin Password: ${PASSWORD}
echo  Database Name: ${DB_NAME}
