package main

import (
	"testing"

	"bytes"
	"io/ioutil"

	"time"

	"github.com/bootjp/go_twitter_bot/item"

	"github.com/mmcdole/gofeed"
)

func TestFilterSuccess(t *testing.T) {
	f, _ := ioutil.ReadFile("../testdata/bootjp.me/feed.xml")

	// Get actual value
	fp := gofeed.NewParser()
	feed, _ := fp.Parse(bytes.NewReader(f))

	if len(item.FilterDate(feed.Items, time.Now())) != 0 {
		t.Fatalf("count miss match.")
	}

	feed, _ = fp.Parse(bytes.NewReader(f))
	tm, _ := time.Parse("2006-01-02 15:04:05", "2018-04-01 00:33:06")

	if len(item.FilterDate(feed.Items, tm)) != 0 {
		t.Fatalf("item count miss match. expect 0 got %d.", len(item.FilterDate(feed.Items, tm)))
	}

	feed, _ = fp.Parse(bytes.NewReader(f))
	tm, _ = time.Parse("2006-01-02 15:04:05", "2018-04-01 00:33:05")
	if len(item.FilterDate(feed.Items, tm)) != 1 {
		t.Fatalf("item count miss match. expect 0 got %d.", len(item.FilterDate(feed.Items, tm)))
	}

	feed, _ = fp.Parse(bytes.NewReader(f))
	tm, _ = time.Parse("2006-01-02 15:04:05", "2018-04-01 00:33:04")
	if len(item.FilterDate(feed.Items, tm)) != 1 {
		t.Fatalf("item count miss match. expect 0 got %d.", len(item.FilterDate(feed.Items, tm)))
	}
}
