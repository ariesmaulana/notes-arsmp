# ---- Build Stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and embeddable files
COPY . .

# Build the binary (CGO disabled for static linking)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/server \
    .

# ---- Final Stage ----
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose the default port
EXPOSE 8080

# The app's entrypoint is `serve`; allow overriding via CMD/args
ENTRYPOINT ["/app/server"]
CMD ["serve"]
