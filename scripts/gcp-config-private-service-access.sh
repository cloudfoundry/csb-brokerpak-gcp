#!/usr/bin/env bash

# derived from https://cloud.google.com/sql/docs/mysql/create-instance

set -o pipefail
set -o nounset
#set -e

if [ $# -lt 2 ]; then
    echo "Usage: ${0} <project> <network>"
    exit 1
fi

PROJECT=${1}; shift
VPC_NETWORK_NAME=${1}; shift

gcloud compute addresses create google-managed-services-${VPC_NETWORK_NAME} \
    --global \
    --purpose=VPC_PEERING \
    --prefix-length=23 \
    --network=${VPC_NETWORK_NAME} \
    --project=${PROJECT}

gcloud services vpc-peerings connect \
    --service=servicenetworking.googleapis.com \
    --ranges=google-managed-services-${VPC_NETWORK_NAME} \
    --network=${VPC_NETWORK_NAME} \
    --project=${PROJECT} || \
gcloud services vpc-peerings update \
    --service=servicenetworking.googleapis.com \
    --ranges=google-managed-services-${VPC_NETWORK_NAME} \
    --network=${VPC_NETWORK_NAME} \
    --project=${PROJECT} \
    --force    

