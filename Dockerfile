FROM golang:1.12.7-buster as builder
WORKDIR /go/src/github.com/scm-manager/ci-plugin-snapshot
COPY *.go ./
COPY go* ./
COPY center ./center
COPY Makefile ./
RUN make

FROM alpine:3.11.3
RUN apk add --update ca-certificates
COPY --from=builder /go/src/github.com/scm-manager/ci-plugin-snapshot/ci-plugin-snapshot /ci-plugin-snapshot

ENTRYPOINT [ "/ci-plugin-snapshot" ]
