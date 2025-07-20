ARG GO_VERSION=1.24.5
ARG DEBIAN_VERSION=12

ARG BUILD_DIR=/app

FROM golang:${GO_VERSION} AS build

ARG BUILD_DIR
WORKDIR ${BUILD_DIR}

COPY go.* .
RUN go mod download

COPY . .

ARG CGO_ENABLED=0
RUN make release

FROM debian:${DEBIAN_VERSION}-slim AS tzdata

ARG DEBIAN_FRONTEND=noninteractive
RUN \
  apt-get update && \
  apt-get install -y --no-install-recommends tzdata && \
  rm -rf /var/lib/apt/lists/*

FROM gcr.io/distroless/static-debian${DEBIAN_VERSION}

ARG BUILD_DIR

COPY --from=tzdata /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build ${BUILD_DIR}/bin/attender /

EXPOSE 8080

CMD ["/attender"]
