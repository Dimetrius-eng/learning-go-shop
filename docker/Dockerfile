# Builder stage
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache git

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN mkdir -p /app/bin && \
    for cmd in cmd/*/; do \
        if [ -d "$cmd" ]; then \
            binary=$(basename $cmd); \
            echo "Building $binary..."; \
            CGO_ENABLED=0 GOOS=linux go build -o bin/$binary ./$cmd; \
        fi \
    done

# Final stage
FROM alpine:3.18

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk update \
    && apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/* ./
COPY --from=builder /app/docs ./docs

CMD ["./api"]
