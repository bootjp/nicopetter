package bot

// Behavior is bot business logic behavior switch types.
type Behavior struct {
	TweetFormat       string
	EnableHTTPRequest bool
}

var (
	// Gunyapetter is Nicopedia oekakiko tweet account. https://twitter.com/gunyapetter
	Gunyapetter = &Behavior{"dic.nicovideo.jp/b/a/", false}

	// DulltterTmp is Nicopedia pikokakiko tweet account. https://twitter.com/dulltter_tmp
	DulltterTmp = &Behavior{"dic.nicovideo.jp/b/u/", true}

	// Nicopetter is Nicopedia general article tweet account.
	Nicopetter = &Behavior{"", true}
)
