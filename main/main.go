package main

import (
	"os"
	"sort"
	"strconv"

	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"io/ioutil"

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

// SNS base struct.
type SNS struct {
	mytwitter.Authorization
}

// Twitter is Item to SNS post.
func (t *SNS) Twitter(i *gofeed.Item, rd *nicopedia.Redirect, mode *bot.Behavior) error {
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
func FetchRedirectTitle(u *url.URL) (string, error) {

	const TitleSuffix = `location.replace('http://dic.nicovideo.jp/a/`
	c := http.Client{Timeout: time.Duration(10) * time.Second}
	res, err := c.Get(u.String())
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	row, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	body := string(row)

	redirect := strings.Contains(body, `location.replace`)
	if !redirect {
		return "", ErrNoRedirect
	}
	f := strings.Index(body, TitleSuffix)
	if f == -1 {
		return "", ErrNoRedirect
	}

	body = body[f+len(TitleSuffix):]
	i := strings.Index(body, `'`)
	body = body[:i]

	title, err := url.QueryUnescape(body)
	if err != nil {
		return "", err
	}

	return title, nil
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
	defer func() {
		cerr := r.Close()
		if cerr == nil {
			return
		}
		err = fmt.Errorf("failed to close: %v, the original error was %v", cerr, err)
	}()

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

	sns := SNS{createTwitterAuth()}

	lastPublish := t

	for _, v := range f {
		var skip bool
		var red *nicopedia.Redirect
		skip, red, err = checkRedirectConditionIsSkip(mode, v)
		if err != nil {
			return err
		}
		if skip {
			continue
		}

		if err = markAs(mode, r, v); err != nil {
			return err
		}

		err = sns.Twitter(v, red, mode)

		if err != nil {
			return fmt.Errorf("original error %v, last error %v", err, r.SetLastUpdateTime(lastPublish))
		}

		lastPublish = *v.PublishedParsed
	}

	return err
}

func createTwitterAuth() mytwitter.Authorization {
	return mytwitter.Authorization{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	}
}

func markAs(mode *bot.Behavior, r *store.Redis, i *gofeed.Item) error {
	var err error
	switch mode {
	case bot.Gunyapetter, bot.DulltterTmp:
		err = r.SetLastUpdateTime(*i.PublishedParsed)

	case bot.NicopetterNewRedirectArticle, bot.NicopetterNewArticle, bot.NicopetterModifyRedirectArticle:
		err = r.MarkedAsPosted(i.Link)
	}

	return err
}

func checkRedirectConditionIsSkip(mode *bot.Behavior, i *gofeed.Item) (bool, *nicopedia.Redirect, error) {
	if !mode.CheckRedirect {
		return false, nil, nil
	}
	red, err := extractRedirect(i)
	if err != nil {
		return true, nil, err
	}

	switch mode.FollowRedirect {
	case false:
		// 新着モードでリダイレクトしているものは無視する
		if red.Exits {
			return true, red, nil
		}
	case true:
		// リダイレクトモードでリダイレクト先が見つからないものは無視する
		if !red.Exits {
			return true, red, nil
		}
	}

	return false, red, nil
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

	return &nicopedia.Redirect{Exits: true, Title: title}, nil
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
