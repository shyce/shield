# Stage 1: Build
FROM golang:1.20.1-alpine as builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git jq bash openssl

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN chmod +x build.sh && ./build.sh

# Stage 2: Run
FROM alpine:3.18

# Install openssl
RUN apk add --no-cache openssl

# Copy binary from builder to /usr/bin
COPY --from=builder /build/shield /usr/bin/shield

# Create a new user 'user' and set it as the default.
RUN addgroup -g 1000 user && adduser -u 1000 -G user -s /bin/sh -D user

# Change the ownership of the /usr/bin/shield to the new user and group
RUN chown 1000:1000 /usr/bin/shield

# Switch to the new user
USER user

# Set working directory to /app 
WORKDIR /app

ENTRYPOINT ["shield"]
