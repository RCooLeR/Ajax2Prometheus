# --- Stage 1: Build the Go binary ---
FROM golang:1.26.4-alpine3.24 AS builder

ARG TARGETOS
ARG TARGETARCH

# Install git and CA certs (if needed for Go modules)
RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags="-s -w" -o /out/ajaxbridge ./cmd/ajaxbridge

# --- Stage 2: Create a lightweight image with the binary only ---
FROM alpine:3.24

# Add CA certs for HTTPS support
RUN apk upgrade --no-cache && apk add --no-cache ca-certificates

RUN addgroup -S ajaxbridge && adduser -S -G ajaxbridge ajaxbridge
RUN mkdir -p /data && chown ajaxbridge:ajaxbridge /data

USER ajaxbridge
WORKDIR /app

COPY --from=builder /out/ajaxbridge /app/ajaxbridge
COPY --chown=ajaxbridge:ajaxbridge logo.png /app/logo.png

EXPOSE 8080 8099

ENV AJAXBRIDGE_HTTP_ADDR=:8080
ENV AJAXBRIDGE_SIA_ADDR=:8099
ENV AJAXBRIDGE_LOG_PRETTY=false
ENV AJAXBRIDGE_DEVICES_PATH=/data/devices.json
ENV AJAXBRIDGE_NOTIFICATIONS_PATH=/data/notifications.json
ENV AJAXBRIDGE_JEEDOM_STORE_PATH=/data/jeedom.json

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD ["/app/ajaxbridge", "healthcheck"]

ENTRYPOINT ["/app/ajaxbridge"]
