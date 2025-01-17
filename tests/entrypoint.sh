#!/usr/bin/env bash
set -euxo pipefail

substreams init --state-file /app/generator.json --force-download-cwd

cat > buf.gen.yaml <<EOF
version: v1
plugins:
- name: neoeinstein-prost
  out: ./src/pb
  opt:
    - file_descriptor_set=false
  path: /app/protoc-gen-prost-protoc-gen-prost-v0.4.0/target/release/protoc-gen-prost
  

- plugin: neoeinstein-prost-crate
  out: ./src/pb
  opt:
    - no_features
  path: /app/protoc-gen-prost-protoc-gen-prost-crate-v0.4.1/target/release/protoc-gen-prost-crate
EOF

ls
cat buf.gen.yaml

substreams build

## To validate the manifest
substreams info
