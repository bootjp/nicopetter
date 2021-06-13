FROM golang:alpine AS build

ENV GO111MODULE=on

WORKDIR $GOPATH/src/github.com/bootjp/go_twitter_bot_for_nicopedia

COPY . .

RUN GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a -o out main/main.go && cp out /app

FROM gcr.io/distroless/static:latest-arm64

COPY --from=build /app /app

CMD ["/app"]
