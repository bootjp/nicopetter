package main

import (
	"os"
	"strconv"

	"sort"

	"net/url"

	"github.com/bootjp/go_twitter_bot_for_nicopedia/store"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/item"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mmcdole/gofeed"
)

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

		ar := nicopedia.ParseArticleType(u)

		_, err = api.PostTweet(e.Title+ar.PostArticleExpression+" に "+e.Description+"というお絵カキコが投稿されたよ。"+e.Link, v)
		if err != nil {
			return e, err
		}
	}

	return nil, nil
}

func routine() error {

	f, err := item.Fetch("https://dic.nicovideo.jp/feed/rss/n/oekaki")
	if err != nil {
		return err
	}

	// ON DEVELOP ONLY.
	err = os.Setenv("REDIS_HOST", "localhost")
	if err != nil {
		return err
	}
	err = os.Setenv("REDIS_INDEX", "0")
	if err != nil {
		return err
	}

	i, err := strconv.Atoi(os.Getenv("REDIS_INDEX"))
	if err != nil {
		return err
	}
	r := store.NewRedisClient(os.Getenv("REDIS_HOST"), i, "gunyapetter:")
	defer r.Close()

	t, err := r.GetLastUpdateTime()
	if err != nil {
		return err
	}
	f = item.FilterDate(f, t)

	if len(f) == 0 {
		// has not new item data
		return nil
	}

	// sort
	sort.Slice(f, func(i, j int) bool {
		return f[i].PublishedParsed.Before(*f[j].PublishedParsed)
	})

	ef, err := PostTwitter(f)
	if err != nil {
		println(ef)

		return err
	}

	err = r.Close()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := routine()
	if err != nil {
		panic(err)
	}
}
