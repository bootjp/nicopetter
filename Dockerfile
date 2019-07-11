FROM golang:alpine AS build
ENV GO111MODULE=on

RUN apk add --no-cache git
RUN \
  cd $GOPATH/src/ && \
  mkdir -p github.com/bootjp && \
  cd github.com/bootjp && \
  git clone https://github.com/bootjp/go_twitter_bot_for_nicopedia.git && \
  cd ./go_twitter_bot_for_nicopedia && \
  GOOS=linux CGO_ENABLED=0 go build -a -o out main/main.go && \
  cp out /app

FROM alpine
RUN apk add --no-cache tzdata ca-certificates
COPY --from=build /app /app

CMD ["/app"]
