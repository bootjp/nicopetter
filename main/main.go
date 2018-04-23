package main

import (
	"os"
	"strconv"

	"time"

	"sort"

	"net/url"

	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/go-redis/redis"
	"github.com/mmcdole/gofeed"
)

const dateFormat = "2006-01-02 15:04:05"

// General is general article type.
var (
	General = &ArticleType{"dic.nicovideo.jp/b/a/", "【単】"}
	// User is user article type.
	User = &ArticleType{"dic.nicovideo.jp/b/u/", "【ユ】"}

	// Live is live article type.
	Live = &ArticleType{"dic.nicovideo.jp/b/l/", "【生】"}

	// Video is movie article type.
	Video = &ArticleType{"dic.nicovideo.jp/b/v/", "【動】"}

	// Market is ichiba article type.
	Market = &ArticleType{"dic.nicovideo.jp/b/i/", "【市】"}

	// Community is community article type.
	Community = &ArticleType{"dic.nicovideo.jp/b/c/", "【コ】"}

	// Other is undefined article type.
	Other = &ArticleType{"", "【？】"}
)

// ArticleType is Nicopedia of article type.
type ArticleType struct {
	URLPrefix             string
	PostArticleExpression string
}

// FetchFeed is got url to fetch and return rss.
func FetchFeed(URL string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	return fp.ParseURL(URL)
}

// Redis is a convenient type of Redis operation for bot.
type Redis struct {
	c *redis.Client
	p string
}

// RedisI is a convenient interface of Redis operation for bot.
type RedisI interface {
	GetLastUpdateTime() (time.Time, error)
	Close() error
}

// NewRedisClient is a new redis connect instance  operation for bot.
func NewRedisClient(host string, index int, prefix string) *Redis {
	r := &Redis{}
	r.p = prefix
	r.c = redis.NewClient(&redis.Options{
		Addr: host + ":6379",
		DB:   index,
	})

	return r
}

// FilterFeedDate はfeedを受け取り，timeより新しい記事だけに限定したフィルタを書けます.
func FilterFeedDate(f []*gofeed.Item, t time.Time) []*gofeed.Item {
	var res []*gofeed.Item
	for i, elem := range f {
		if t.Before(*elem.PublishedParsed) {
			res = append(res, f[i])
		}
	}

	return res
}

// GetLastUpdateTime は最後にいつRSSの更新があったかを返す関数です.
func (c *Redis) GetLastUpdateTime() (time.Time, error) {
	res, err := c.c.Get(c.p + "lastDate").Result()
	if err != nil {
		return time.Now(), err
	}

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Now(), err
	}

	return time.ParseInLocation(dateFormat, res, loc)
}

// Close is Redis connection close.
func (c *Redis) Close() error {
	return c.c.Close()
}

// ParseArticleType from url to ArticleType.
func ParseArticleType(url *url.URL) *ArticleType {
	switch {
	case strings.HasPrefix(url.Host+url.Path, General.URLPrefix):
		return General
	case strings.HasPrefix(url.Host+url.Path, User.URLPrefix):
		return User
	case strings.HasPrefix(url.Host+url.Path, Video.URLPrefix):
		return Video
	case strings.HasPrefix(url.Host+url.Path, Live.URLPrefix):
		return Live
	case strings.HasPrefix(url.Host+url.Path, Market.URLPrefix):
		return Market
	case strings.HasPrefix(url.Host+url.Path, Community.URLPrefix):
		return Community
	default:
		return Other
	}
}

// PostTwitter is Item to Twitter post.
func PostTwitter(f []*gofeed.Item) (*gofeed.Item, error) {
	api := anaconda.NewTwitterApiWithCredentials(
		os.Getenv("ACCESS_TOKEN"),
		os.Getenv("ACCESS_TOKEN_SECRET"),
		os.Getenv("CONSUMER_KEY"),
		os.Getenv("CONSUMER_SECRET"))
	v := url.Values{}
	for _, e := range f {
		u, err := url.Parse(e.Link)
		if err != nil {
			return e, err
		}

		ar := ParseArticleType(u)

		_, err = api.PostTweet(e.Title+ar.PostArticleExpression+" に "+e.Description+"というお絵カキコが投稿されたよ。"+e.Link, v)
		if err != nil {
			return e, err
		}
	}

	return nil, nil
}

func main() {
	f, err := FetchFeed("https://dic.nicovideo.jp/feed/rss/n/oekaki")
	if err != nil {
		panic(err)
	}

	// ONDEVELOP ONLY.
	err = os.Setenv("REDIS_HOST", "localhost")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("REDIS_INDEX", "0")
	if err != nil {
		panic(err)
	}

	i, err := strconv.Atoi(os.Getenv("REDIS_INDEX"))
	if err != nil {
		panic(err)
	}
	r := NewRedisClient(os.Getenv("REDIS_HOST"), i, "gunyapetter:")
	defer r.Close()

	t, err := r.GetLastUpdateTime()
	if err != nil {
		panic(err)
	}
	f.Items = FilterFeedDate(f.Items, t)

	if len(f.Items) == 0 {
		// has not new feed data
		os.Exit(0)
	}
	// sort
	sort.Slice(f.Items, func(i, j int) bool {
		return f.Items[i].PublishedParsed.Before(*f.Items[j].PublishedParsed)
	})

	ef, err := PostTwitter(f.Items)
	if err != nil {
		println(ef)
		panic(err)
	}

	err = r.Close()
	if err != nil {
		panic(err)
	}

}
