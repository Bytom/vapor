# Build Bytom in a stock Go builder container
FROM golang:1.12-alpine as builder

RUN apk add --no-cache make git

ADD . /go/src/github.com/vapor
RUN mkdir /root/.vapor
ADD ./config/config.toml /root/.vapor/
ADD ./config/federation.json /root/.vapor/
RUN cd /go/src/github.com/vapor && make bytomd && make bytomcli

# Pull Bytom into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/vapor/cmd/bytomd/bytomd /usr/local/bin/
COPY --from=builder /go/src/github.com/vapor/cmd/bytomcli/bytomcli /usr/local/bin/

EXPOSE 9889 56659 46658
