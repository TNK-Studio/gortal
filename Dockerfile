FROM golang:1.12-alpine

LABEL maintainer="Elf Gzp <gzp@741424975@gmail.com> (https://elfgzp.cn)"

RUN apk add --no-cache curl jq git build-base

WORKDIR /app

RUN cd /app

RUN mkdir /root/.ssh

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build

EXPOSE 2222

VOLUME [ "/root/.ssh" ]

CMD ["./gortal"]
