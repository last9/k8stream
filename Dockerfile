FROM alpine:3.9.5 AS builder
RUN apk add ca-certificates

FROM builder
WORKDIR /app

COPY deploy/run.sh run.sh
RUN chmod a+x run.sh
