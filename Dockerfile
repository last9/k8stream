FROM alpine:3.9.5 AS builder
RUN apk add ca-certificates wget

FROM builder
WORKDIR /app

ENTRYPOINT wget --show-progress -qO k8stream https://github.com/last9/k8stream/releases/${VERSION:-latest}/download/k8stream_linux_x86_64 && chmod a+x k8stream && /app/k8stream --config /data/config.json
