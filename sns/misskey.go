package sns

import (
	"github.com/sirupsen/logrus"
	"github.com/yitsushi/go-misskey"
	"github.com/yitsushi/go-misskey/core"
	"github.com/yitsushi/go-misskey/models"
	"github.com/yitsushi/go-misskey/services/notes"
)

type Misskey struct {
	client misskey.Client
}

func NewMisskey(token string) (*Misskey, error) {
	client, err := misskey.NewClientWithOptions(misskey.WithSimpleConfig("https://misskey.bootjp.me", token))
	if err != nil {
		return nil, err
	}
	client.LogLevel(logrus.ErrorLevel)

	return &Misskey{client: *client}, nil
}

func (m *Misskey) Post(post string) error {
	_, err := m.client.Notes().Create(notes.CreateRequest{
		Text:              core.NewString(post),
		Visibility:        models.VisibilityHome,
		NoExtractEmojis:   true,
		NoExtractHashtags: true,
		NoExtractMentions: true,
	})

	return err
}

func (m *Misskey) String() string {
	return "misskey"
}
