#!/usr/bin/env bash
set -euxo pipefail

substreams init --state-file /app/generator.json --force-download-cwd

cat > myfile.yaml <<EOF

version: v1
plugins:
- name: neoeinstein-prost
  out: ./src/pb
  opt:
    - file_descriptor_set=false
  path: /app/protoc-gen-prost-protoc-gen-tonic-v0.4.1/target/release/protoc-gen-prost
  

- plugin: neoeinstein-prost-crate
  out: ./src/pb
  opt:
    - no_features
  path: /app/protoc-gen-prost-protoc-gen-tonic-v0.4.1/target/release/protoc-gen-crate
EOF

substreams build

## To validate the manifest
substreams info
