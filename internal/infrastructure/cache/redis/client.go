package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	db  *redis.Client
	ttl time.Duration
}

func NewClient(addr, password string, db int, ttl time.Duration) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return &Client{
		db:  rdb,
		ttl: ttl,
	}
}

func (c *Client) context()context.Context {
	return context.Background()
}
func (c *Client) Raw() *redis.Client {
	return c.db
}
