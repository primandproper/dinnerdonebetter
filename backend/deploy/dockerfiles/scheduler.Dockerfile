# build stage
FROM golang:1.26-trixie AS build-stage

WORKDIR /go/src/github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend

COPY . .

RUN go build -trimpath -o /action github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/cmd/workers/scheduler

# final stage
FROM debian:bullseye

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
COPY --from=build-stage /action /action

ENTRYPOINT ["/action"]
