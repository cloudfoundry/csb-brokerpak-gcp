#!/usr/bin/env bash

echo "Starting prepare_env.sh"

# -----------------------------------------------------------------------------------

: "${ORG:="pivotal"}"

if [[ -z ${ENVIRONMENT_LOCK_METADATA} ]]; then
  echo "ENVIRONMENT_LOCK_METADATA is not set"
fi

if [[ ! -f ${ENVIRONMENT_LOCK_METADATA} ]]; then
  echo "ENVIRONMENT_LOCK_METADATA is set to a file that does not exist"
fi

if [[ -z ${GOOGLE_CREDENTIALS} ]]; then
  echo "GOOGLE_CREDENTIALS is not set"
fi

if [[ -z ${GOOGLE_PROJECT} ]]; then
  echo "GOOGLE_PROJECT is not set"
fi

if [[ -z ${GCP_PAS_NETWORK} ]]; then
  echo "GCP_PAS_NETWORK is not set"
fi

# -----------------------------------------------------------------------------------


# Set Ops Manager variables
echo "Setting Ops Manager variables from environment lock metadata file: ${ENVIRONMENT_LOCK_METADATA}"
# shellcheck disable=1090
source <(smith om -l "${ENVIRONMENT_LOCK_METADATA}")

# Set BOSH variables
echo "Setting BOSH variables from environment lock metadata file: ${ENVIRONMENT_LOCK_METADATA}"
# shellcheck disable=1090
source <(smith bosh -l "${ENVIRONMENT_LOCK_METADATA}")
# Log in to CF
echo "Logging in to CF Org: ${ORG}"
smith -l "${ENVIRONMENT_LOCK_METADATA}" cf-login <<< "${ORG}" &> /dev/null

# Get DeploymentGUID from Ops Manager. We need to get the credentials for the CredHub client
PRODS="$(om -k curl -s -p /api/v0/staged/products)"
CF_DEPLOYMENT_GUID="$(echo "$PRODS" | jq -r '.[] | select(.type == "cf") | .guid')"
UAA_CREDS="$(om -k curl -s -p "/api/v0/deployed/products/$CF_DEPLOYMENT_GUID/credentials/.uaa.credhub_admin_client_client_credentials")"
CH_UAA_CLIENT_NAME="$(echo "${UAA_CREDS}" | jq -r .credential.value.identity)"
CH_UAA_CLIENT_SECRET="$(echo "${UAA_CREDS}" | jq -r .credential.value.password)"
CH_UAA_URL="https://uaa.service.cf.internal:8443"
CH_CRED_HUB_URL="https://credhub.service.cf.internal:8844"
CF_API_PASS=$(credhub get --key password -n "/opsmgr/$CF_DEPLOYMENT_GUID/uaa/admin_credentials" -j )
CF_API_DOMAIN="$(cf api | head -1 | cut -f3- -d/ )"
GCP_PAS_NETWORK_ID="https://www.googleapis.com/compute/v1/projects/$GOOGLE_PROJECT/global/networks/$GCP_PAS_NETWORK"
# shellcheck disable=SC2089
GSB_PROVISION_DEFAULTS="{\"authorized_network_id\":\"${GCP_PAS_NETWORK_ID}\"}"

export CF_DEPLOYMENT_GUID
export UAA_CREDS
export CH_UAA_CLIENT_NAME
export CH_UAA_CLIENT_SECRET
export CH_UAA_URL
export CH_CRED_HUB_URL
export CF_API_PASS
export CF_API_DOMAIN
# shellcheck disable=SC2090
export GSB_PROVISION_DEFAULTS