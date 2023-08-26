FROM golang:1.21-alpine AS build-stage
WORKDIR /app
COPY . /app/

ARG tag="dev"

RUN apk add git
RUN CGO_ENABLED=0 GOOS=linux go build -o /codebuild-multirunner -ldflags "\
    -X main.version=${tag} \
    -X main.date=$(date -Iseconds -u) \
    -X main.commit=$(git rev-parse HEAD) \
    "

FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /codebuild-multirunner /usr/local/bin/codebuild-multirunner

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
