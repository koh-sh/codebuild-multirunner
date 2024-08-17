FROM golang:1.23-alpine AS build-stage
WORKDIR /app
COPY . /app/

RUN CGO_ENABLED=0 GOOS=linux go build -o /runcbs

FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /runcbs /usr/local/bin/runcbs

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
