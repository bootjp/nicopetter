package main

import (
	"fmt"
	"os"
	"strconv"

	"time"

	"github.com/go-redis/redis"
	"github.com/mmcdole/gofeed"
)

func FetchFeed(url string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	return fp.ParseURL(url)
}

type Redis struct {
	c      *redis.Client
	prefix string
}

func NewRedisClient(host string, index int, prefix string) *Redis {
	r := &Redis{}
	r.c = redis.NewClient(&redis.Options{
		Addr: host + ":6379",
		DB:   index,
	})
	r.prefix = prefix

	return r
}

func FilterFeedDate(f []*gofeed.Item, t time.Time) []*gofeed.Item {
	var res []*gofeed.Item
	for i, elem := range f {
		if t.Before(*elem.PublishedParsed) {
			fmt.Printf("%v\n", elem)
			res = append(res, f[i])
		}
	}
	return res
}

func main() {
	f, err := FetchFeed("https://dic.nicovideo.jp/feed/rss/n/oekaki")
	if err != nil {
		panic(err)
	}

	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_INDEX", "0")

	i, err := strconv.Atoi(os.Getenv("REDIS_INDEX"))
	if err != nil {
		panic(err)
	}

	c := NewRedisClient(os.Getenv("REDIS_HOST"), i, "gunyapetter:")

	res, err := c.c.Get(c.prefix + "lastDate").Result()
	loc, _ := time.LoadLocation("Asia/Tokyo")
	var ld time.Time
	if err != nil {
		ld = time.Date(2000, 1, 1, 1, 1, 1, 1, loc)
		panic(err)
	} else {

		ld, err = time.ParseInLocation("2006-01-02 15:04:05", res, loc)
		fmt.Println(res)
		fmt.Println(ld)
		if err != nil {
			panic(err)
		}
	}

	it := f.Items
	f.Items = FilterFeedDate(it, ld)
	// for _, v := range f.Items {
	// 	fmt.Printf("%v\n", v)
	//
	// }

}
