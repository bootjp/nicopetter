package main

import (
	"testing"

	"bytes"
	"io/ioutil"

	"go_twitter_bot/item"

	"time"

	"fmt"

	"github.com/mmcdole/gofeed"
)

func TestFilterSuccess(t *testing.T) {
	f, _ := ioutil.ReadFile("../testdata/bootjp.me/feed.rss")

	// Get actual value
	fp := gofeed.NewParser()
	feed, _ := fp.Parse(bytes.NewReader(f))

	if len(item.FilterItemDate(feed.Items, time.Now())) != 0 {
		t.Fatalf("count miss match.")
	}

	feed, _ = fp.Parse(bytes.NewReader(f))

	tm, _ := time.Parse("2006-01-02 15:04:05", "2018-04-01 00:33:06")

	fmt.Printf("%v\n", tm)
	if len(item.FilterItemDate(feed.Items, tm)) != 0 {
		t.Fatalf("count miss match.")
	}

	tm, _ = time.Parse("2006-01-02 15:04:05", "2018-04-01 00:33:05")

	fmt.Printf("%v\n", tm)
	if len(item.FilterItemDate(feed.Items, tm)) != 1 {
		t.Fatalf("count miss match.")
	}
}
