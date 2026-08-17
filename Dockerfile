FROM golang:alpine AS builder

WORKDIR /app

COPY go.* .
RUN go mod download

COPY . .

RUN apk add --no-cache ca-certificates tzdata
RUN update-ca-certificates

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 go build -ldflags "-w -s" -trimpath -o /insti-attend-gsheets .

FROM scratch

COPY --from=builder /insti-attend-gsheets /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

VOLUME ["/data"]

EXPOSE 8090

ENTRYPOINT ["/insti-attend-gsheets"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/data"]
