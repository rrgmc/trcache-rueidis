package trcache_rueidis

import (
	"context"
	"time"

	"github.com/rrgmc/trcache"
	"github.com/rueian/rueidis"
)

type RedisGetFunc[K comparable, V any] interface {
	Get(ctx context.Context, c *Cache[K, V], keyValue string, customParams any, clientSideDuration time.Duration) (string, error)
}

type RedisSetFunc[K comparable, V any] interface {
	Set(ctx context.Context, c *Cache[K, V], keyValue string, valueValue string, expiration time.Duration, customParams any) error
}

type RedisDelFunc[K comparable, V any] interface {
	Delete(ctx context.Context, c *Cache[K, V], keyValue string, customParams any) error
}

// Interface funcs

type RedisGetFuncFunc[K comparable, V any] func(ctx context.Context, c *Cache[K, V], keyValue string, customParams any, clientSideDuration time.Duration) (string, error)

func (o RedisGetFuncFunc[K, V]) Get(ctx context.Context, c *Cache[K, V], keyValue string, customParams any, clientSideDuration time.Duration) (string, error) {
	return o(ctx, c, keyValue, customParams, clientSideDuration)
}

type RedisSetFuncFunc[K comparable, V any] func(ctx context.Context, c *Cache[K, V], keyValue string, valueValue string, expiration time.Duration, customParams any) error

func (o RedisSetFuncFunc[K, V]) Set(ctx context.Context, c *Cache[K, V], keyValue string, valueValue string, expiration time.Duration, customParams any) error {
	return o(ctx, c, keyValue, valueValue, expiration, customParams)
}

type RedisDelFuncFunc[K comparable, V any] func(ctx context.Context, c *Cache[K, V], keyValue string, customParams any) error

func (o RedisDelFuncFunc[K, V]) Delete(ctx context.Context, c *Cache[K, V], keyValue string, customParams any) error {
	return o(ctx, c, keyValue, customParams)
}

// Default

type DefaultRedisGetFunc[K comparable, V any] struct {
}

func (f DefaultRedisGetFunc[K, V]) Get(ctx context.Context, c *Cache[K, V], keyValue string, _ any, clientSideDuration time.Duration) (string, error) {
	cmd := c.Handle().B().Get().Key(keyValue).Cache()
	res := c.Handle().DoCache(ctx, cmd, clientSideDuration)
	value, err := res.ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "", trcache.ErrNotFound
		}
		return "", err
	}
	return value, nil
}

type DefaultRedisSetFunc[K comparable, V any] struct {
}

func (f DefaultRedisSetFunc[K, V]) Set(ctx context.Context, c *Cache[K, V], keyValue string, valueValue string,
	expiration time.Duration, _ any) error {
	cmd := c.Handle().B().Set().Key(keyValue).Value(valueValue).ExSeconds(expiration.Milliseconds() / 1000).Build()
	return c.Handle().Do(ctx, cmd).Error()
}

type DefaultRedisDelFunc[K comparable, V any] struct {
}

func (f DefaultRedisDelFunc[K, V]) Delete(ctx context.Context, c *Cache[K, V], keyValue string, _ any) error {
	return c.Handle().Do(ctx, c.Handle().B().Del().Key(keyValue).Build()).Error()
}
