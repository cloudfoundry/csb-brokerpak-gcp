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

cf push --no-start -f "${MANIFEST}" --var app=${APP_NAME}

if [[ -z ${SECURITY_USER_NAME} ]]; then
  echo "Missing SECURITY_USER_NAME variable"
  exit 1
fi

if [[ -z ${SECURITY_USER_PASSWORD} ]]; then
  echo "Missing SECURITY_USER_PASSWORD variable"
  exit 1
fi

cf set-env "${APP_NAME}" SECURITY_USER_PASSWORD "${SECURITY_USER_PASSWORD}"
cf set-env "${APP_NAME}" SECURITY_USER_NAME "${SECURITY_USER_NAME}"
cf set-env "${APP_NAME}" BROKERPAK_UPDATES_ENABLED true

if [[ ${GSB_PROVISION_DEFAULTS} ]]; then
  cf set-env "${APP_NAME}" GSB_PROVISION_DEFAULTS "${GSB_PROVISION_DEFAULTS}"
fi

if [[ ${GOOGLE_CREDENTIALS} ]]; then
  cf set-env "${APP_NAME}" GOOGLE_CREDENTIALS "${GOOGLE_CREDENTIALS}"
fi

if [[ ${GOOGLE_PROJECT} ]]; then
  cf set-env "${APP_NAME}" GOOGLE_PROJECT "${GOOGLE_PROJECT}"
fi

if [[ ${ARM_SUBSCRIPTION_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_SUBSCRIPTION_ID "${ARM_SUBSCRIPTION_ID}"
fi

if [[ ${ARM_TENANT_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_TENANT_ID "${ARM_TENANT_ID}"
fi

if [[ ${ARM_CLIENT_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_CLIENT_ID "${ARM_CLIENT_ID}"
fi

if [[ ${ARM_CLIENT_SECRET} ]]; then
  cf set-env "${APP_NAME}" ARM_CLIENT_SECRET "${ARM_CLIENT_SECRET}"
fi

if [[ ${AWS_ACCESS_KEY_ID} ]]; then
  cf set-env "${APP_NAME}" AWS_ACCESS_KEY_ID "${AWS_ACCESS_KEY_ID}"
fi

if [[ ${AWS_SECRET_ACCESS_KEY} ]]; then
  cf set-env "${APP_NAME}" AWS_SECRET_ACCESS_KEY "${AWS_SECRET_ACCESS_KEY}"
fi

if [[ ${GSB_BROKERPAK_BUILTIN_PATH} ]]; then
  cf set-env "${APP_NAME}" GSB_BROKERPAK_BUILTIN_PATH "${GSB_BROKERPAK_BUILTIN_PATH}"
fi

if [[ ${CH_CRED_HUB_URL} ]]; then
  cf set-env "${APP_NAME}" CH_CRED_HUB_URL "${CH_CRED_HUB_URL}"
fi

if [[ ${CH_UAA_URL} ]]; then
  cf set-env "${APP_NAME}" CH_UAA_URL "${CH_UAA_URL}"
fi

if [[ ${CH_UAA_CLIENT_NAME} ]]; then
  cf set-env "${APP_NAME}" CH_UAA_CLIENT_NAME "${CH_UAA_CLIENT_NAME}"
fi

if [[ ${CH_UAA_CLIENT_SECRET} ]]; then
  cf set-env "${APP_NAME}" CH_UAA_CLIENT_SECRET "${CH_UAA_CLIENT_SECRET}"
fi

if [[ ${CH_SKIP_SSL_VALIDATION} ]]; then
  cf set-env "${APP_NAME}" CH_SKIP_SSL_VALIDATION "${CH_SKIP_SSL_VALIDATION}"
fi

if [[ ${DB_TLS} ]]; then
  cf set-env "${APP_NAME}" DB_TLS "${DB_TLS}"
fi

if [[ ${GSB_DEBUG} ]]; then
  cf set-env "${APP_NAME}" GSB_DEBUG "${GSB_DEBUG}"
fi

if [[ -z "$GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS" ]]; then
  GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS='[{"name":"small","id":"85b27a04-8695-11ea-818a-274131861b81","description":"PostgreSQL v11, shared CPU, minumum 0.6GB ram, 10GB storage","display_name":"small","cores":0.6,"postgres_version":"POSTGRES_11","storage_gb":10},{"name":"medium","id":"b41ee300-8695-11ea-87df-cfcb8aecf3bc","description":"PostgreSQL v11, shared CPU, minumum 1.7GB ram, 20GB storage","display_name":"medium","cores":1.7,"postgres_version":"POSTGRES_11","storage_gb":20},{"name":"large","id":"2a57527e-b025-11ea-b643-bf3bcf6d055a","description":"PostgreSQL v11, minumum 8 cores, minumum 8GB ram, 50GB storage","display_name":"large","cores":8,"postgres_version":"POSTGRES_11","storage_gb":50}]'
fi
cf set-env "${APP_NAME}" GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS "${GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS}"

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

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
