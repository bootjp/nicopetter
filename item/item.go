package item

import (
	"time"

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

// FetchFeed is got url to fetch and return rss.
func Fetch(URL string) ([]*gofeed.Item, error) {
	fp := gofeed.NewParser()
	f, err := fp.ParseURL(URL)
	if err != nil {
		return nil, err
	}

	return f.Items, nil
}
