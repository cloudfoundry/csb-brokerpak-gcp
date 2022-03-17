#!/usr/bin/env bash

set -o errexit
set -o pipefail

if [ $# -lt 1 ]; then
    echo "Usage: ${0} <service name>"
    exit 1
fi

SERVICE_NAME=${1}; shift

SECURITY_USER_NAME=${SECURITY_USER_NAME:=aws-broker}
SECURITY_USER_PASSWORD=${SECURITY_USER_PASSWORD:=aws-broker-pw}
DOCKER_OPTS="--rm -v $(PWD):/brokerpak -w /brokerpak --network=host"
CSB=cfplatformeng/csb
GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=true

docker run ${DOCKER_OPTS} \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e USER \
	${CSB} client run-examples --service-name ${SERVICE_NAME}

