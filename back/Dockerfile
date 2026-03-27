# syntax=docker/dockerfile:1

# ============================================
# Stage 1: Build dependencies
# ============================================
FROM golang:1.25-alpine AS deps

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.* ./
RUN go mod download

# ============================================
# Stage 2: Development (with hot reload)
# ============================================
FROM deps AS development

# Install dev tools (sqlc + migrate) for local generation/migrations
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install air for hot reloading
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copy source code (will be overwritten by volume mount)
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

# ============================================
# Stage 3: Builder
# ============================================
FROM deps AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o /server ./cmd/server

# ============================================
# Stage 4: Production
# ============================================
FROM alpine:3.19 AS production

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /server /server

EXPOSE 8080

CMD ["/server"]
