#!/bin/sh

substreams init --state-file /app/generator.json --force-download-cwd --codegen-endpoint https://codegen-staging.substreams.dev
substreams build
