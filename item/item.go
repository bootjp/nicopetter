package item

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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

// Fetch is got url to fetch and return rss.
func Fetch(URL string) ([]*gofeed.Item, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := http.Client{Transport: tr, Timeout: 15 * time.Second}

	res, err := c.Get(URL)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	p := gofeed.NewParser()
	f, err := p.Parse(res.Body)
	if err != nil {
		// try parse string with ignore invalid utf8
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		// skip no print char
		printOnly := func(r rune) rune {
			if !unicode.IsPrint(r) {
				return -1
			}
			if !utf8.ValidRune(r) {
				return -1
			}
			return r
		}
		f, err = p.ParseString(strings.Map(printOnly, string(body)))
		if err != nil {
			log.Println(URL)
			log.Println(string(body))
			log.Printf("%+v", err)
			return nil, err
		}
	}

	return f.Items, nil
}
