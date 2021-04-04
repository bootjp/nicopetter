FROM golang:alpine AS build
ENV GO111MODULE=on


WORKDIR $GOPATH/src/github.com/bootjp/go_twitter_bot_for_nicopedia

COPY . .

RUN GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a -o out main/main.go && cp out /app

FROM golang:latest@sha256:e7de4081f3cb640bb4a0fd2f32402f551cbf0752b17f8b4ba8d5e49b9b49a170

COPY --from=build /app /app

CMD ["/app"]
