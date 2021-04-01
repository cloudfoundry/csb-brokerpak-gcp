#!/usr/bin/env bash

set -o pipefail
set -o nounset
set -e

if [ $# -lt 4 ]; then
    echo "Usage: ${0} <hostname> <username> <password> <db_name>"
    exit 1
fi

HOSTNAME=${1}; shift
USERNAME=${1}; shift
PASSWORD=${1}; shift
DB_NAME=${1}; shift
PORT=3306

cf cups csb-sql -p "{\"hostname\":\"${HOSTNAME}\", \
                     \"username\":\"${USERNAME}\", \
                     \"password\":\"${PASSWORD}\", \
                     \"name\":\"${DB_NAME}\", \
                     \"port\":${PORT}, \
                     \"uri\":\"mysql://${USERNAME}:${PASSWORD}@${HOSTNAME}:${PORT}/${DB_NAME}\"}" \
                -t mysql