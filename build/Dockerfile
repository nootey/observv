# Stage 1: Build
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy source code
COPY cmd ./cmd
COPY config ./config
COPY internal ./internal
COPY pkg ./pkg

# Build all binaries
ENV CGO_ENABLED=0
RUN go build -ldflags="-w -s" -o bin/api ./cmd/api
RUN go build -ldflags="-w -s" -o bin/simulator ./cmd/simulator
RUN go build -ldflags="-w -s" -o bin/migrate ./cmd/migrate
RUN go build -ldflags="-w -s" -o bin/seed ./cmd/seed

# Stage 2: Runtime
FROM alpine:3.24

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Copy binaries
COPY --from=builder /app/bin/ /usr/local/bin/

# Copy migrations
COPY migrations ./migrations

ENTRYPOINT ["api"]
