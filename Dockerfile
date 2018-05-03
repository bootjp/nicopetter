# TODO multi stage build.
# TODO timezone gobuild.
FROM alpine
# registry.gitlab.com/bootjp/twitterbot

ADD ./out /app

RUN apk add --no-cache ca-certificates

ENV GOROOT /usr/local/go
ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

CMD ["/app"]
