package sns

import (
	auth "github.com/bootjp/go_twitter_bot_for_nicopedia/domain/twitter"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"time"
)

// Twitter base struct.
type Twitter struct {
	client *twitter.Client
}

func NewTwitter(auth auth.Authorization) (*Twitter, error) {
	config := oauth1.NewConfig(auth.ConsumerKey, auth.ConsumerSecret)
	token := oauth1.NewToken(auth.AccessToken, auth.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	httpClient.Timeout = 10 * time.Second
	client := twitter.NewClient(httpClient)

	return &Twitter{client: client}, nil
}

// Post is Item to Twitter post.
func (t *Twitter) Post(post string) error {
	_, _, err := t.client.Statuses.Update(post, nil)
	return err
}

func (m *Twitter) String() string {
	// twitter は互換性のため空文字とする
	return ""
}
