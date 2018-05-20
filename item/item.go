package item

import (
	"time"

	"net/http"

	"log"

	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/bot"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/store"
	"github.com/mmcdole/gofeed"
)

// FilterDate time after item return.
func FilterDate(f []*gofeed.Item, t time.Time) []*gofeed.Item {
	var itm []*gofeed.Item
	for _, elem := range f {
		if elem.PublishedParsed.After(t) {
			itm = append(itm, elem)
		}
	}

	return itm
}

// FilterMarkedAsPost no redis mark as post return item.
func FilterMarkedAsPost(f []*gofeed.Item, r *store.Redis, mode *bot.Behavior) ([]*gofeed.Item, error) {
	var itm []*gofeed.Item
	for _, elem := range f {
		var ng bool
		var err error
		switch mode {
		case bot.NicopetterNewArticle:
			ng, err = r.URLPosted(elem.Link, -1)
		case bot.NicopetterModifyRedirectArticle:
			ng, err = r.URLPosted(elem.Link, 86400)
		case bot.NicopetterNewRedirectArticle:
			ng, err = r.URLPosted(elem.Link, 86400)
		}

		if err != nil {
			return nil, err
		}
		if !ng {
			itm = append(itm, elem)
		}
	}

	return itm, nil
}

// Fetch is got url to fetch and return rss.
func Fetch(URL string) ([]*gofeed.Item, error) {
	p := gofeed.Parser{Client: &http.Client{Timeout: time.Duration(10) * time.Second}}
	f, err := p.ParseURL(URL)
	if err != nil {
		log.Println(URL)
		return nil, err
	}

	return f.Items, nil
}
