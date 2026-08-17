FROM alpine:latest AS sysinfo

RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

FROM --platform=$BUILDPLATFORM golang:alpine AS builder

WORKDIR /app

COPY go.* .
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags "-w -s" -trimpath -o /insti-attend-gsheets .

FROM scratch

COPY --from=builder /insti-attend-gsheets /
COPY --from=sysinfo /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=sysinfo /usr/share/zoneinfo /usr/share/zoneinfo

VOLUME ["/data"]

EXPOSE 8090

ENTRYPOINT ["/insti-attend-gsheets"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/data"]
