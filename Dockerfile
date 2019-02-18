FROM golang:alpine AS builder
LABEL maintainer="Sylvain Laurent <s@6120.eu>"
RUN apk add --no-cache wget curl git build-base
WORKDIR /go/src/app
RUN mkdir -p /go/src/app
COPY cmd /go/src/app/cmd
RUN go get -v ./... && \
    go build -v -o /go/bin/stress /go/src/app/cmd/stress/*.go

FROM alpine:latest
RUN mkdir -p /go/bin/stress
COPY --from=builder /go/bin/stress /go/bin/app/stress

ENTRYPOINT ["/go/bin/app/stress"]

# SendTransactionAsync HTTP server
EXPOSE 18547
