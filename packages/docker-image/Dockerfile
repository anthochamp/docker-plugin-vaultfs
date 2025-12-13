FROM golang:alpine3.21@sha256:b4dbd292a0852331c89dfd64e84d16811f3e3aae4c73c13d026c4d200715aff6 AS builder
ARG appVersion
ARG commitHash
ARG buildDate
COPY src /tmp/src
WORKDIR /tmp/src
RUN CGO_ENABLED=0 go build -v -ldflags "-s -w -X main.appVersion=$appVersion -X main.commitHash=$commitHash -X main.buildDate=$buildDate"

FROM alpine:3.23.0@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375
# hadolint ignore=DL3018
RUN apk add --no-cache fuse
COPY --from=builder --chown=0:0 /tmp/src/docker-plugin-vaultfs /usr/local/bin
CMD [ "/usr/local/bin/docker-plugin-vaultfs" ]
