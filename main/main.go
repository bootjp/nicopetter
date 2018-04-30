package main

import (
	"os"
	"strconv"

	"sort"

	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/nicopedia"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/nicopedia/twitter"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/item"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/store"
	"github.com/mmcdole/gofeed"
)

// PostTwitter is Item to Twitter post.
func PostTwitter(i *gofeed.Item, authorization *twitter.Authorization) error {
	api := anaconda.NewTwitterApiWithCredentials(
		authorization.AccessToken,
		authorization.AccessTokenSecret,
		authorization.ConsumerKey,
		authorization.ConsumerSecret)
	v := url.Values{}

	u, err := url.Parse(i.Link)
	if err != nil {
		return err
	}
	ar := nicopedia.ParseArticleType(u)

	_, err = api.PostTweet(i.Title+ar.PostArticleExpression+" に "+i.Description+"というお絵カキコが投稿されたよ。"+i.Link, v)
	if err != nil {
		return err
	}

	return nil
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
	au := &twitter.Authorization{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	}

	for _, v := range f {
		err = PostTwitter(v, au)
		if err != nil {
			println(v)

			return err
		}
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
