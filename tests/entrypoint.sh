#!/usr/bin/env bash
set -euxo pipefail

substreams init --state-file /app/generator.json --force-download-cwd

substreams build

## To validate the manifest
substreams info
