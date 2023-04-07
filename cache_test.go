package trcache_rueidis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/RangelReale/trcache"
	"github.com/RangelReale/trcache/codec"
	"github.com/RangelReale/trcache/mocks"
	"github.com/golang/mock/gomock"
	"github.com/rueian/rueidis/mock"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	mockRedis.EXPECT().
		Do(gomock.Any(), mock.Match("SET", "a", "12", "EX", "60")).
		Return(mock.Result(mock.RedisString("")))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "a"), gomock.Any()).
		Return(mock.Result(mock.RedisString("12")))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "a"), gomock.Any()).
		Return(mock.Result(mock.RedisNil()))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "z"), gomock.Any()).
		Return(mock.Result(mock.RedisNil()))

	c, err := New[string, string](mockRedis,
		WithValueCodec[string, string](codec.NewForwardCodec[string]()),
		WithDefaultDuration[string, string](time.Minute),
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

func TestCacheValidator(t *testing.T) {
	ctx := context.Background()

	mockValidator := mocks.NewValidator[string](t)

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	mockRedis.EXPECT().
		Do(gomock.Any(), mock.Match("SET", "a", "12", "EX", "60")).
		Return(mock.Result(mock.RedisString("")))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "a"), gomock.Any()).
		Return(mock.Result(mock.RedisString("12")))

	mockValidator.EXPECT().
		ValidateGet(mock2.Anything, "12").
		Return(trcache.ErrNotFound).
		Once()

	c, err := New[string, string](mockRedis,
		WithValueCodec[string, string](codec.NewForwardCodec[string]()),
		WithValidator[string, string](mockValidator),
		WithDefaultDuration[string, string](time.Minute),
	)
	require.NoError(t, err)

	err = c.Set(ctx, "a", "12")
	require.NoError(t, err)

	_, err = c.Get(ctx, "a")
	require.ErrorIs(t, err, trcache.ErrNotFound)
}

func TestCacheCodecError(t *testing.T) {
	ctx := context.Background()

	mockCodec := mocks.NewCodec[string](t)

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "a"), gomock.Any()).
		Return(mock.Result(mock.RedisNil()))

	mockCodec.EXPECT().
		Marshal(mock2.Anything, "12").
		Return(nil, errors.New("my error"))

	c, err := New[string, string](mockRedis,
		WithValueCodec[string, string](mockCodec),
		WithDefaultDuration[string, string](time.Minute),
	)
	require.NoError(t, err)

	err = c.Set(ctx, "a", "12")
	require.ErrorAs(t, err, &trcache.CodecError{})

	_, err = c.Get(ctx, "a")
	require.ErrorIs(t, err, trcache.ErrNotFound)
}

func TestCacheJSONCodec(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	mockRedis.EXPECT().
		Do(gomock.Any(), mock.Match("SET", "a", `"12"`, "EX", "60")).
		Return(mock.Result(mock.RedisString("")))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "a"), gomock.Any()).
		Return(mock.Result(mock.RedisString(`"12"`)))

	c, err := New[string, string](mockRedis,
		WithValueCodec[string, string](codec.NewJSONCodec[string]()),
		WithDefaultDuration[string, string](time.Minute),
	)
	require.NoError(t, err)

	err = c.Set(ctx, "a", "12")
	require.NoError(t, err)

	v, err := c.Get(ctx, "a")
	require.NoError(t, err)
	require.Equal(t, "12", v)
}

func TestCacheJSONCodecInt(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	mockRedis.EXPECT().
		Do(gomock.Any(), mock.Match("SET", "a", "12", "EX", "60")).
		Return(mock.Result(mock.RedisString("")))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "a"), gomock.Any()).
		Return(mock.Result(mock.RedisString("12")))

	c, err := New[string, int](mockRedis,
		WithValueCodec[string, int](codec.NewJSONCodec[int]()),
		WithDefaultDuration[string, int](time.Minute),
	)
	require.NoError(t, err)

	err = c.Set(ctx, "a", 12)
	require.NoError(t, err)

	v, err := c.Get(ctx, "a")
	require.NoError(t, err)
	require.Equal(t, 12, v)
}

func TestCacheFuncCodecInt(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	mockRedis.EXPECT().
		Do(gomock.Any(), mock.Match("SET", "a", "12", "EX", "60")).
		Return(mock.Result(mock.RedisString("")))
	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "a"), gomock.Any()).
		Return(mock.Result(mock.RedisString("12")))

	c, err := New[string, int](mockRedis,
		WithValueCodec[string, int](codec.NewFuncCodec[int](
			func(ctx context.Context, data int) (any, error) {
				return fmt.Sprint(data), nil
			}, func(ctx context.Context, data any) (int, error) {
				return strconv.Atoi(fmt.Sprint(data))
			})),
		WithDefaultDuration[string, int](time.Minute),
	)
	require.NoError(t, err)

	err = c.Set(ctx, "a", 12)
	require.NoError(t, err)

	v, err := c.Get(ctx, "a")
	require.NoError(t, err)
	require.Equal(t, 12, v)
}

func TestCacheCodecInvalidInt(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	c, err := New[string, int](mockRedis,
		WithValueCodec[string, int](codec.NewForwardCodec[int]()),
		WithDefaultDuration[string, int](time.Minute),
	)
	require.NoError(t, err)

	err = c.Set(ctx, "a", 12)
	require.ErrorAs(t, err, new(*trcache.InvalidValueTypeError))
}

func TestCacheRefresh(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockRedis := mock.NewClient(ctrl)

	mockRedis.EXPECT().
		DoCache(gomock.Any(), mock.Match("GET", "a"), gomock.Any()).
		Return(mock.Result(mock.RedisNil()))
	mockRedis.EXPECT().
		Do(gomock.Any(), mock.Match("SET", "a", "abc123", "EX", "60")).
		Return(mock.Result(mock.RedisString("")))

	c, err := NewRefresh[string, string](mockRedis,
		WithValueCodec[string, string](codec.NewForwardCodec[string]()),
		WithDefaultDuration[string, string](time.Minute),
		trcache.WithDefaultRefreshFunc[string, string](func(ctx context.Context, key string, options trcache.RefreshFuncOptions) (string, error) {
			return fmt.Sprintf("abc%d", options.Data), nil
		}),
	)
	require.NoError(t, err)

	value, err := c.GetOrRefresh(ctx, "a", trcache.WithRefreshData[string, string](123))
	require.NoError(t, err)
	require.Equal(t, "abc123", value)
}
