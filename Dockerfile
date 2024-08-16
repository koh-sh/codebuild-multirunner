FROM golang:1.23-alpine AS build-stage
WORKDIR /app
COPY . /app/

RUN CGO_ENABLED=0 GOOS=linux go build -o /codebuild-multirunner

FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /codebuild-multirunner /usr/local/bin/codebuild-multirunner

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
