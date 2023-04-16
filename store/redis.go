package store

import (
	"github.com/bootjp/go_twitter_bot_for_nicopedia/sns"
	"time"

	"errors"

	"github.com/go-redis/redis"
)

const dateFormat = "2006-01-02 15:04:05"

// Redis is a convenient type of Redis operation for bot.
type Redis struct {
	c *redis.Client
	p string
}

// Store is a convenient interface of Redis operation for bot.
type Store interface {
	Close() error
	URLPosted(u string, s sns.SNS) (bool, error)
	MarkedAsPosted(u string, s sns.SNS) error
}

// NewRedisClient is a new store connect instance  operation for bot.
func NewRedisClient(host string, index int, prefix string, password string) *Redis {
	r := &Redis{}
	r.p = prefix
	r.c = redis.NewClient(&redis.Options{
		Addr:     host,
		DB:       index,
		Password: password,
	})

	return r
}

// Close is Redis connection close.
func (c *Redis) Close() error {
	return c.c.Close()
}

// URLPosted is check url tweeted.
func (c *Redis) URLPosted(u string, s sns.SNS) (bool, error) {
	res, err := c.c.Exists(c.p + s.String() + u).Result()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}

// MarkedAsPosted is url is tweeted mark.
func (c *Redis) MarkedAsPosted(u string, s sns.SNS) error {
	res, err := c.c.Set(c.p+s.String()+u, "", 24*time.Hour*7).Result()
	if err != nil {
		return err
	}

	if res == "OK" {
		return nil
	}

	return errors.New("redis set error MarkedAsPosted")
}
