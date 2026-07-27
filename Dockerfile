# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY rex-node/go.mod ./
COPY rex-node/go.sum* ./
RUN go mod download
COPY rex-node/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o rex-node .

# Final minimal scratch container
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata bash systemd
WORKDIR /app
COPY --from=builder /app/rex-node /usr/local/bin/rex-node
COPY rex-node/config.yaml /etc/rex/config.yaml
COPY rex-node/allowlist.yaml /etc/rex/allowlist.yaml

EXPOSE 7443
ENTRYPOINT ["/usr/local/bin/rex-node"]
CMD ["--config", "/etc/rex/config.yaml"]
