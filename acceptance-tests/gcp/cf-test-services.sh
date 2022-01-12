#!/usr/bin/env bash

set -e
set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=0

${SCRIPT_DIR}/cf-test-stack-driver.sh && ${SCRIPT_DIR}/cf-test-dataproc.sh
RESULT=$?

wait

if [ ${RESULT} -eq 0 ]; then
  echo "$0 SUCCEEDED"
else
  echo "$0 FAILED"
fi

exit ${RESULT}
