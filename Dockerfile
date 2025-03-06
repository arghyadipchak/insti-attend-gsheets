# syntax=docker/dockerfile:1

# Base images
ARG GO_IMAGE=golang:1.24
ARG DIST_IMAGE=debian:12-slim

# Build config
ARG DEBIAN_FRONTEND=noninteractive
ARG WORKDIR=/app

FROM ${GO_IMAGE} AS build

ARG WORKDIR
WORKDIR ${WORKDIR}

COPY go.* .
RUN go mod download

COPY . .

ARG CGO_ENABLED=1
RUN make release

FROM ${DIST_IMAGE}

ARG DEBIAN_FRONTEND
ARG WORKDIR

WORKDIR ${WORKDIR}

RUN \
  apt-get update && \
  apt-get install -y --no-install-recommends ca-certificates tzdata && \
  rm -rf /var/lib/apt/lists/*

COPY --from=build ${WORKDIR}/bin/attender-cs219 /usr/local/bin/attender-cs219

EXPOSE 8080

CMD ["attender-cs219"]
