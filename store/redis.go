package store

import (
	"time"

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
	Close() error
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
