##########################################################
# The actual test. It runs inside a Cloudfoundry APP.
##########################################################
#!/usr/bin/env bash
set -euxo pipefail

SERV_NAME="csb-google-postgres"
PLAN_NAME="small"
SERV_INSTANCE="test-postgres-bindings-$(echo $RANDOM | md5sum | head -c 20; echo;)"
BINDING1="binding1-$(echo $RANDOM | md5sum | head -c 20; echo;)"
BINDING2="binding2-$(echo $RANDOM | md5sum | head -c 20; echo;)"

function cleanup {
  set +e
  # set +e ensures all clean commands will get executed (doesn't exit on errors)
  # set +e ensures subsequent commands don't count for determining the exit code
  cf delete-service-key "${SERV_INSTANCE}" "${BINDING2}" --wait -f
  cf delete-service-key "${SERV_INSTANCE}" "${BINDING1}" --wait -f
  cf delete-service "${SERV_INSTANCE}" --wait -f
}
trap 'cleanup' EXIT


# The following instruction picks org `1. pivotal` when `cf login` asks for it
echo "1" | /cf-login.bash

cf create-service --wait "${SERV_NAME}" "${PLAN_NAME}" "${SERV_INSTANCE}"
cf create-service-key "${SERV_INSTANCE}" "${BINDING1}" --wait

BINDING1_GUID="$(cf service-key "${SERV_INSTANCE}" "${BINDING1}" --guid)"
cf curl "/v2/service_keys/${BINDING1_GUID}" > "/tmp/${BINDING1_GUID}.json"

# https://www.postgresql.org/docs/current/libpq-envars.html
export PGHOST="$(    jq -r '.entity.credentials.hostname' "/tmp/${BINDING1_GUID}.json")"
export PGDATABASE="$(jq -r '.entity.credentials.name'     "/tmp/${BINDING1_GUID}.json")"
export PGUSER="$(    jq -r '.entity.credentials.username' "/tmp/${BINDING1_GUID}.json")"
export PGPASSWORD="$(jq -r '.entity.credentials.password' "/tmp/${BINDING1_GUID}.json")"
export PGSSLMODE="prefer"
export PGSSLKEY="/tmp/sslkey"
export PGSSLCERT="/tmp/sslcert"
export PGSSLROOTCERT="/tmp/sslrootcert"
jq -r '.entity.credentials.sslcert'     "/tmp/${BINDING1_GUID}.json" > "${PGSSLCERT}"     && chmod 0600 "${PGSSLCERT}"
jq -r '.entity.credentials.sslkey'      "/tmp/${BINDING1_GUID}.json" > "${PGSSLKEY}"      && chmod 0600 "${PGSSLKEY}"
jq -r '.entity.credentials.sslrootcert' "/tmp/${BINDING1_GUID}.json" > "${PGSSLROOTCERT}" && chmod 0600 "${PGSSLROOTCERT}"

psql <<EOF
\conninfo
CREATE TABLE PUBLIC.TEST_POSTGRES_BINDINGS();
EOF

cf create-service-key "${SERV_INSTANCE}" "${BINDING2}" --wait
