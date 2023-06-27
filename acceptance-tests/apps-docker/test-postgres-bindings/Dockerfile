# syntax=docker/dockerfile
FROM postgres:14
RUN apt-get update && apt-get install -y --no-install-recommends jq
COPY --from=cflondonservices/csb-ci-main /usr/local/bin/om /usr/bin/om
COPY --from=cflondonservices/csb-ci-main /usr/local/bin/cf /usr/bin/cf
COPY cf-login.bash /cf-login.bash
COPY run-test.bash /run-test.bash

ENTRYPOINT ["/run-test"]
