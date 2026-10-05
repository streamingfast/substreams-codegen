FROM ghcr.io/streamingfast/substreams:develop AS substreams
# Pinned: an untagged image floats to :latest, so a buf release changes this test with no
# commit here. Keep in step with the version developers run.
FROM bufbuild/buf:1.67.0 AS buf

FROM rust:1.89

WORKDIR /app

COPY entrypoint.sh /app/entrypoint.sh

COPY --from=substreams /app/substreams /usr/local/bin/substreams
COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf

RUN rustup target add wasm32-unknown-unknown

RUN chmod +x /usr/local/bin/substreams && \
    chmod +x /app/entrypoint.sh

CMD ["/app/entrypoint.sh"]
