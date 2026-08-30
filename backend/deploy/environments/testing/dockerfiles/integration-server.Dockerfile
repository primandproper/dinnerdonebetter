# syntax=docker/dockerfile:1
FROM golang:1.27-trixie AS build-stage

WORKDIR /go/src/github.com/primandproper/dinnerdonebetter/backend

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY go.mod go.mod
COPY go.sum go.sum
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY scripts scripts

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod ./scripts/build.sh -o /dinnerdonebetter github.com/primandproper/dinnerdonebetter/backend/cmd/ddb

# final stage
FROM debian:bullseye

# RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
COPY --from=build-stage /dinnerdonebetter /dinnerdonebetter

ENTRYPOINT ["/dinnerdonebetter"]
CMD ["serve"]
