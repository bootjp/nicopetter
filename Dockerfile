# TODO multi stage build.
# TODO timezone gobuild.
FROM alpine
# registry.gitlab.com/bootjp/twitterbot

RUN apk add --no-cache ca-certificates git && mkdir -p /usr/local/go/src/ && cd /usr/local/go/src/ && git clone https://github.com/bootjp/go_twitter_bot_for_nicopedia.git

ENV GOROOT /usr/local/go
ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

CMD ["/app"]
