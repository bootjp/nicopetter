package main

import (
	"os"
	"strconv"

	"sort"

	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/nicopedia"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/twitter"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/item"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/store"
	"github.com/mmcdole/gofeed"
)

type Twitter struct {
	twitter.Authorization
}

type SendSNS interface {
	Twitter(i *gofeed.Item, authorization *twitter.Authorization) error
}

// Twitter is Item to Twitter post.
func (t *Twitter) PostTwitter(i *gofeed.Item) error {
	api := anaconda.NewTwitterApiWithCredentials(
		t.AccessToken,
		t.AccessTokenSecret,
		t.ConsumerKey,
		t.ConsumerSecret,
	)

	v := url.Values{}

	u, err := url.Parse(i.Link)
	if err != nil {
		return err
	}
	ar := nicopedia.ParseArticleType(u)

	_, err = api.PostTweet(i.Title+ar.PostArticleExpression+" に "+i.Description+"というお絵カキコが投稿されたよ。"+i.Link, v)
	if err != nil {
		println(i.Title + ar.PostArticleExpression + " に " + i.Description + "というお絵カキコが投稿されたよ。" + i.Link)
		return err
	}

	return nil
}

func routine() error {

	f, err := item.Fetch("https://dic.nicovideo.jp/feed/rss/n/oekaki")
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
	au := twitter.Authorization{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	}

	sns := Twitter{au}

	for _, v := range f {
		err = sns.PostTwitter(v)
		if err != nil {
			println(v)
			return err
		}
		r.SetLastUpdateTime(*v.PublishedParsed)
	}

	if err != nil {
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
