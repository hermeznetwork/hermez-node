FROM golang:alpine

VOLUME /build/dist
WORKDIR /build

RUN apk add --update --no-cache git build-base musl-dev linux-headers bash make
RUN go get github.com/gobuffalo/packr/v2/packr2

COPY . .
