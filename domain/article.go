package domain

import (
	"net/url"
	"strings"
)

var (
	// General is general article type.
	General = &ArticleType{"dic.nicovideo.jp/b/a/", "【単】"}
	// User is user article type.
	User = &ArticleType{"dic.nicovideo.jp/b/u/", "【ユ】"}

	// Live is live article type.
	Live = &ArticleType{"dic.nicovideo.jp/b/l/", "【生】"}

	// Video is movie article type.
	Video = &ArticleType{"dic.nicovideo.jp/b/v/", "【動】"}

	// Market is ichiba article type.
	Market = &ArticleType{"dic.nicovideo.jp/b/i/", "【市】"}

	// Community is community article type.
	Community = &ArticleType{"dic.nicovideo.jp/b/c/", "【コ】"}

	// Other is undefined article type.
	Other = &ArticleType{"", "【？】"}
)

// ArticleType is Nicopedia of article type.
type ArticleType struct {
	URLPrefix             string
	PostArticleExpression string
}

// ParseArticleType from url to ArticleType.
func ParseArticleType(url *url.URL) *ArticleType {
	switch {
	case strings.HasPrefix(url.Host+url.Path, General.URLPrefix):
		return General
	case strings.HasPrefix(url.Host+url.Path, User.URLPrefix):
		return User
	case strings.HasPrefix(url.Host+url.Path, Video.URLPrefix):
		return Video
	case strings.HasPrefix(url.Host+url.Path, Live.URLPrefix):
		return Live
	case strings.HasPrefix(url.Host+url.Path, Market.URLPrefix):
		return Market
	case strings.HasPrefix(url.Host+url.Path, Community.URLPrefix):
		return Community
	default:
		return Other
	}
}
