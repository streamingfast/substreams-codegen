# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN apk add build-base git

RUN --mount=type=secret,id=github_token git config --global url."https://`cat /run/secrets/github_token`:x-oauth-basic@github.com/".insteadOf "https://github.com/"

ARG TARGETOS TARGETARCH
RUN --mount=target=. \
    --mount=type=ssh \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /app/substreams-codegen ./cmd/substreams-codegen

FROM --platform=$TARGETPLATFORM alpine:3.14.2 AS downlader

RUN apk --no-cache add curl

RUN mkdir /app
RUN curl -L https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.9/grpc_health_probe-linux-amd64 -o /app/grpc_health_probe && chmod +x /app/grpc_health_probe

FROM alpine:3.14.2

RUN apk --no-cache add openssl ca-certificates
RUN apk --no-cache add bash

RUN mkdir /app

COPY --from=builder /app/substreams-codegen /app/substreams-codegen
COPY --from=downlader /app/grpc_health_probe /app/grpc_health_probe


WORKDIR /app

EXPOSE 8080

CMD [ "/app/substreams-codegen" ]
