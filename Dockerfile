FROM alpine

WORKDIR /app

RUN apk add ca-certificates
COPY k8stream agent
