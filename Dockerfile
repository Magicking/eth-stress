FROM golang:alpine AS builder

LABEL maintainer="Sylvain Laurent <s@6120.eu>"

# Install build environment dependencies
RUN apk add --no-cache wget curl git build-base

# Create source directory
RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Install go dependencies

# Build app executable
COPY cmd /go/src/app/cmd
RUN go get -v ./...
RUN go build -v -o /go/bin/stress /go/src/app/cmd/stress/*.go

# Run stage: expose application binary
FROM alpine:latest

# Copy binary
RUN mkdir -p /go/bin/stress
COPY --from=builder /go/bin/stress /go/bin/app/stress

# Set runtime command
ENTRYPOINT ["/go/bin/app/stress"]

# Expose port
EXPOSE 18547
