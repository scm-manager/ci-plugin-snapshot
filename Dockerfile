FROM golang:1.12.7-buster as builder
WORKDIR /go/src/bitbucket.org/scm-manager/scm-plugin-snapshot
COPY *.go ./
COPY go* ./
COPY center ./center
COPY Makefile ./
RUN make

FROM alpine:3.10.1
RUN apk add --update ca-certificates
COPY --from=builder /go/src/bitbucket.org/scm-manager/scm-plugin-snapshot/scm-plugin-snapshot /scm-plugin-snapshot

ENTRYPOINT [ "/scm-plugin-snapshot" ]
