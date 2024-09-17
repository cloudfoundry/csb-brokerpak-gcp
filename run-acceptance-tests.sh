#!/usr/bin/env bash

set -ex

ORG=${ORG:-pivotal}
SPACE=${SPACE:-broker-cf-test}

ENVIRONMENT_LOCK_METADATA=${ENVIRONMENT_LOCK_METADATA:-environment/metadata}
ENV_NAME=$(jq .name -r < "${ENVIRONMENT_LOCK_METADATA}")

# shellcheck disable=1090
source <(smith om -l "${ENVIRONMENT_LOCK_METADATA}")
# shellcheck disable=1090
source <(smith bosh -l "${ENVIRONMENT_LOCK_METADATA}")
smith -l "${ENVIRONMENT_LOCK_METADATA}" cf-login <<< "${ORG}" &> /dev/null
DB_PASSWORD=$(echo "${ENV_NAME}" | sha256sum | cut -f1 -d' ')
export DB_PASSWORD
ENCRYPTION_PASSWORDS='[{"password": {"secret":"'${DB_PASSWORD}'"},"label":"first-encryption","primary":true}]'
if ! cf service-key  csb-sql csb-sql; then
  cf create-service-key csb-sql csb-sql
fi
CSB_DB_DATA_RAW=$( cf service-key  csb-sql csb-sql | tail -n+2)
CSB_DB_DATA=$( \
jq ".credentials | 
    {
      host: .hostname, 
      encryption: { enabled: true, passwords: $ENCRYPTION_PASSWORDS }, 
      ca: { cert: .tls.cert.ca}, 
      name: \"service_instance_db\",
      user: .username, 
      password: .password,
      port: .port
    }" <<< "${CSB_DB_DATA_RAW}"  
)
PRODS="$(om -k curl -s -p /api/v0/staged/products)"
CF_DEPLOYMENT_ID="$(echo "$PRODS" | jq -r '.[] | select(.type == "cf") | .guid')"
UAA_CREDS="$(om -k curl -s -p "/api/v0/deployed/products/$CF_DEPLOYMENT_ID/credentials/.uaa.credhub_admin_client_client_credentials")"

CH_UAA_CLIENT_NAME="$(echo "${UAA_CREDS}" | jq -r .credential.value.identity)"
CH_UAA_CLIENT_SECRET="$(echo "${UAA_CREDS}" | jq -r .credential.value.password)"
CH_UAA_URL="https://uaa.service.cf.internal:8443"
CH_CRED_HUB_URL="https://credhub.service.cf.internal:8844"
CF_API_PASS=$( credhub get --key password -n "/opsmgr/$CF_DEPLOYMENT_ID/uaa/admin_credentials" -j )
CF_API_DOMAIN="$(cf api | head -1 | cut -f3- -d/ )"

GCP_PAS_NETWORK="$(jq -r .service_network_name "${ENVIRONMENT_LOCK_METADATA}")"
export GCP_PAS_NETWORK
GOOGLE_PROJECT="$(jq -r .project "${ENVIRONMENT_LOCK_METADATA}")"
export GOOGLE_PROJECT
GOOGLE_CREDENTIALS="$(vault kv get -field key  /concourse/tas-services-enablement/gcp_cloud_service_broker)"
export GOOGLE_CREDENTIALS
GCP_PAS_NETWORK_ID="https://www.googleapis.com/compute/v1/projects/$GOOGLE_PROJECT/global/networks/$GCP_PAS_NETWORK"
GSB_PROVISION_DEFAULTS="{\"authorized_network_id\":\"${GCP_PAS_NETWORK_ID}\"}"
GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS='[{"name":"small","id":"5b45de36-cb90-11ec-a755-77f8be95a49d","description":"PostgreSQL with default version, shared CPU, minimum 0.6GB ram, 10GB storage","metadata":{"displayName":"small"},"tier":"db-f1-micro","storage_gb":10},{"name":"medium","id":"a3359fa6-cb90-11ec-bcb6-cb68544eda78","description":"PostgreSQL with default version, shared CPU, minimum 1.7GB ram, 20GB storage","metadata":{"displayName":"medium"},"tier":"db-g1-small","storage_gb":20},{"name":"db-custom-2-7680","id":"5f9a82f3-8b0a-4bbd-9a00-7ba9a3c0098d","postgres_version":"POSTGRES_15","description":"PostgreSQL 15, 2 CPU, 8GB ram, 100GB storage","metadata":{"displayName":"db-custom-2-7680"},"tier":"db-custom-4-15360","storage_gb":100}]'
GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS='[{"name":"default","id":"eec62c9b-b25e-4e65-bad5-6b74d90274bf","description":"Default MySQL v8.0 10GB storage","metadata":{"displayName":"default"},"mysql_version":"MYSQL_8_0","storage_gb":10,"tier":"db-n1-standard-2"}]'
GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS='[{"name":"default", "id":"2875f0f0-a69f-4fe6-a5ec-5ed7f6e89a01", "region": "us", "description":"Cloud Storage Bucket service with default configuration","metadata":{"display_name":"default"}}]'
# We keep the old Redis plans to be able to execute two examples (see service definition):
#  * Cloud Memorystore for Redis service with no failover.
#  * Cloud Memorystore for Redis service with automatic failover.
GSB_SERVICE_CSB_GOOGLE_REDIS_PLANS='[{"name":"basic","id":"2ed1d5b7-b21b-41ec-8da4-397a6b124484","description":"Cloud Memorystore for Redis service with no failover","metadata":{"display_name":"basic"},"service_tier": "BASIC"},{"name":"ha","id":"366b6758-11c0-4892-9aa1-6e4b7bd8b974","description":"Cloud Memorystore for Redis service with automatic failover","metadata":{"display_name":"ha"},"service_tier": "STANDARD_HA"}]'
#GSB_SERVICE_CSB_GOOGLE_PUBSUB_PLANS='[{"name":"default","id":"0690bcd8-e29e-4317-9387-f8152501403d","description":"PubSub service with topic ans subscription","metadata":{"display_name":"default"}}]'
GSB_BROKERPAK_CONFIG='{"global_labels":[{"key":"key1","value":"value1"},{"key":"key2","value":"value2"}]}'


cat << EOF > acceptance-tests/assets/vars.yml
env_name: $ENV_NAME
azs:
- us-central1-c

gcp:
  credentials: '${GOOGLE_CREDENTIALS}'
  project: '${GOOGLE_PROJECT}'

cf_api_url: ${CF_API_DOMAIN}
cf_admin_pass: "${CF_API_PASS}"
brokerpak:
  builtin:
    path: ./
  config: '${GSB_BROKERPAK_CONFIG}'
  sources: |
    {}
  terraform:
    upgrades:
      enabled: true
  updates:
    enabled: true
compatibility:
  enable-beta-services: true
credhub:
  ca_cert_file: credhub_ca_cert.pem
  skip_ssl_validation: false
  uaa_client_name: ${CH_UAA_CLIENT_NAME}
  uaa_client_secret: ${CH_UAA_CLIENT_SECRET}
  uaa_url: ${CH_UAA_URL}
  url: ${CH_CRED_HUB_URL}
db: ${CSB_DB_DATA}

provision:
  defaults: |
    ${GSB_PROVISION_DEFAULTS}
request:
  property:
    validation:
      disabled: false
service:
  csb-google-postgres:
    plans: |+
      ${GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS} 
    provision:
      defaults: |+
        ${GSB_PROVISION_DEFAULTS}
  csb-google-mysql:
    plans: |+
      ${GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS}
    provision:
      defaults: |+
        ${GSB_PROVISION_DEFAULTS}
  csb-google-storage-bucket:
    plans: |+
      ${GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS}
    provision:
      defaults: |+
        ${GSB_PROVISION_DEFAULTS}
  csb-google-redis:
    plans: |+
      ${GSB_SERVICE_CSB_GOOGLE_REDIS_PLANS}
    provision:
      defaults: |+
        ${GSB_PROVISION_DEFAULTS}
    
EOF

# This Manifest is used to deploy the broker in the upgrade tests when upgrading to a VM.
# Do not confuse with the deployment that is used to run the acceptance tests.
# The temporary manifest will be modify by the upgrade tests by using opsfiles.
# We need to have a encryption block with the default label and the same value as
# the one used when creating the broker in the initial phase of the upgrade tests.
# CSB will perform a integrity check of the password and will fail if the password
# is not the same when upgrading to a VM.
bosh int ./acceptance-tests/assets/manifest.yml  \
  -l ./acceptance-tests/assets/vars.yml \
  -v release_repo_path="$(pwd)/../csb-gcp-release/" > /tmp/tmp-manifest.yml

# This deployment is used to run the acceptance tests
# Do not confuse with the deployment that is used to run the upgrade tests
DEPLOYMENT_NAME=cloud-service-broker-gcp
bosh -d "$DEPLOYMENT_NAME" deploy ./acceptance-tests/assets/manifest.yml  \
  -l ./acceptance-tests/assets/vars.yml \
  -v name="$DEPLOYMENT_NAME" \
  -v release_repo_path="$(pwd)/../csb-gcp-release/" \
  --no-redact -n

ginkgo --procs 4 --vv acceptance-tests/upgrade

