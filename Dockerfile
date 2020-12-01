FROM golang:1.14.3-alpine AS build

ENV CGO_ENABLED=0
ENV GO111MODULE=auto
COPY . /src
WORKDIR /src/cli/node

RUN go build -o app .

##FROM scratch AS bin-unix
##COPY --from=build /cli/node/cfg.buidler.toml .
##COPY --from=build /cli/node/ ./app
EXPOSE 8545
CMD ["./app", "--mode", "sync","--cfg", "cfg.buidler.toml", "run"]