FROM ghcr.io/streamingfast/substreams:develop AS substreams
FROM bufbuild/buf AS buf

# Forcing 1.86 until we fix starknet-events which requires 1.86 otherwise
# does not compile properly.
FROM rust:1.86

WORKDIR /app

COPY entrypoint.sh /app/entrypoint.sh

COPY --from=substreams /app/substreams /usr/local/bin/substreams
COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf

RUN rustup target add wasm32-unknown-unknown

RUN chmod +x /usr/local/bin/substreams && \
    chmod +x /app/entrypoint.sh

CMD ["/app/entrypoint.sh"]
