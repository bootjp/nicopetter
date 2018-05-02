package bot

import (
	"strings"

	"github.com/pkg/errors"
)

// Behavior is bot business logic behavior switch types.
type Behavior struct {
	TweetFormat       string
	FeedURL           string
	EnableHTTPRequest bool
	StorePrefix       string
}

var (
	// Gunyapetter is Nicopedia oekakiko tweet account. https://twitter.com/gunyapetter
	Gunyapetter = &Behavior{
		"%s%s に %s というお絵カキコが投稿されたよ。%s",
		"https://dic.nicovideo.jp/feed/rss/n/oekaki",
		false,
		"gunyapetter:",
	}

	// DulltterTmp is Nicopedia pikokakiko tweet account. https://twitter.com/dulltter_tmp
	DulltterTmp = &Behavior{
		"%s%s に %s というピコカキコが投稿されたよ。%s",
		"https://dic.nicovideo.jp/feed/rss/n/mml",
		false,
		"dulltter:",
	}

	// NicopetterNewArticle is Nicopedia new general article tweet account.
	NicopetterNewArticle = &Behavior{
		"%s の記事ができたよ。%s",
		"https://dic.nicovideo.jp/feed/rss/n/a",
		true,
		"nicopetter_new:",
	}

	// NicopetterRedirectArticle is Nicopedia general article is to redirect tweet account.
	NicopetterRedirectArticle = &Behavior{
		"%s から %s へのリダイレクトができたよ。 %s",
		"https://dic.nicovideo.jp/feed/rss/u/a",
		true,
		"nicopetter_redirect:",
	}
)

// NewBehavior is cli string from Behavior pointers.
func NewBehavior(mode string) (*Behavior, error) {
	switch strings.ToLower(mode) {
	case "gunyapetter":
		return Gunyapetter, nil
	case "dulltter":
		return DulltterTmp, nil
	case "nicopetter_new":
		return NicopetterNewArticle, nil
	case "nicopetter_redirect":
		return NicopetterRedirectArticle, nil
	default:
		return nil, errors.New("mode is invalid string")
	}
}
