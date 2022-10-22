FROM golang:1.18

WORKDIR /go/src/github.com/boost-telegram-bot
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/boost-telegram-bot github.com/boost-telegram-bot

###

FROM alpine:3.9

COPY --from=0 /usr/local/bin/boost-telegram-bot /usr/local/bin/boost-telegram-bot
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["boost-telegram-bot"]
