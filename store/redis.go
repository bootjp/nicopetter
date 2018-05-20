package store

import (
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

// RedisI is a convenient interface of Redis operation for bot.
type RedisI interface {
	GetLastUpdateTime() (time.Time, error)
	SetLastUpdateTime(time.Time) error
	Close() error
	URLPosted(string, int) (bool, error)
	MarkedAsPosted(u string) error
}

// NewRedisClient is a new store connect instance  operation for bot.
func NewRedisClient(host string, index int, prefix string) *Redis {
	r := &Redis{}
	r.p = prefix
	r.c = redis.NewClient(&redis.Options{
		Addr: host + ":6379",
		DB:   index,
	})

	return r
}

// GetLastUpdateTime は最後にいつRSSの更新があったかを返す関数です.
func (c *Redis) GetLastUpdateTime() (time.Time, error) {
	res, err := c.c.Get(c.p + "lastDate").Result()
	if err != nil {
		return time.Now(), err
	}

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Now(), err
	}

	return time.ParseInLocation(dateFormat, res, loc)
}

// Close is Redis connection close.
func (c *Redis) Close() error {
	return c.c.Close()
}

// SetLastUpdateTime set rss last update time.
func (c *Redis) SetLastUpdateTime(t time.Time) error {
	res := c.c.Set(c.p+"lastDate", t.Format(dateFormat), time.Duration(-1))
	_, err := res.Result()

	return err
}

// URLPosted is check url tweeted.
func (c *Redis) URLPosted(u string, exp int) (bool, error) {
	res, err := c.c.Exists(c.p + u).Result()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}

// MarkedAsPosted is url is tweeted mark.
func (c *Redis) MarkedAsPosted(u string) error {
	res, err := c.c.Set(c.p+u, "", time.Duration(24)*time.Hour*7).Result()
	if err != nil {
		return err
	}

	if res == "OK" {
		return nil
	}

	return errors.New("redis set error MarkedAsPosted")
}
