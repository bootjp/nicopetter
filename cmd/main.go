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

	"github.com/PuerkitoBio/goquery"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/bot"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/nicopedia"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/item"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/sns"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/store"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"strings"
)

var ErrServer = errors.New("ignore")

// FetchArticleMeta is Nicopedia user redirect setting article redirect page title.
func FetchArticleMeta(u *url.URL) (nicopedia.MetaData, error) {
	const TitleSuffix = `location.replace('https://dic.nicovideo.jp/a/`
	c := http.Client{Timeout: 15 * time.Second}
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
	case "4":
		return nicopedia.MetaData{}, fmt.Errorf("got %s status code", res.Status)
	case "5":
		log.Println("got 5xx status code ignore")
		return nicopedia.MetaData{}, ErrServer
	case "3":
		log.Println("warn got 30x status code")
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

func run(mode *bot.Behavior) error {
	f, err := item.Fetch(mode.FeedURL)
	if err != nil {
		return err
	}

	i, err := strconv.Atoi(os.Getenv("REDIS_INDEX"))
	if err != nil {
		return err
	}
	r := store.NewRedisClient(os.Getenv("REDIS_HOST"), i, mode.StorePrefix, os.Getenv("REDIS_PASSWORD"))
	defer func() {
		_ = r.Close()
	}()

	if len(f) == 0 {
		return nil
	}

	// sort
	sort.Slice(f, func(i, j int) bool {
		return f[i].PublishedParsed.Before(*f[j].PublishedParsed)
	})

	m, err := sns.NewMisskey(os.Getenv("MISSKEY_TOKEN"))
	if err != nil {
		return err
	}

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

		post, err := bot.FormatPost(mode, meta, v)
		if err != nil {
			return err
		}

		already, err := r.URLPosted(v.Link, m)
		if err != nil {
			return err
		}

		if already {
			continue
		}

		err = m.Post(post)

		if err != nil && err.Error() == "json: cannot unmarshal object into Go struct field User.createdNote.user.emojis of type []models.Emoji" {
			err = nil
		}

		if err != nil {
			log.Println(err)
			return err
		}

		if err = r.MarkedAsPosted(v.Link, m); err != nil {
			return err
		}
	}

	return nil
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
		return run(mode)
	}
	if err := app.Run(os.Args); err != nil {
		// 5xx エラーはこっちでどうにもできないので無視する
		if err == ErrServer {
			return
		}
		log.Fatal(err)
	}
}
