package main

import (
	"os"
	"strconv"

	"sort"

	"net/url"

	"fmt"

	"log"

	"net/http"
	"time"

	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/bot"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/nicopedia"
	mytwitter "github.com/bootjp/go_twitter_bot_for_nicopedia/domain/twitter"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/item"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/store"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v2"
)

// Twitter base struct.
type Twitter struct {
	mytwitter.Authorization
}

// SendSNS is testable interface.
type SendSNS interface {
	PostTwitter(i *gofeed.Item, rd *nicopedia.Redirect, mode *bot.Behavior) error
}

// PostTwitter is Item to Twitter post.
func (t *Twitter) PostTwitter(i *gofeed.Item, rd *nicopedia.Redirect, mode *bot.Behavior) error {
	config := oauth1.NewConfig(t.Authorization.ConsumerKey, t.Authorization.ConsumerSecret)
	token := oauth1.NewToken(t.Authorization.AccessToken, t.Authorization.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	httpClient.Timeout = time.Duration(10) * time.Second
	client := twitter.NewClient(httpClient)

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

	case bot.NicopetterNewRedirectArticle:
		out = fmt.Sprintf(mode.TweetFormat, i.Title, rd.Title, i.Link)
	case bot.NicopetterModifyRedirectArticle:
		out = fmt.Sprintf(mode.TweetFormat, i.Title, rd.Title, i.Link)
	}

	tweet, resp, err := client.Statuses.Update(out, nil)

	if err != nil {

		fmt.Printf("%v\n", tweet)
		fmt.Printf("%v\n", resp)
		println(out)
		return err
	}

	return nil
}

// FetchRedirectTitle is Nicopedia user redirect setting article redirect page title.
func FetchRedirectTitle(u *url.URL) (*string, error) {
	const TitleSuffix = `location.replace('http://dic.nicovideo.jp/a/`
	c := http.Client{Timeout: time.Duration(10) * time.Second}
	res, err := c.Get(u.String())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	var head string
	doc.Find("head").Each(func(i int, s *goquery.Selection) {
		head = s.Text()
	})

	redirect := strings.Contains(head, `location.replace`)
	if !redirect {
		return nil, ErrNoRedirect
	}
	f := strings.Index(head, TitleSuffix)
	if f == -1 {
		return nil, ErrNoRedirect
	}

	head = head[f+len(TitleSuffix):]
	i := strings.Index(head, `'`)
	head = head[:i]

	title, err := url.QueryUnescape(head)
	if err != nil {
		return nil, err
	}

	return &title, nil
}

// ErrNoRedirect not redirect article err.
var ErrNoRedirect = errors.New("no redirect in response")

func routine(mode *bot.Behavior) error {
	f, err := item.Fetch(mode.FeedURL)
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

	switch mode {
	case bot.Gunyapetter, bot.DulltterTmp:
		f = item.FilterDate(f, t)
	case bot.NicopetterModifyRedirectArticle, bot.NicopetterNewArticle, bot.NicopetterNewRedirectArticle:
		f, err = item.FilterMarkedAsPost(f, r)
		if err != nil {
			return err
		}
	}

	if len(f) == 0 {
		return nil
	}

	// sort
	sort.Slice(f, func(i, j int) bool {
		return f[i].PublishedParsed.Before(*f[j].PublishedParsed)
	})

	sns := Twitter{createTwitterAuth()}

	lastPublish := t
	for _, v := range f {
		red := &nicopedia.Redirect{Exits: false}
		switch mode {
		case bot.NicopetterNewArticle:
			red, err = extractRedirect(v)
			if err != nil {
				return err
			}
			// 新着モードでリダイレクトしているものは無視する
			if red.Exits {
				continue
			}
		case bot.NicopetterModifyRedirectArticle, bot.NicopetterNewRedirectArticle:
			red, err = extractRedirect(v)
			if err != nil {
				return err
			}
			// リダイレクトモードでリダイレクト先が見つからないものは無視する
			if !red.Exits {
				continue
			}
		}

		switch mode {
		case bot.Gunyapetter, bot.DulltterTmp:
			if err = r.SetLastUpdateTime(*v.PublishedParsed); err != nil {
				return err
			}
		case bot.NicopetterNewRedirectArticle, bot.NicopetterNewArticle, bot.NicopetterModifyRedirectArticle:
			if err = r.MarkedAsPosted(v.Link); err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}

		err = sns.PostTwitter(v, red, mode)

		if mode == bot.NicopetterNewRedirectArticle || mode == bot.NicopetterNewArticle || mode == bot.NicopetterModifyRedirectArticle {

			switch {
			// RSSがソートされていない関係上，すべてのRSSを見るようにする配慮
			case err != nil && err.Error() == "twitter: 187 Status is a duplicate.":
				log.Print(err)
				continue
			case err != nil:
				return err
			}
		}
		if err != nil {
			log.Fatal(err)
			if err = r.SetLastUpdateTime(lastPublish); err != nil {
				return err
			}
			return err
		}

		lastPublish = *v.PublishedParsed
	}

	return nil
}

func createTwitterAuth() mytwitter.Authorization {
	return mytwitter.Authorization{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	}
}

func extractRedirect(f *gofeed.Item) (*nicopedia.Redirect, error) {
	u, err := url.Parse(f.Link)
	if err != nil {
		return nil, err
	}

	title, err := FetchRedirectTitle(u)
	if err != nil {
		if err.Error() == "no redirect in response" {
			return &nicopedia.Redirect{Exits: false}, nil
		}
		return nil, err
	}

	return &nicopedia.Redirect{Exits: true, Title: *title}, nil
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
