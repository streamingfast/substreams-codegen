FROM ghcr.io/streamingfast/substreams:5e372ed as substreams
FROM bufbuild/buf as buf
FROM rust:latest

WORKDIR /app

COPY entrypoint.sh /app/entrypoint.sh

COPY --from=substreams /app/substreams /usr/local/bin/substreams
COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf

RUN rustup target add wasm32-unknown-unknown
RUN chmod +x /usr/local/bin/substreams && \
    chmod +x /app/entrypoint.sh

CMD ["/app/entrypoint.sh"]







