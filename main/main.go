package main

import (
	"os"
	"strconv"

	"sort"

	"net/url"

	"fmt"

	"log"

	"github.com/ChimeraCoder/anaconda"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/bot"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/nicopedia"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/twitter"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/item"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/store"
	"github.com/mmcdole/gofeed"
	"gopkg.in/urfave/cli.v2"
)

// Twitter base struct.
type Twitter struct {
	twitter.Authorization
}

// SendSNS is testable interface.
type SendSNS interface {
	Twitter(i *gofeed.Item, authorization *twitter.Authorization) error
}

// PostTwitter is Item to Twitter post.
func (t *Twitter) PostTwitter(i *gofeed.Item, mode *bot.Behavior) error {
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

	var out string
	switch mode {
	case bot.Gunyapetter:
		out = fmt.Sprintf(mode.TweetFormat, i.Title, ar.PostArticleExpression, i.Description, i.Link)

	case bot.DulltterTmp:
		out = fmt.Sprintf(mode.TweetFormat, i.Title, ar.PostArticleExpression, i.Description, i.Link)

	case bot.NicopetterNewArticle:
		out = fmt.Sprintf(mode.TweetFormat, i.Title, i.Link)

	case bot.NicopetterRedirectArticle:
		// TODD GET REDIRECT
		out = fmt.Sprintf(mode.TweetFormat, i.Title, "redirect", i.Link)
	}

	if _, err = api.PostTweet(out, v); err != nil {
		println(fmt.Sprintf(mode.TweetFormat, i.Title, ar.PostArticleExpression, i.Description, i.Link))
		return err
	}

	return nil
}

func routine(mode *bot.Behavior) error {
	f, err := item.Fetch("https://dic.nicovideo.jp/feed/rss/n/oekaki")
	if err != nil {
		return err
	}

	i, err := strconv.Atoi(os.Getenv("REDIS_INDEX"))
	if err != nil {
		return err
	}
	r := store.NewRedisClient(os.Getenv("REDIS_HOST"), i, mode.StorePrefix)
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
		if err = sns.PostTwitter(v, mode); err != nil {
			println(v)
			return err
		}

		if err = r.SetLastUpdateTime(*v.PublishedParsed); err != nil {
			return nil
		}
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
	app := cli.App{}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "mode, m",
			Value: "test",
			Usage: "bot behavior mode.",
		},
	}
	app.Action = func(c *cli.Context) error {
		mode, err := bot.NewBehavior(c.String("mode"))
		if err != nil {
			return err
		}
		return routine(mode)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
