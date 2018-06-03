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
	"gopkg.in/urfave/cli.v2"
)

// SNS base struct.
type SNS struct {
	mytwitter.Authorization
}

// Twitter is Item to SNS post.
func (t *SNS) Twitter(i *gofeed.Item, meta *nicopedia.MetaData, mode *bot.Behavior) error {
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
func FetchArticleMeta(u fmt.Stringer) (nicopedia.MetaData, error) {

	const TitleSuffix = `location.replace('http://dic.nicovideo.jp/a/`
	c := http.Client{Timeout: time.Duration(10) * time.Second}
	res, err := c.Get(u.String())
	if err != nil {
		return nicopedia.MetaData{}, err
	}
	defer func() {
		err = res.Body.Close()
		if err == nil {
			return
		}
		log.Fatal("failed to close response : ", err)
	}()

	switch res.Status[:1] {
	case "4", "5":
		return nicopedia.MetaData{}, fmt.Errorf("got %s status code", res.Status)
	case "3":
		log.Println("warn got 30x statsu code")
	}

	row, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nicopedia.MetaData{}, err
	}
	body := string(row)
	meta := nicopedia.MetaData{}

	// if len(doc.Nodes) == 0 {
	// 	return nicopedia.MetaData{}, errors.New("got empty response")
	// }
	//
	// var head string
	// doc.Find("head").Each(func(i int, s *goquery.Selection) {
	// 	head = s.Text()
	// })
	//

	// doc.Find("#article").Each(func(i int, selection *goquery.Selection) {
	// 	var html = selection.Text()
	// 	const checkLen = len("初版作成日") + 2
	// 	const dateLen = len(`YY/MM/DD HH:MM`)
	// 	const newArticleTag = `<span style="color:red;">`
	// 	const newArticleTagOne = `<`
	//
	// 	cin := strings.LastIndex(html, "初版作成日")
	// 	if cin == -1 {
	// 		return
	// 	}
	//
	// 	start := cin + checkLen
	// 	if html[start:start+1] == newArticleTagOne {
	// 		start += len(newArticleTag)
	// 	}
	// 	end := start + dateLen
	// 	meta.CreateAt, err = time.Parse("06/01/02 15:04", html[start:end])
	// 	if err != nil {
	// 		log.Println(u.String(), start, end, html[start:end])
	// 		log.Fatal(err)
	// 	}
	// })

	redirect := strings.Contains(body, `location.replace`)
	if !redirect {
		meta.IsRedirect = false
		return meta, nil
	}

	f := strings.Index(body, TitleSuffix)
	if f == -1 {
		meta.IsRedirect = false
		return meta, nil
	}

	body = body[f+len(TitleSuffix):]
	i := strings.Index(body, `'`)
	body = body[:i]

	meta.IsRedirect = true
	meta.FromTitle, err = url.QueryUnescape(body)
	return meta, err
}

func routine(mode *bot.Behavior, r *store.Redis) error {
	itm, err := item.Fetch(mode.FeedURL)
	if err != nil {
		return err
	}

	var lastPublish time.Time
	lastPublish, err = r.GetLastUpdateTime()
	if err != nil {
		return err
	}

	switch mode {
	case bot.Gunyapetter, bot.DulltterTmp:
		itm = item.FilterDate(itm, lastPublish)
	case bot.NicopetterModifyRedirectArticle, bot.NicopetterNewArticle, bot.NicopetterNewRedirectArticle:
		itm, err = item.FilterMarkedAsPost(itm, r, mode)
		if err != nil {
			return err
		}
	}

	if len(itm) == 0 {
		return nil
	}

	// sort
	sort.Slice(itm, func(i, j int) bool {
		return itm[i].PublishedParsed.Before(*itm[j].PublishedParsed)
	})

	sns := SNS{createTwitterAuth()}

	for _, v := range itm {
		var skip bool
		var meta *nicopedia.MetaData
		skip, meta, err = checkRedirectConditionIsSkip(mode, v)
		if err != nil {
			return err
		}
		if skip {
			continue
		}

		if err = markAs(mode, r, v); err != nil {
			return err
		}

		if err = sns.Twitter(v, meta, mode); err != nil {
			return fmt.Errorf("original error %v, rollback error %v", err, r.SetLastUpdateTime(lastPublish))
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

func checkRedirectConditionIsSkip(mode *bot.Behavior, i *gofeed.Item) (bool, *nicopedia.MetaData, error) {
	if !mode.CheckRedirect {
		return false, nil, nil
	}
	meta, err := extractRedirect(i)
	if err != nil {
		return true, nil, err
	}

	switch mode.FollowRedirect {
	case false:
		// 新着モードでリダイレクトしているものは無視する
		if meta.IsRedirect {
			return true, &meta, nil
		}
	case true:
		// リダイレクトモードでリダイレクト先が見つからないものは無視する
		if !meta.IsRedirect {
			return true, &meta, nil
		}
	}

	return false, &meta, nil
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

		redI, err := strconv.Atoi(os.Getenv("REDIS_INDEX"))
		if err != nil {
			return err
		}
		r := store.NewRedisClient(os.Getenv("REDIS_HOST"), redI, mode.StorePrefix)
		defer func() {
			err = r.Close()
			if err == nil {
				return
			}
			log.Fatal("failed to close: redis : ", err)
		}()
		if err != nil {
			return err
		}
		return routine(mode, r)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
