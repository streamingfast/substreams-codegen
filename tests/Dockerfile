FROM ghcr.io/streamingfast/substreams:develop AS substreams
FROM bufbuild/buf AS buf
FROM rust:latest

WORKDIR /app

COPY entrypoint.sh /app/entrypoint.sh

COPY --from=substreams /app/substreams /usr/local/bin/substreams
COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf

RUN curl -LO https://github.com/neoeinstein/protoc-gen-prost/archive/refs/tags/protoc-gen-prost-v0.3.1.tar.gz \
&& tar -xzf protoc-gen-prost-v0.3.1.tar.gz \
&& rm protoc-gen-prost-v0.3.1.tar.gz
RUN cd protoc-gen-prost-protoc-gen-prost-v0.3.1 && cargo build --release --locked

RUN rustup target add wasm32-unknown-unknown

RUN chmod +x /usr/local/bin/substreams && \
    chmod +x /app/entrypoint.sh

CMD ["/app/entrypoint.sh"]
