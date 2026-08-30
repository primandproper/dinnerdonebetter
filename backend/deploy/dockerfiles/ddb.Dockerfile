# build stage
FROM golang:1.27-trixie AS build-stage

WORKDIR /go/src/github.com/primandproper/dinnerdonebetter/backend

COPY . .

RUN ./scripts/build.sh -o /ddb github.com/primandproper/dinnerdonebetter/backend/cmd/ddb

# final stage
FROM debian:bullseye

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
COPY --from=build-stage /ddb /ddb

# No default subcommand: every workload names the one it wants in its manifest's args, so a
# deployment that forgets to gets help text and a non-zero exit rather than the wrong workload.
ENTRYPOINT ["/ddb"]
