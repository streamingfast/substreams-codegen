#!/usr/bin/env bash
set -euxo pipefail

LOCAL = $TEST_LOCAL_CODEGEN
listen_address=
if [[ "$LOCAL" == "true" ]]; then
      if [[ -n "${CI:-}" ]]; then
         # Codegen address when running test
         listen_address="http://172.17.0.1:9000"
      else
         listen_address="http://host.docker.internal:9000"
      fi
else
    listen_address= https://codegen-staging.substreams.dev
fi

substreams init --state-file /app/generator.json --force-download-cwd --codegen-endpoint $listen_address

substreams build

## To validate the manifest
substreams info