# go_twitter_bot

## how to use 

dependence Redis is require.

```bash
cd $GOPATH/src/ 
mkdir -p github.com/bootjp
cd github.com/bootjp
git clone git@github.com:bootjp/go_twitter_bot_for_nicopedia.git
go get -u github.com/golang/dep/...
dep ensure
GOOS=linux CGO_ENABLED=0 go build -a -o out main/main.go
REDIS_HOST='localhost' REDIS_INDEX=0 CONSUMER_SECRET='c_sec_here' CONSUMER_KEY='c_key_here' ACCESS_TOKEN='token_here' ACCESS_TOKEN_SECRET='token_secret_here' ./out 
```

[docker images](https://hub.docker.com/r/bootjp/go_twitter_bot_for_nicopedia/) 

