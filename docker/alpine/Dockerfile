# build stage
ARG VERSION
FROM golang:1.21 AS builder

RUN mkdir /build
WORKDIR /build
ADD . .
RUN go get
RUN	CGO_ENABLED=0 go build -ldflags="-X 'main.Version=$VERSION'" -o gotf-extract  cmd/main.go

# final stage
FROM alpine:latest

# LABEL maintainer=""

RUN mkdir /app
WORKDIR /app

COPY --from=builder /build/gotf-extract /usr/local/bin/gotf-extract

ENTRYPOINT /usr/local/bin/gotf-extract