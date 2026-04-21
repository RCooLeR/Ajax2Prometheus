# --- Stage 1: Build the Go binary ---
FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

# Install git and CA certs (if needed for Go modules)
RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/ajax2prometheus ./cmd/ajax2prometheus

# --- Stage 2: Create a lightweight image with the binary only ---
FROM alpine:3.23

# Add CA certs for HTTPS support
RUN apk upgrade --no-cache && apk add --no-cache ca-certificates

RUN addgroup -S ajax2prometheus && adduser -S -G ajax2prometheus ajax2prometheus
RUN mkdir -p /data && chown ajax2prometheus:ajax2prometheus /data

USER ajax2prometheus
WORKDIR /app

COPY --from=builder /out/ajax2prometheus /app/ajax2prometheus

EXPOSE 8080 8099

ENV AJAX2PROM_HTTP_ADDR=:8080
ENV AJAX2PROM_SIA_ADDR=:8099
ENV AJAX2PROM_LOG_PRETTY=false
ENV AJAX2PROM_DEVICES_PATH=/data/devices.json

ENTRYPOINT ["/app/ajax2prometheus"]
