#!/usr/bin/env bash
#set -euxo pipefail

if [[ ${#BUF_TOKEN} == 0 ]]; then
    echo "BUF_TOKEN is not set"
    exit 1
fi
#substreams init --state-file /app/generator.json --force-download-cwd
#
#substreams build
#
### To validate the manifest
#substreams info
#
