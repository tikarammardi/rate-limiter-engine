FROM golang:1.24-alpine AS builder
WORKDIR /app

# Cache dependencies by copying only go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build a static, stripped binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /guard ./cmd/guard/main.go


FROM alpine:3.18
# Install CA certs for TLS and create an unprivileged user in one layer
RUN apk --no-cache add ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app
COPY --from=builder /guard /app/guard

# Run as non-root user
USER app

EXPOSE 50051

ENTRYPOINT ["/app/guard"]
