package trcache_rueidis

import (
	"context"
	"fmt"

	"github.com/rrgmc/trcache"
	"github.com/rrgmc/trcache/codec"
	"github.com/rueian/rueidis"
)

type Cache[K comparable, V any] struct {
	options rootOptionsImpl[K, V]
	redis   rueidis.Client
}

func New[K comparable, V any](redis rueidis.Client, options ...trcache.RootOption) (*Cache[K, V], error) {
	ret := &Cache[K, V]{
		redis: redis,
		options: rootOptionsImpl[K, V]{
			defaultDuration: 0, // 0 means default for go-redis
			redisGetFunc:    DefaultRedisGetFunc[K, V]{},
			redisSetFunc:    DefaultRedisSetFunc[K, V]{},
			redisDelFunc:    DefaultRedisDelFunc[K, V]{},
		},
	}
	optErr := trcache.ParseOptions(&ret.options, options)
	if optErr.Err() != nil {
		return nil, optErr.Err()
	}
	if ret.options.valueCodec == nil {
		ret.options.valueCodec = codec.NewGOBCodec[V]()
	}
	if ret.options.keyCodec == nil {
		ret.options.keyCodec = codec.NewStringKeyCodec[K]()
	}
	return ret, nil
}

func (c *Cache[K, V]) Handle() rueidis.Client {
	return c.redis
}

func (c *Cache[K, V]) Name() string {
	return c.options.name
}

func (c *Cache[K, V]) Get(ctx context.Context, key K, options ...trcache.GetOption) (V, error) {
	optns := getOptionsImpl[K, V]{
		clientSideDuration: c.options.defaultClientSideDuration,
		redisGetFunc:       c.options.redisGetFunc,
	}
	optErr := trcache.ParseOptions(&optns, c.options.callDefaultGetOptions, options)
	if optErr.Err() != nil {
		var empty V
		return empty, optErr.Err()
	}

	keyValue, err := c.parseKey(ctx, key)
	if err != nil {
		var empty V
		return empty, err
	}

	value, err := optns.redisGetFunc.Get(ctx, c, keyValue, optns.customParams, optns.clientSideDuration)
	if err != nil {
		var empty V
		return empty, err
	}

	dec, err := c.options.valueCodec.Decode(ctx, value)
	if err != nil {
		var empty V
		return empty, trcache.CodecError{err}
	}

	if c.options.validator != nil {
		if err = c.options.validator.ValidateGet(ctx, dec); err != nil {
			var empty V
			return empty, err
		}
	}

	return dec, nil
}

func (c *Cache[K, V]) Set(ctx context.Context, key K, value V, options ...trcache.SetOption) error {
	optns := setOptionsImpl[K, V]{
		redisSetFunc: c.options.redisSetFunc,
		duration:     c.options.defaultDuration,
	}
	optErr := trcache.ParseOptions(&optns, c.options.callDefaultSetOptions, options)
	if optErr.Err() != nil {
		return optErr.Err()
	}

	enc, err := c.options.valueCodec.Encode(ctx, value)
	if err != nil {
		return trcache.CodecError{err}
	}

	keyValue, err := c.parseKey(ctx, key)
	if err != nil {
		return err
	}

	var strvalue string
	switch tv := enc.(type) {
	case string:
		strvalue = tv
	case []byte:
		strvalue = string(tv)
	default:
		return &trcache.InvalidValueTypeError{fmt.Sprintf("invalid type '%T' for redis value", enc)}
	}

	return optns.redisSetFunc.Set(ctx, c, keyValue, strvalue, c.options.defaultDuration, optns.customParams)
}

func (c *Cache[K, V]) Delete(ctx context.Context, key K, options ...trcache.DeleteOption) error {
	optns := deleteOptionsImpl[K, V]{
		redisDelFunc: c.options.redisDelFunc,
	}
	optErr := trcache.ParseOptions(&optns, c.options.callDefaultDeleteOptions, options)
	if optErr.Err() != nil {
		return optErr.Err()
	}

	keyValue, err := c.parseKey(ctx, key)
	if err != nil {
		return err
	}

	return optns.redisDelFunc.Delete(ctx, c, keyValue, optns.customParams)
}

func (c *Cache[K, V]) parseKey(ctx context.Context, key K) (string, error) {
	keyValue, err := c.options.keyCodec.Convert(ctx, key)
	if err != nil {
		return "", trcache.CodecError{err}
	}

	switch kv := keyValue.(type) {
	case string:
		return kv, nil
	case []byte:
		return string(kv), nil
	default:
		return "", trcache.CodecError{
			&trcache.InvalidValueTypeError{fmt.Sprintf("invalid type '%T' for redis key", keyValue)},
		}
	}
}
