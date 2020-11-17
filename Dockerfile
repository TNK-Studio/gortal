FROM golang:1.14.12-alpine AS builder
RUN apk add --no-cache curl jq git build-base
WORKDIR /opt
RUN cd /opt
RUN mkdir /root/.ssh
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build ./
# CGO_ENABLED=0
# Change go-release.action

FROM alpine:latest
LABEL maintainer="Elf Gzp <gzp@741424975@gmail.com> (https://elfgzp.cn)"
COPY --from=builder /opt/gortal ./
RUN chmod +x /gortal
RUN mkdir -p /root/.ssh
EXPOSE 2222
VOLUME [ "/root", "/root/.ssh"]
CMD ["/gortal"]
