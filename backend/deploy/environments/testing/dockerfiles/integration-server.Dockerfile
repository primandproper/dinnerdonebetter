# syntax=docker/dockerfile:1
FROM golang:1.26-trixie AS build-stage

WORKDIR /go/src/github.com/primandproper/dinnerdonebetter/backend

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY scripts scripts
COPY vendor vendor
COPY go.mod go.mod
COPY go.sum go.sum

RUN --mount=type=cache,target=/root/.cache/go-build ./scripts/build.sh -o /dinnerdonebetter github.com/primandproper/dinnerdonebetter/backend/cmd/services/api

# final stage
FROM debian:bullseye

# RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
COPY --from=build-stage /dinnerdonebetter /dinnerdonebetter

ENTRYPOINT ["/dinnerdonebetter"]
CMD ["serve"]
