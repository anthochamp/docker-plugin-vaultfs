FROM golang:alpine3.21@sha256:933e5a0829a1f97cc99917e3b799ebe450af30236f5a023a3583d26b5ef9166f AS builder
ARG appVersion
ARG commitHash
ARG buildDate
COPY src /tmp/src
WORKDIR /tmp/src
RUN CGO_ENABLED=0 go build -v -ldflags "-s -w -X main.appVersion=$appVersion -X main.commitHash=$commitHash -X main.buildDate=$buildDate"

FROM alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715
# hadolint ignore=DL3018
RUN apk add --no-cache fuse
COPY --from=builder --chown=0:0 /tmp/src/docker-plugin-vaultfs /usr/local/bin
CMD [ "/usr/local/bin/docker-plugin-vaultfs" ]
