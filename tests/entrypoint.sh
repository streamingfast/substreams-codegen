#!/usr/bin/env bash
set -euxo pipefail

listen_address=
if [[ ${TEST_STAGING:-false} == "true" ]]; then
    listen_address="https://codegen-staging.substreams.dev"
else
      if [[ -z ${CI} ]]; then
        listen_address="http://host.docker.internal:9000"
      else
         # Codegen address when running test
         listen_address="http://172.17.0.1:9000"
      fi
fi

SUBSTREAMS_CODEGEN_ENDPOINT=$listen_address substreams init --state-file /app/generator.json --force-download-cwd

substreams build

## To validate the manifest
substreams info
