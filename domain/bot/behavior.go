package bot

// Behavior is bot business logic behavior switch types.
type Behavior struct {
	TweetFormat       string
	EnableHTTPRequest bool
}

var (
	// Gunyapetter is Nicopedia oekakiko tweet account. https://twitter.com/gunyapetter
	Gunyapetter = &Behavior{"%s%s に %s というお絵カキコが投稿されたよ。%s", false}

	// DulltterTmp is Nicopedia pikokakiko tweet account. https://twitter.com/dulltter_tmp
	DulltterTmp = &Behavior{"%s%s に %s というピコカキコが投稿されたよ。%s", false}

	// NicopetterNewArticle is Nicopedia new general article tweet account.
	NicopetterNewArticle = &Behavior{"%s の記事ができたよ。%s", true}

	// NicopetterRedirectArticle is Nicopedia general article is to redirect tweet account.
	NicopetterRedirectArticle = &Behavior{"%s から %s へのリダイレクトができたよ。 %s", true}
)
