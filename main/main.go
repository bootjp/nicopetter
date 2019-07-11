package main

import (
	"crypto/tls"
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
	PostTwitter(i *gofeed.Item, rd *nicopedia.MetaData, mode *bot.Behavior) error
}

// PostTwitter is Item to Twitter post.
func (t *Twitter) PostTwitter(i *gofeed.Item, meta nicopedia.MetaData, mode *bot.Behavior) error {
	config := oauth1.NewConfig(t.Authorization.ConsumerKey, t.Authorization.ConsumerSecret)
	token := oauth1.NewToken(t.Authorization.AccessToken, t.Authorization.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	httpClient.Timeout = time.Duration(10 * time.Second)
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
		out = fmt.Sprintf(mode.TweetFormat, i.Title, meta.FromTitle, i.Link)
	case bot.NicopetterModifyRedirectArticle:
		out = fmt.Sprintf(mode.TweetFormat, i.Title, meta.FromTitle, i.Link)
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

// FetchArticleMeta is Nicopedia user redirect setting article redirect page title.
func FetchArticleMeta(u *url.URL) (nicopedia.MetaData, error) {
	const TitleSuffix = `location.replace('https://dic.nicovideo.jp/a/`
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := http.Client{Transport: tr, Timeout: time.Duration(15 * time.Second)}
	res, err := c.Get(u.String())
	if err != nil {
		log.Println(u.String())
		return nicopedia.MetaData{}, err
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	switch res.Status[:1] {
	case "4", "5":
		return nicopedia.MetaData{}, fmt.Errorf("got %s status code", res.Status)
	case "3":
		log.Println("warn got 30x statsu code")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nicopedia.MetaData{}, err
	}

	if len(doc.Nodes) == 0 {
		return nicopedia.MetaData{}, errors.New("got empty response")
	}

	var head string
	doc.Find("head").Each(func(i int, s *goquery.Selection) {
		head = s.Text()
	})

	meta := nicopedia.MetaData{}
	doc.Find("#article").Each(func(i int, selection *goquery.Selection) {
		var html = selection.Text()
		const checkLen = len("初版作成日") + 2
		const dateLen = len(`YY/MM/DD HH:MM`)
		const newArticleTag = `<span style="color:red;">`
		const newArticleTagOne = `<`

		cin := strings.LastIndex(html, "初版作成日")
		if cin == -1 {
			return
		}

		start := cin + checkLen
		if html[start:start+1] == newArticleTagOne {
			start += len(newArticleTag)
		}
		end := start + dateLen
		meta.CreateAt, err = time.Parse("06/01/02 15:04", html[start:end])
		if err != nil {
			log.Println(u.String(), start, end, html[start:end])
			log.Fatal(err)
		}
	})

	redirect := strings.Contains(head, `location.replace`)
	if !redirect {
		meta.IsRedirect = false
		return meta, nil
	}
	f := strings.Index(head, TitleSuffix)
	if f == -1 {
		meta.IsRedirect = false
		return meta, nil
	}

	head = head[f+len(TitleSuffix):]
	i := strings.Index(head, `'`)
	head = head[:i]

	meta.IsRedirect = true
	meta.FromTitle, err = url.QueryUnescape(head)

	if err != nil {
		return meta, err
	}

	return meta, nil
}

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
		_ = r.Close()
	}()

	t, err := r.GetLastUpdateTime()
	if err != nil {
		return err
	}

	switch mode {
	case bot.Gunyapetter, bot.DulltterTmp:
		f = item.FilterDate(f, t)
	case bot.NicopetterModifyRedirectArticle, bot.NicopetterNewArticle, bot.NicopetterNewRedirectArticle:
		f, err = item.FilterMarkedAsPost(f, r, mode)
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
		meta := nicopedia.MetaData{IsRedirect: false}
		switch mode {
		case bot.NicopetterNewArticle:
			meta, err = extractRedirect(v)
			if err != nil {
				return err
			}
			// 新着モードでリダイレクトしているものは無視する
			if meta.IsRedirect {
				continue
			}

			if v.PublishedParsed.Format("2006-01-02 15:04") != meta.CreateAt.Format("2006-01-02 15:04") {
				continue
			}

		case bot.NicopetterModifyRedirectArticle, bot.NicopetterNewRedirectArticle:
			meta, err = extractRedirect(v)
			if err != nil {
				return err
			}
			// リダイレクトモードでリダイレクト先が見つからないものは無視する
			if !meta.IsRedirect {
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

		err = sns.PostTwitter(v, meta, mode)

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

func extractRedirect(f *gofeed.Item) (nicopedia.MetaData, error) {
	u, err := url.Parse(f.Link)
	if err != nil {
		return nicopedia.MetaData{}, err
	}

	return FetchArticleMeta(u)

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
	}
}
