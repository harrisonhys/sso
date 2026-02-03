# 1. Build UI Stage
FROM node:20-alpine as ui-builder

WORKDIR /ui

# Copy package files
COPY sso-server/ui/package.json sso-server/ui/package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY sso-server/ui .

# Build static files
RUN npm run generate

# 2. Build Backend Stage
FROM golang:1.21-alpine as backend-builder

WORKDIR /server

# Install git for fetch
RUN apk add --no-cache git

# Copy go mod files
COPY sso-server/go.mod sso-server/go.sum ./
RUN go mod download

# Copy source code
COPY sso-server .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 3. Final Stage
FROM alpine:latest

# Certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from backend-builder
COPY --from=backend-builder /server/main ./server

# Copy static files from ui-builder
COPY --from=ui-builder /ui/.output/public ./static

# Expose port
EXPOSE 3000

# Run
CMD ["./server"]
