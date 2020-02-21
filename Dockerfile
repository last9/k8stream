FROM alpine:3.9.5 AS builder
RUN apk add ca-certificates

FROM builder
WORKDIR /app
COPY k8stream agent
