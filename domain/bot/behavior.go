package bot

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/nicopedia"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
)

// Behavior is bot business logic behavior switch types.
type Behavior struct {
	TweetFormat         string
	FeedURL             string
	EnableRedirectTitle bool
	FollowRedirect      bool
	StorePrefix         string
}

var (
	// Gunyapetter is Nicopedia oekakiko tweet account. https://twitter.com/gunyapetter
	Gunyapetter = &Behavior{
		TweetFormat: "%s%s に %s というお絵カキコが投稿されたよ。%s",
		FeedURL:     "https://dic.nicovideo.jp/feed/rss/n/oekaki",
		StorePrefix: "gunyapetter:",
	}

	// DulltterTmp is Nicopedia pikokakiko tweet account. https://twitter.com/dulltter_tmp
	DulltterTmp = &Behavior{
		TweetFormat: "%s%s に %s というピコカキコが投稿されたよ。%s",
		FeedURL:     "https://dic.nicovideo.jp/feed/rss/n/mml",
		StorePrefix: "dulltter:",
	}

	// NicopetterNewArticle is Nicopedia new general article tweet account.
	NicopetterNewArticle = &Behavior{
		TweetFormat:         "%s の記事ができたよ。%s",
		FeedURL:             "https://dic.nicovideo.jp/feed/rss/n/a",
		EnableRedirectTitle: true,
		StorePrefix:         "nicopetter_new:",
	}

	// NicopetterNewRedirectArticle is Nicopedia general article is to redirect tweet account.
	NicopetterNewRedirectArticle = &Behavior{
		TweetFormat:         "%s から %s へのリダイレクトができたよ。 %s",
		FeedURL:             "https://dic.nicovideo.jp/feed/rss/n/a",
		EnableRedirectTitle: true,
		FollowRedirect:      true,
		StorePrefix:         "nicopetter_new_redirect:",
	}
	NicopetterModifyRedirectArticle = &Behavior{
		TweetFormat:         "%s から %s へのリダイレクトができたよ。 %s",
		FeedURL:             "https://dic.nicovideo.jp/feed/rss/u/a",
		EnableRedirectTitle: true,
		FollowRedirect:      true,
		StorePrefix:         "nicopetter_new_redirect:",
	}
	//NicopetterModifyMovieArticle = &Behavior{
	//	TweetFormat:         "%s の
	//	FeedURL:             "https://dic.nicovideo.jp/feed/rss/u/v",
	//	EnableRedirectTitle: true,
	//	FollowRedirect:      true,
	//	StorePrefix:         "nicopetter_movie_update:",
	//}
	NicopetterNewMovieArticle = &Behavior{
		TweetFormat:         "%s の動画記事ができたよ。%s",
		FeedURL:             "https://dic.nicovideo.jp/feed/rss/n/v",
		EnableRedirectTitle: true,
		FollowRedirect:      false,
		StorePrefix:         "nicopetter_movie_new:",
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
	case "nicopetter_new_redirect":
		return NicopetterNewRedirectArticle, nil
	case "nicopetter_modify_redirect":
		return NicopetterModifyRedirectArticle, nil
	case "nicopetter_new_movie":
		return NicopetterNewMovieArticle, nil
	default:
		return nil, errors.New("mode is invalid string")
	}
}

func FormatPost(mode *Behavior, meta nicopedia.MetaData, i *gofeed.Item) (string, error) {
	var u, err = url.Parse(i.Link)
	if err != nil {
		return "", err
	}
	ar := nicopedia.ParseArticleType(u)

	switch mode {
	case Gunyapetter:
		return fmt.Sprintf(mode.TweetFormat, i.Title, ar.PostArticleExpression, i.Description, i.Link), nil
	case DulltterTmp:
		return fmt.Sprintf(mode.TweetFormat, i.Title, ar.PostArticleExpression, i.Description, i.Link), nil
	case NicopetterNewArticle:
		return fmt.Sprintf(mode.TweetFormat, i.Title, i.Link), nil
	case NicopetterNewRedirectArticle:
		return fmt.Sprintf(mode.TweetFormat, i.Title, meta.FromTitle, i.Link), nil
	case NicopetterModifyRedirectArticle:
		return fmt.Sprintf(mode.TweetFormat, i.Title, meta.FromTitle, i.Link), nil
	case NicopetterNewMovieArticle:
		return fmt.Sprintf(mode.TweetFormat, i.Title, i.Link), nil
	default:
		return "", errors.New("mode is invalid string")
	}
}
