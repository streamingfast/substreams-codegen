#!/usr/bin/env bash
set -euxo pipefail

if [ "$TEST_LOCAL_CODEGEN" = "true" ]; then
    substreams init --state-file /app/generator.json --force-download-cwd --codegen-endpoint http://host.docker.internal:9000
else
    substreams init --state-file /app/generator.json --force-download-cwd --codegen-endpoint https://codegen-staging.substreams.dev
fi

substreams build

## To validate the manifest
substreams info