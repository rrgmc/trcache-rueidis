package trcache_rueidis

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rrgmc/trcache"
	"github.com/rrgmc/trcache/codec"
	"github.com/rueian/rueidis"
	"github.com/rueian/rueidis/mock"
	"github.com/stretchr/testify/require"
)

func TestFuncGet(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	mockRedis.EXPECT().
		Do(gomock.Any(), mock.Match("HSET", "a", "f1", "12")).
		Return(mock.Result(mock.RedisString("")))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("HGET", "a", "f1"), gomock.Any()).
		Return(mock.Result(mock.RedisString("12")))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("HGET", "a", "f1"), gomock.Any()).
		Return(mock.Result(mock.RedisNil()))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("HGET", "z", "f1"), gomock.Any()).
		Return(mock.Result(mock.RedisNil()))

	c, err := New[string, string](mockRedis,
		WithValueCodec[string, string](codec.NewForwardCodec[string]()),
		WithDefaultDuration[string, string](time.Minute),
		trcache.WithCallDefaultGetOptions[string, string](
			WithGetRedisGetFuncFunc[string, string](func(ctx context.Context, c *Cache[string, string], keyValue string, customParams any, clientSideDuration time.Duration) (string, error) {
				cmd := c.Handle().B().Hget().Key(keyValue).Field("f1").Cache()
				res := c.Handle().DoCache(ctx, cmd, clientSideDuration)
				value, err := res.ToString()
				if err != nil {
					if rueidis.IsRedisNil(err) {
						return "", trcache.ErrNotFound
					}
					return "", err
				}
				return value, nil
			}),
		),
		trcache.WithCallDefaultSetOptions[string, string](
			WithSetRedisSetFuncFunc[string, string](func(ctx context.Context, c *Cache[string, string], keyValue string, valueValue string, expiration time.Duration, customParams any) error {
				cmd := c.Handle().B().Hset().Key(keyValue).FieldValue().FieldValue("f1", valueValue).Build()
				return c.Handle().Do(ctx, cmd).Error()
			}),
		),
		trcache.WithCallDefaultDeleteOptions[string, string](
			WithDeleteRedisDelFuncFunc[string, string](func(ctx context.Context, c *Cache[string, string], keyValue string, customParams any) error {
				return c.Handle().Do(ctx, c.Handle().B().Hdel().Key(keyValue).Field("f1").Build()).Error()
			}),
		),
	)
	require.NoError(t, err)

	err = c.Set(ctx, "a", "12")
	require.NoError(t, err)

	v, err := c.Get(ctx, "a")
	require.NoError(t, err)
	require.Equal(t, "12", v)

	v, err = c.Get(ctx, "a")
	require.ErrorIs(t, err, trcache.ErrNotFound)

	v, err = c.Get(ctx, "z")
	require.ErrorIs(t, err, trcache.ErrNotFound)
}
