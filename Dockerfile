FROM golang:1.12-alpine AS builder
RUN apk add --no-cache curl jq git build-base
WORKDIR /opt
RUN cd /opt
RUN mkdir /root/.ssh
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build

FROM alpine:latest
LABEL maintainer="Elf Gzp <gzp@741424975@gmail.com> (https://elfgzp.cn)"
COPY --from=builder /opt/gortal ./
RUN chmod +x /gortal
RUN mkdir -p /root/.ssh
EXPOSE 2222
VOLUME [ "/root", "/root/.ssh"]
CMD ["/gortal"]
