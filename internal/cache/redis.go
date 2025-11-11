package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    rdb *redis.Client
}

func NewRedisCache(rdb *redis.Client) *RedisCache {
    return &RedisCache{rdb: rdb}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest any) (bool, error) {
    s, err := c.rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    if err := json.Unmarshal([]byte(s), dest); err != nil {
        return false, err
    }
    return true, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, v any, ttl time.Duration) error {
    b, err := json.Marshal(v)
    if err != nil {
        return err
    }
    return c.rdb.Set(ctx, key, b, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.rdb.Del(ctx, key).Err()
}