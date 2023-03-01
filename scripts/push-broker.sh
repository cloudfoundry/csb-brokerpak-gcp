#!/usr/bin/env bash

set +x # Hide secrets
set -o errexit
set -o pipefail
set -e

if [[ -z ${MANIFEST} ]]; then
  MANIFEST=manifest.yml
fi

if [[ -z ${APP_NAME} ]]; then
  APP_NAME=cloud-service-broker
fi

if [[ -z ${SECURITY_USER_NAME} ]]; then
  echo "Missing SECURITY_USER_NAME variable"
  exit 1
fi

if [[ -z ${SECURITY_USER_PASSWORD} ]]; then
  echo "Missing SECURITY_USER_PASSWORD variable"
  exit 1
fi

cfmf="/tmp/cf-manifest.$$.yml"
touch "$cfmf"
trap "rm -f $cfmf" EXIT
chmod 600 "$cfmf"
cat "$MANIFEST" >$cfmf

echo "  env:" >>$cfmf
echo "    SECURITY_USER_PASSWORD: ${SECURITY_USER_PASSWORD}" >>$cfmf
echo "    SECURITY_USER_NAME: ${SECURITY_USER_NAME}" >>$cfmf
echo "    TERRAFORM_UPGRADES_ENABLED: ${TERRAFORM_UPGRADES_ENABLED:-true}" >>$cfmf
echo "    BROKERPAK_UPDATES_ENABLED: ${BROKERPAK_UPDATES_ENABLED:-true}" >>$cfmf
echo "    GSB_COMPATIBILITY_ENABLE_BETA_SERVICES: ${GSB_COMPATIBILITY_ENABLE_BETA_SERVICES:-true}" >>$cfmf

if [[ ${GSB_PROVISION_DEFAULTS} ]]; then
  echo "    GSB_PROVISION_DEFAULTS: $(echo "$GSB_PROVISION_DEFAULTS" | jq @json)" >>$cfmf
fi

if [[ ${GOOGLE_CREDENTIALS} ]]; then
  echo "    GOOGLE_CREDENTIALS: $(echo "$GOOGLE_CREDENTIALS" | jq @json)" >>$cfmf
fi

if [[ ${GOOGLE_PROJECT} ]]; then
  echo "    GOOGLE_PROJECT: ${GOOGLE_PROJECT}" >>$cfmf
fi

if [[ ${GSB_BROKERPAK_BUILTIN_PATH} ]]; then
  echo "    GSB_BROKERPAK_BUILTIN_PATH: ${GSB_BROKERPAK_BUILTIN_PATH}" >>$cfmf
fi

if [[ ${CH_CRED_HUB_URL} ]]; then
  echo "    CH_CRED_HUB_URL: ${CH_CRED_HUB_URL}" >>$cfmf
fi

if [[ ${CH_UAA_URL} ]]; then
  echo "    CH_UAA_URL: ${CH_UAA_URL}" >>$cfmf
fi

if [[ ${CH_UAA_CLIENT_NAME} ]]; then
  echo "    CH_UAA_CLIENT_NAME: ${CH_UAA_CLIENT_NAME}" >>$cfmf
fi

if [[ ${CH_UAA_CLIENT_SECRET} ]]; then
  echo "    CH_UAA_CLIENT_SECRET: ${CH_UAA_CLIENT_SECRET}" >>$cfmf
fi

if [[ ${CH_SKIP_SSL_VALIDATION} ]]; then
  echo "    CH_SKIP_SSL_VALIDATION: ${CH_SKIP_SSL_VALIDATION}" >>$cfmf
fi

if [[ ${DB_TLS} ]]; then
  echo "    DB_TLS: ${DB_TLS}" >>$cfmf
fi

if [[ ${GSB_DEBUG} ]]; then
  echo "    GSB_DEBUG: ${GSB_DEBUG}" >>$cfmf
fi

if [[ -z "$GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS varaible"
  exit 1
fi
echo "    GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS: $(echo "$GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS: $(echo "$GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS: $(echo "$GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_GOOGLE_REDIS_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_GOOGLE_REDIS_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_GOOGLE_REDIS_PLANS: $(echo "$GSB_SERVICE_CSB_GOOGLE_REDIS_PLANS" | jq @json)" >>$cfmf

cf push --no-start -f "${cfmf}" --var app=${APP_NAME}

if [[ -z ${MSYQL_INSTANCE} ]]; then
  MSYQL_INSTANCE=csb-sql
fi

cf bind-service "${APP_NAME}" "${MSYQL_INSTANCE}"

if ! cf start "${APP_NAME}"
then
	cf logs "${APP_NAME}" --recent
	exit 1
fi

if [[ -z ${BROKER_NAME} ]]; then
  BROKER_NAME=csb-$USER
fi

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(LANG=EN cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(LANG=EN cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
