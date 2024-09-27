#!/usr/bin/env bash


set +x # Hide secrets
set -o errexit
set -o pipefail
set -e

# Change to the directory where the script is located
cd "$(dirname "$0")"

# Declare variables
DEVELOPMENT_BUILD_DIR="../"
CSB_GCP_RELEASE_DIR="../../csb-gcp-release"
CLOUD_SERVICE_BROKER_DIR="../../cloud-service-broker"
TMP_RELEASE_PATH="/tmp/csb-gcp-release"
FIXED_ENCRYPT="02630426-1d06-47b0-b712-5c74dd4f8182"

# Convert relative paths to absolute paths
DEVELOPMENT_BUILD_DIR=$(realpath "$DEVELOPMENT_BUILD_DIR")
CSB_GCP_RELEASE_DIR=$(realpath "$CSB_GCP_RELEASE_DIR")
CLOUD_SERVICE_BROKER_DIR=$(realpath "$CLOUD_SERVICE_BROKER_DIR")

# Run local release modifier
echo "Running local release modifier - vendoring the brokerpak, cloud-service-broker and iaas release - destination $TMP_RELEASE_PATH"
go run -C "$DEVELOPMENT_BUILD_DIR/acceptance-tests/boshifier/app/vendirlocalrelease" . \
  -brokerpak-path "$DEVELOPMENT_BUILD_DIR" \
  -cloud-service-broker-path "$CLOUD_SERVICE_BROKER_DIR" \
  -iaas-release-path "$CSB_GCP_RELEASE_DIR" \
  -tmp-release-path "$TMP_RELEASE_PATH"

# Run manifest creator
# Note: Use a fixed DB name if you want to avoid creating a new schema every time
DB_NAME=$(uuidgen | tr -d '-' | cut -c1-20)
go run -C "$DEVELOPMENT_BUILD_DIR/acceptance-tests/boshifier/app/manifestcreator" . \
  -brokerpak-path "$DEVELOPMENT_BUILD_DIR" \
  -iaas-release-path "$TMP_RELEASE_PATH" \
  -db-name "$DB_NAME" \
  -db-secret "$FIXED_ENCRYPT"

# Run deployer
go run -C "$DEVELOPMENT_BUILD_DIR/acceptance-tests/boshifier/app/deployer" . \
  -brokerpak-path "$DEVELOPMENT_BUILD_DIR" \
  -iaas-release-path "$TMP_RELEASE_PATH" \
  -bosh-deployment-name "cloud-service-broker-gcp"

