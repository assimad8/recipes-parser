package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type JobGuard struct {
	client *goredis.Client
	ttl    time.Duration
}

func NewJobGuard(client *goredis.Client, ttl time.Duration) *JobGuard {
	return &JobGuard{
		client: client,
		ttl:    ttl,
	}
}

func (g *JobGuard) Acquire(ctx context.Context, key string) (string, bool, error) {
	token := uuid.NewString()

	ok, err := g.client.SetNX(
		ctx,
		key,
		token,
		g.ttl,
	).Result()

	if err != nil {
		return "", false, err
	}

	return token, ok, nil
}

var releaseScript = goredis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`)

func (g *JobGuard) Release(ctx context.Context, key, token string) error {
	return releaseScript.Run(
		ctx,
		g.client,
		[]string{key},
		token,
	).Err()
}
