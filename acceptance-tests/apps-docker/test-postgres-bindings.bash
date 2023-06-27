##########################################################
# Entrypoint for the test. Execute it locally or in CI.
# It expects a SMITH METADATA as its first argument.
##########################################################
#!/usr/bin/env bash
set -euxo pipefail

ENVIRONMENT_LOCK_METADATA=${1?This script requires a SMITH METADATA as the first argument.}
export ENVIRONMENT_LOCK_METADATA

APPNAME="app-test-postgres-bindings-$(echo $RANDOM | md5sum | head -c 20; echo;)"

function cleanup {
  set +e
  # set +e ensures all clean commands will get executed (doesn't exit on errors)
  # set +e ensures subsequent commands don't count for determining the exit code
  cf delete "${APPNAME}" -fr
  # we are already deleting registry app immediately after using it but we add it here
  # too in case the script failed early and never reached the registry deletion command
  cf delete registry -fr
  cf disable-feature-flag diego_docker
  rm "${SCRIPT_DIR}/test-postgres-bindings/cf-login.bash"
}
trap 'cleanup' EXIT


SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
pushd "${SCRIPT_DIR}/test-postgres-bindings/"
  # The -f flag creates a bash script with all necessary information to be able to interact with CF.
  # We will be embedding this script into the Docker image which we'll use as the base for our APP,
  # Thanks to this, the APP will be able to run any CF commands and test almost any functionality.
  CFLOGIN_SCRIPT="$(smith cf-login -f | tail -n1)"
  mv "${CFLOGIN_SCRIPT}" "cf-login.bash"
  chmod +x "cf-login.bash"

  # The following instruction picks org `1. pivotal` when `cf login` asks for it
  echo "1" | ./cf-login.bash

  # Notice how two of the instructions from the Dockerfile are:
  ############################################################
  # COPY cf-login.bash /cf-login.bash
  # COPY run-test.bash /run-test.bash
  ############################################################
  # This is how `run-test.bash` ends up inside the APP filesystem and how,
  # by executing `cf-login.bash` we can run any CF commands within the APP
  ############################################################
  docker build --tag "${APPNAME}" --file Dockerfile .
popd

# Using Docker in Cloud Foundry
# https://docs.cloudfoundry.org/adminguide/docker.html
cf enable-feature-flag diego_docker

# For pushing container images as an app CF has to download them from a registry. We can't upload a local image directly.
# To workaround this we push our own registry in the CF environment we are testing. This strategy brings many advantages.
# - It doesn't require an account in a third-party service
# - No need to worry about images' lifecycle as they're built on the fly
# - We can include security sensitive binaries and files in images
# FUTURE IMPROVEMENTS
# - The registry is publicly accessible during a very small period of time, then its deleted.
#   We could try to enable basic authentication in the future, if needed.
cf push registry --docker-image registry:2 --start-command '/entrypoint.sh /etc/docker/registry/config.yml' --health-check-type process --no-manifest
REGISTRY_URL="$(cf curl /v2/apps/$(cf app registry --guid)/env | jq -rc ".application_env_json.VCAP_APPLICATION.application_uris[0]")"

# After pushing the image and the app based on it delete the registry immediately.
# This minimises the risk of publicly exposing our images. We could try to enable basic authentication in the future.
docker tag "${APPNAME}" "${REGISTRY_URL}/${APPNAME}"
docker push "${REGISTRY_URL}/${APPNAME}"
cf push "${APPNAME}" --docker-image "${REGISTRY_URL}/${APPNAME}" --start-command "/bin/bash" --health-check-type process --no-manifest --no-route
cf delete registry -fr

# The following command runs the actual script. Previous commands just prepared the environment.
# Notice that `run-tests.bash` is running inside a CF APP so it should be able
# to communicate with any service or database associated to this CF instance.
cf ssh "${APPNAME}" --command "/run-test.bash"

